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

type Service interface {
	CreateApplication(ctx context.Context, input ApplicationInput) (*ApplicationCredential, error)
	UpdateApplication(ctx context.Context, id string, input ApplicationInput) (*argo_model.Application, error)
	RotateCredential(ctx context.Context, id string) (*ApplicationCredential, error)
	ListApplications(ctx context.Context) ([]argo_model.Application, error)
	Identify(ctx context.Context, slug, credential string) Identity
	RecordAttempt(ctx context.Context, attempt *argo_model.MessageAttempt) error
	ListAttempts(ctx context.Context, filters argo_model.AttemptFilters) ([]argo_model.MessageAttempt, error)
	AttemptSummary(ctx context.Context, filters argo_model.AttemptFilters) (*argo_model.AttemptSummary, error)
}

type service struct {
	repository argo_repository.Repository
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
	return s.repository.ListApplications(ctx)
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
	return &service{repository: repository}
}
