package argo_service

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"net/url"
	"regexp"
	"strings"
	"time"

	argo_model "github.com/evolution-foundation/evolution-go/pkg/argo/model"
	argo_repository "github.com/evolution-foundation/evolution-go/pkg/argo/repository"
)

var slugPattern = regexp.MustCompile(`^[a-z0-9][a-z0-9-]{1,98}[a-z0-9]$`)

var ErrInvalidApplicationCredential = errors.New("invalid application credential")

type ApplicationInput struct {
	Slug                     string `json:"slug"`
	Name                     string `json:"name"`
	Environment              string `json:"environment"`
	Owner                    string `json:"owner"`
	BaseURL                  string `json:"base_url"`
	HealthURL                string `json:"health_url"`
	Active                   *bool  `json:"active,omitempty"`
	ExpectedHeartbeatSeconds int    `json:"expected_heartbeat_seconds"`
}

type ApplicationCredential struct {
	Application *argo_model.Application `json:"application"`
	Credential  string                  `json:"credential"`
}

type Identity struct {
	ApplicationID   *string
	ApplicationSlug string
	Verified        bool
}

type HeartbeatInput struct {
	Status    string `json:"status"`
	LatencyMS int64  `json:"latency_ms"`
	Version   string `json:"version"`
	Component string `json:"component"`
	Message   string `json:"message"`
}

type Service interface {
	CreateApplication(ctx context.Context, input ApplicationInput) (*ApplicationCredential, error)
	UpdateApplication(ctx context.Context, id string, input ApplicationInput) (*argo_model.Application, error)
	RotateCredential(ctx context.Context, id string) (*ApplicationCredential, error)
	ListApplications(ctx context.Context) ([]argo_model.Application, error)
	Identify(ctx context.Context, slug, credential string) Identity
	RecordAttempt(ctx context.Context, attempt *argo_model.MessageAttempt) error
	ListAttempts(ctx context.Context, filters argo_model.AttemptFilters) ([]argo_model.MessageAttempt, error)
	AttemptSummary(ctx context.Context, filters argo_model.AttemptFilters) (*argo_model.AttemptSummary, error)
	RecordHeartbeat(ctx context.Context, slug, credential string, input HeartbeatInput) (*argo_model.IntegrationHeartbeat, error)
	ListHeartbeats(ctx context.Context, filters argo_model.HeartbeatFilters) ([]argo_model.IntegrationHeartbeat, error)
	HealthSummary(ctx context.Context, filters argo_model.HeartbeatFilters) (*argo_model.HealthSummary, error)
}

type service struct {
	repository argo_repository.Repository
	now        func() time.Time
}

func (s *service) CreateApplication(ctx context.Context, input ApplicationInput) (*ApplicationCredential, error) {
	if err := validateInput(input, true); err != nil {
		return nil, err
	}
	credential, hash, err := newCredential()
	if err != nil {
		return nil, err
	}
	active := true
	if input.Active != nil {
		active = *input.Active
	}
	application := &argo_model.Application{
		Slug:                     normalizeSlug(input.Slug),
		Name:                     strings.TrimSpace(input.Name),
		Environment:              normalizeEnvironment(input.Environment),
		Owner:                    strings.TrimSpace(input.Owner),
		BaseURL:                  strings.TrimSpace(input.BaseURL),
		HealthURL:                strings.TrimSpace(input.HealthURL),
		CredentialHash:           hash,
		Active:                   active,
		ExpectedHeartbeatSeconds: normalizeHeartbeat(input.ExpectedHeartbeatSeconds),
	}
	if err := s.repository.CreateApplication(ctx, application); err != nil {
		return nil, err
	}
	return &ApplicationCredential{Application: application, Credential: credential}, nil
}

func (s *service) UpdateApplication(ctx context.Context, id string, input ApplicationInput) (*argo_model.Application, error) {
	if err := validateInput(input, false); err != nil {
		return nil, err
	}
	application, err := s.repository.GetApplicationByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if application == nil {
		return nil, errors.New("application not found")
	}
	if input.Slug != "" {
		application.Slug = normalizeSlug(input.Slug)
	}
	if input.Name != "" {
		application.Name = strings.TrimSpace(input.Name)
	}
	if input.Environment != "" {
		application.Environment = normalizeEnvironment(input.Environment)
	}
	application.Owner = strings.TrimSpace(input.Owner)
	application.BaseURL = strings.TrimSpace(input.BaseURL)
	application.HealthURL = strings.TrimSpace(input.HealthURL)
	if input.Active != nil {
		application.Active = *input.Active
	}
	if input.ExpectedHeartbeatSeconds > 0 {
		application.ExpectedHeartbeatSeconds = normalizeHeartbeat(input.ExpectedHeartbeatSeconds)
	}
	if err := s.repository.SaveApplication(ctx, application); err != nil {
		return nil, err
	}
	return application, nil
}

func (s *service) RotateCredential(ctx context.Context, id string) (*ApplicationCredential, error) {
	application, err := s.repository.GetApplicationByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if application == nil {
		return nil, errors.New("application not found")
	}
	credential, hash, err := newCredential()
	if err != nil {
		return nil, err
	}
	application.CredentialHash = hash
	if err := s.repository.SaveApplication(ctx, application); err != nil {
		return nil, err
	}
	return &ApplicationCredential{Application: application, Credential: credential}, nil
}

func (s *service) ListApplications(ctx context.Context) ([]argo_model.Application, error) {
	applications, err := s.repository.ListApplications(ctx)
	if err != nil {
		return nil, err
	}
	now := s.now().UTC()
	for index := range applications {
		applications[index].HealthState, applications[index].HeartbeatAgeSeconds = applicationHealth(&applications[index], now)
	}
	return applications, nil
}

func (s *service) Identify(ctx context.Context, rawSlug, credential string) Identity {
	slug := normalizeSlug(rawSlug)
	if slug == "" {
		return Identity{ApplicationSlug: "legacy/unknown"}
	}
	identity := Identity{ApplicationSlug: slug}
	application, err := s.repository.GetApplicationBySlug(ctx, slug)
	if err != nil || application == nil || !application.Active {
		return identity
	}
	identity.ApplicationID = &application.ID
	if credential == "" {
		return identity
	}
	providedHash := hashCredential(credential)
	identity.Verified = subtle.ConstantTimeCompare([]byte(providedHash), []byte(application.CredentialHash)) == 1
	return identity
}

func (s *service) RecordAttempt(ctx context.Context, attempt *argo_model.MessageAttempt) error {
	return s.repository.RecordAttempt(ctx, attempt)
}

func (s *service) ListAttempts(ctx context.Context, filters argo_model.AttemptFilters) ([]argo_model.MessageAttempt, error) {
	return s.repository.ListAttempts(ctx, filters)
}

func (s *service) AttemptSummary(ctx context.Context, filters argo_model.AttemptFilters) (*argo_model.AttemptSummary, error) {
	return s.repository.AttemptSummary(ctx, filters)
}

func (s *service) RecordHeartbeat(ctx context.Context, slug, credential string, input HeartbeatInput) (*argo_model.IntegrationHeartbeat, error) {
	identity := s.Identify(ctx, slug, credential)
	if !identity.Verified || identity.ApplicationID == nil {
		return nil, ErrInvalidApplicationCredential
	}
	status := strings.ToLower(strings.TrimSpace(input.Status))
	if status == "" {
		status = "healthy"
	}
	if status != "healthy" && status != "degraded" && status != "unhealthy" {
		return nil, errors.New("status must be healthy, degraded or unhealthy")
	}
	if input.LatencyMS < 0 || input.LatencyMS > int64((10*time.Minute).Milliseconds()) {
		return nil, errors.New("latency_ms must be between 0 and 600000")
	}
	heartbeat := &argo_model.IntegrationHeartbeat{
		ApplicationID: *identity.ApplicationID,
		Status:        status,
		LatencyMS:     input.LatencyMS,
		Version:       cleanValue(input.Version, 80),
		Component:     cleanValue(input.Component, 100),
		Message:       cleanValue(input.Message, 500),
		ReceivedAt:    s.now().UTC(),
	}
	if err := s.repository.RecordHeartbeat(ctx, heartbeat); err != nil {
		return nil, err
	}
	return heartbeat, nil
}

func (s *service) ListHeartbeats(ctx context.Context, filters argo_model.HeartbeatFilters) ([]argo_model.IntegrationHeartbeat, error) {
	return s.repository.ListHeartbeats(ctx, filters)
}

func (s *service) HealthSummary(ctx context.Context, filters argo_model.HeartbeatFilters) (*argo_model.HealthSummary, error) {
	applications, err := s.ListApplications(ctx)
	if err != nil {
		return nil, err
	}
	summary := &argo_model.HealthSummary{From: filters.From, To: filters.To}
	for index := range applications {
		application := &applications[index]
		if !application.Active || (filters.ApplicationSlug != "" && application.Slug != filters.ApplicationSlug) {
			continue
		}
		summary.Applications++
		switch application.HealthState {
		case "healthy":
			summary.Healthy++
		case "degraded":
			summary.Degraded++
		case "offline":
			summary.Offline++
		default:
			summary.Unknown++
		}
	}
	summary.HeartbeatEvents, summary.UnhealthyEvents, summary.AverageLatencyMS, err = s.repository.HeartbeatMetrics(ctx, filters)
	if err != nil {
		return nil, err
	}
	return summary, nil
}

func applicationHealth(application *argo_model.Application, now time.Time) (string, int64) {
	if !application.Active {
		return "disabled", 0
	}
	if application.LastHeartbeatAt == nil {
		return "unknown", 0
	}
	age := now.Sub(application.LastHeartbeatAt.UTC())
	if age < 0 {
		age = 0
	}
	ageSeconds := int64(age.Seconds())
	expected := time.Duration(normalizeHeartbeat(application.ExpectedHeartbeatSeconds)) * time.Second
	if age > 2*expected {
		return "offline", ageSeconds
	}
	if application.LastHeartbeatStatus != "healthy" {
		return "degraded", ageSeconds
	}
	return "healthy", ageSeconds
}

func cleanValue(value string, limit int) string {
	value = strings.TrimSpace(strings.ReplaceAll(strings.ReplaceAll(value, "\r", " "), "\n", " "))
	if len(value) > limit {
		value = value[:limit]
	}
	return value
}

func validateInput(input ApplicationInput, creating bool) error {
	if creating || input.Slug != "" {
		slug := normalizeSlug(input.Slug)
		if !slugPattern.MatchString(slug) {
			return errors.New("slug must contain 3 to 100 lowercase letters, numbers or hyphens")
		}
	}
	if creating && strings.TrimSpace(input.Name) == "" {
		return errors.New("name is required")
	}
	for field, value := range map[string]string{"base_url": input.BaseURL, "health_url": input.HealthURL} {
		if value == "" {
			continue
		}
		parsed, err := url.ParseRequestURI(value)
		if err != nil || (parsed.Scheme != "http" && parsed.Scheme != "https") || parsed.Host == "" {
			return fmt.Errorf("%s must be a valid http or https URL", field)
		}
	}
	if input.ExpectedHeartbeatSeconds < 0 {
		return errors.New("expected_heartbeat_seconds cannot be negative")
	}
	return nil
}

func normalizeSlug(value string) string {
	return strings.ToLower(strings.TrimSpace(value))
}

func normalizeEnvironment(value string) string {
	value = strings.ToLower(strings.TrimSpace(value))
	if value == "" {
		return "production"
	}
	return value
}

func normalizeHeartbeat(value int) int {
	if value <= 0 {
		return 300
	}
	if value < 30 {
		return 30
	}
	if value > int((24 * time.Hour).Seconds()) {
		return int((24 * time.Hour).Seconds())
	}
	return value
}

func newCredential() (string, string, error) {
	raw := make([]byte, 32)
	if _, err := rand.Read(raw); err != nil {
		return "", "", err
	}
	credential := base64.RawURLEncoding.EncodeToString(raw)
	return credential, hashCredential(credential), nil
}

func hashCredential(value string) string {
	hash := sha256.Sum256([]byte(value))
	return hex.EncodeToString(hash[:])
}

func NewService(repository argo_repository.Repository) Service {
	return &service{repository: repository, now: time.Now}
}
