package argo_upstream

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	argo_model "github.com/evolution-foundation/evolution-go/pkg/argo/model"
)

const (
	DefaultRepository      = "evolution-foundation/evolution-go"
	DefaultBranch          = "main"
	DefaultBaselineSHA     = "9337afc47e10b86cc896a6f432240e40fee95dd1"
	DefaultBaselineVersion = "0.7.2"
)

type Config struct {
	Repository      string
	Branch          string
	BaselineSHA     string
	BaselineVersion string
	Token           string
	APIURL          string
}

type Checker struct {
	config Config
	client *http.Client
	now    func() time.Time
}

func ConfigFromEnvironment() Config {
	return Config{
		Repository:      envOr("ARGO_UPSTREAM_REPOSITORY", DefaultRepository),
		Branch:          envOr("ARGO_UPSTREAM_BRANCH", DefaultBranch),
		BaselineSHA:     envOr("ARGO_UPSTREAM_BASE_SHA", DefaultBaselineSHA),
		BaselineVersion: envOr("ARGO_UPSTREAM_BASE_VERSION", DefaultBaselineVersion),
		Token:           strings.TrimSpace(os.Getenv("ARGO_UPSTREAM_GITHUB_TOKEN")),
		APIURL:          envOr("ARGO_UPSTREAM_GITHUB_API_URL", "https://api.github.com"),
	}
}

func NewChecker(config Config, client *http.Client) *Checker {
	if client == nil {
		client = &http.Client{Timeout: 20 * time.Second}
	}
	return &Checker{config: config, client: client, now: time.Now}
}

func (c *Checker) Check(ctx context.Context) *argo_model.UpstreamSnapshot {
	snapshot := &argo_model.UpstreamSnapshot{Repository: c.config.Repository, Branch: c.config.Branch, BaselineSHA: c.config.BaselineSHA, BaselineVersion: c.config.BaselineVersion, Status: "unavailable", CheckedAt: c.now().UTC(), Changes: []argo_model.UpstreamChange{}}
	if !validRepository(c.config.Repository) || strings.TrimSpace(c.config.BaselineSHA) == "" {
		snapshot.Error = "invalid upstream repository or baseline"
		return snapshot
	}
	var latest struct {
		SHA     string `json:"sha"`
		HTMLURL string `json:"html_url"`
	}
	if err := c.get(ctx, fmt.Sprintf("/repos/%s/commits/%s", c.config.Repository, url.PathEscape(c.config.Branch)), &latest); err != nil {
		snapshot.Error = cleanError(err)
		return snapshot
	}
	snapshot.LatestSHA = latest.SHA
	var tags []struct {
		Name   string `json:"name"`
		Commit struct {
			SHA string `json:"sha"`
		} `json:"commit"`
	}
	if err := c.get(ctx, fmt.Sprintf("/repos/%s/tags?per_page=20", c.config.Repository), &tags); err == nil {
		for _, tag := range tags {
			if tag.Commit.SHA == latest.SHA {
				snapshot.LatestVersion = tag.Name
				break
			}
		}
		if snapshot.LatestVersion == "" && len(tags) > 0 {
			snapshot.LatestVersion = tags[0].Name
		}
	}
	if latest.SHA == c.config.BaselineSHA {
		snapshot.Status = "up_to_date"
		return snapshot
	}
	var comparison struct {
		Status  string `json:"status"`
		AheadBy int    `json:"ahead_by"`
		HTMLURL string `json:"html_url"`
		Commits []struct {
			SHA     string `json:"sha"`
			HTMLURL string `json:"html_url"`
			Commit  struct {
				Message string `json:"message"`
			} `json:"commit"`
		} `json:"commits"`
		Files []struct {
			Filename string `json:"filename"`
			Patch    string `json:"patch"`
		} `json:"files"`
	}
	path := fmt.Sprintf("/repos/%s/compare/%s...%s", c.config.Repository, url.PathEscape(c.config.BaselineSHA), url.PathEscape(latest.SHA))
	if err := c.get(ctx, path, &comparison); err != nil {
		snapshot.Error = cleanError(err)
		return snapshot
	}
	snapshot.BehindBy = comparison.AheadBy
	snapshot.CompareURL = comparison.HTMLURL
	snapshot.Status = "update_available"
	if comparison.Status == "diverged" {
		snapshot.Status = "diverged"
	}
	for _, file := range comparison.Files {
		if strings.EqualFold(file.Filename, "CHANGELOG.md") {
			for _, title := range addedChangelogEntries(file.Patch) {
				if len(snapshot.Changes) >= 20 {
					break
				}
				snapshot.Changes = append(snapshot.Changes, argo_model.UpstreamChange{Title: title, URL: comparison.HTMLURL, Category: classify(title)})
			}
		}
	}
	for index := len(comparison.Commits) - 1; index >= 0 && len(snapshot.Changes) < 20; index-- {
		commit := comparison.Commits[index]
		title := strings.Split(strings.TrimSpace(commit.Commit.Message), "\n")[0]
		snapshot.Changes = append(snapshot.Changes, argo_model.UpstreamChange{SHA: shortSHA(commit.SHA), Title: title, URL: commit.HTMLURL, Category: classify(title)})
	}
	return snapshot
}

func (c *Checker) get(ctx context.Context, path string, target interface{}) error {
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, strings.TrimRight(c.config.APIURL, "/")+path, nil)
	if err != nil {
		return err
	}
	request.Header.Set("Accept", "application/vnd.github+json")
	request.Header.Set("X-GitHub-Api-Version", "2022-11-28")
	request.Header.Set("User-Agent", "argo-evolution-go-upstream-monitor")
	if c.config.Token != "" {
		request.Header.Set("Authorization", "Bearer "+c.config.Token)
	}
	response, err := c.client.Do(request)
	if err != nil {
		return err
	}
	defer response.Body.Close()
	if response.StatusCode < 200 || response.StatusCode >= 300 {
		body, _ := io.ReadAll(io.LimitReader(response.Body, 512))
		return fmt.Errorf("github returned %d: %s", response.StatusCode, strings.TrimSpace(string(body)))
	}
	return json.NewDecoder(io.LimitReader(response.Body, 2*1024*1024)).Decode(target)
}

func classify(title string) string {
	value := strings.ToLower(title)
	switch {
	case strings.Contains(value, "security"), strings.Contains(value, "vulnerab"), strings.Contains(value, "cve"):
		return "security"
	case strings.HasPrefix(value, "fix"), strings.Contains(value, "bug"):
		return "fix"
	case strings.Contains(value, "breaking"), strings.Contains(value, "!:"):
		return "breaking"
	case strings.HasPrefix(value, "feat"):
		return "feature"
	case strings.Contains(value, "depend"), strings.HasPrefix(value, "chore"):
		return "maintenance"
	default:
		return "other"
	}
}

func addedChangelogEntries(patch string) []string {
	entries := make([]string, 0)
	for _, line := range strings.Split(patch, "\n") {
		if !strings.HasPrefix(line, "+") || strings.HasPrefix(line, "+++") {
			continue
		}
		value := strings.TrimSpace(strings.TrimPrefix(line, "+"))
		value = strings.TrimSpace(strings.TrimLeft(value, "-*"))
		if value == "" || strings.HasPrefix(value, "#") || strings.HasPrefix(value, "[") {
			continue
		}
		if len(value) > 240 {
			value = value[:240]
		}
		entries = append(entries, value)
		if len(entries) >= 20 {
			break
		}
	}
	return entries
}

func validRepository(value string) bool {
	parts := strings.Split(value, "/")
	return len(parts) == 2 && parts[0] != "" && parts[1] != "" && !strings.Contains(value, "..")
}
func envOr(name, fallback string) string {
	if value := strings.TrimSpace(os.Getenv(name)); value != "" {
		return value
	}
	return fallback
}
func shortSHA(value string) string {
	if len(value) > 7 {
		return value[:7]
	}
	return value
}
func cleanError(err error) string {
	if err == nil {
		return ""
	}
	value := strings.TrimSpace(err.Error())
	if value == "" {
		value = errors.New("upstream check failed").Error()
	}
	if len(value) > 500 {
		value = value[:500]
	}
	return value
}
