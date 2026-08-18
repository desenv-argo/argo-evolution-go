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
	"mime"
	"net/url"
	"os"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	argo_model "github.com/evolution-foundation/evolution-go/pkg/argo/model"
	argo_repository "github.com/evolution-foundation/evolution-go/pkg/argo/repository"
)

var slugPattern = regexp.MustCompile(`^[a-z0-9][a-z0-9-]{1,98}[a-z0-9]$`)

var ErrInvalidApplicationCredential = errors.New("invalid application credential")
var ErrMessageMediaTooLarge = errors.New("message media exceeds capture limit")

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

type LifecycleBackfillInput struct {
	From    time.Time `json:"from" binding:"required"`
	To      time.Time `json:"to" binding:"required"`
	Limit   int       `json:"limit"`
	Execute bool      `json:"execute"`
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
	GatewayOperationsOverview(ctx context.Context, filters argo_model.AttemptFilters) (*argo_model.GatewayOperationsOverview, error)
	RecordHeartbeat(ctx context.Context, slug, credential string, input HeartbeatInput) (*argo_model.IntegrationHeartbeat, error)
	ListHeartbeats(ctx context.Context, filters argo_model.HeartbeatFilters) ([]argo_model.IntegrationHeartbeat, error)
	HealthSummary(ctx context.Context, filters argo_model.HeartbeatFilters) (*argo_model.HealthSummary, error)
	RecordReceipt(ctx context.Context, instanceID string, providerMessageIDs []string, state string, occurredAt time.Time) error
	ReconcilePendingAged(ctx context.Context) (int64, error)
	ListLifecycleEvents(ctx context.Context, filters argo_model.LifecycleFilters) ([]argo_model.MessageLifecycleEvent, error)
	LifecycleSummary(ctx context.Context, filters argo_model.LifecycleFilters) (*argo_model.LifecycleSummary, error)
	BackfillLifecycle(ctx context.Context, input LifecycleBackfillInput) (*argo_model.LifecycleBackfillReport, error)
	StoreMessageMedia(ctx context.Context, instanceID, providerMessageID, fileName, mimeType string, content []byte) error
	GetMessageMedia(ctx context.Context, instanceID, providerMessageID string) (*argo_model.MessageMedia, error)
	DeleteMessageMediaBefore(ctx context.Context, cutoff time.Time, limit int) (int64, error)
}

type service struct {
	repository    argo_repository.Repository
	now           func() time.Time
	pendingAge    time.Duration
	mediaMaxBytes int64
	runtime       argo_model.GatewayRuntime
}

func (s *service) StoreMessageMedia(ctx context.Context, instanceID, providerMessageID, fileName, mimeType string, content []byte) error {
	instanceID = strings.TrimSpace(instanceID)
	providerMessageID = strings.TrimSpace(providerMessageID)
	if instanceID == "" || providerMessageID == "" || len(content) == 0 {
		return errors.New("instance, message and media content are required")
	}
	if int64(len(content)) > s.mediaMaxBytes {
		return ErrMessageMediaTooLarge
	}
	mimeType = normalizeMediaType(mimeType)
	fileName = sanitizeMediaFileName(fileName, providerMessageID)
	return s.repository.SaveMessageMedia(ctx, &argo_model.MessageMedia{
		InstanceID: instanceID, ProviderMessageID: providerMessageID,
		FileName: fileName, MimeType: strings.TrimSpace(mimeType),
		SizeBytes: int64(len(content)), Content: append([]byte(nil), content...),
	})
}

func normalizeMediaType(value string) string {
	mediaType, _, err := mime.ParseMediaType(strings.TrimSpace(value))
	if err != nil || mediaType == "" {
		return "application/octet-stream"
	}
	return mediaType
}

func (s *service) GetMessageMedia(ctx context.Context, instanceID, providerMessageID string) (*argo_model.MessageMedia, error) {
	if strings.TrimSpace(instanceID) == "" || strings.TrimSpace(providerMessageID) == "" {
		return nil, errors.New("instanceId and messageId are required")
	}
	return s.repository.GetMessageMedia(ctx, strings.TrimSpace(instanceID), strings.TrimSpace(providerMessageID))
}

func (s *service) DeleteMessageMediaBefore(ctx context.Context, cutoff time.Time, limit int) (int64, error) {
	return s.repository.DeleteMessageMediaBefore(ctx, cutoff.UTC(), limit)
}

func sanitizeMediaFileName(value, fallback string) string {
	value = strings.TrimSpace(strings.ReplaceAll(strings.ReplaceAll(value, "\\", "_"), "/", "_"))
	value = cleanValue(value, 255)
	if value == "" {
		return cleanValue(fallback, 128)
	}
	return value
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

func (s *service) GatewayOperationsOverview(ctx context.Context, filters argo_model.AttemptFilters) (*argo_model.GatewayOperationsOverview, error) {
	attempts, err := s.repository.AttemptSummary(ctx, filters)
	if err != nil {
		return nil, err
	}
	applications, instances, err := s.repository.GatewayUsage(ctx, filters)
	if err != nil {
		return nil, err
	}
	lifecycle, err := s.LifecycleSummary(ctx, argo_model.LifecycleFilters{From: filters.From, To: filters.To, ApplicationSlug: filters.ApplicationSlug, InstanceID: filters.InstanceID})
	if err != nil {
		return nil, err
	}
	now := s.now().UTC()
	runtime := s.runtime
	runtime.UptimeSec = int64(now.Sub(runtime.StartedAt).Seconds())
	if runtime.UptimeSec < 0 {
		runtime.UptimeSec = 0
	}
	state, signals := gatewayOperationalState(attempts, lifecycle)
	overview := &argo_model.GatewayOperationsOverview{
		From: filters.From, To: filters.To, GeneratedAt: now, State: state, Runtime: runtime,
		Attempts: *attempts, Lifecycle: *lifecycle, Applications: applications, Instances: instances,
		ErrorCategories: lifecycle.Failures, LegacyUnknown: attempts.UnverifiedIdentity, Signals: signals,
	}
	if attempts.Total > 0 {
		overview.LegacyPercentage = float64(attempts.UnverifiedIdentity) * 100 / float64(attempts.Total)
	}
	return overview, nil
}

func gatewayOperationalState(attempts *argo_model.AttemptSummary, lifecycle *argo_model.LifecycleSummary) (string, []string) {
	failureRate := float64(0)
	if attempts.Total > 0 {
		failureRate = float64(attempts.Failed) * 100 / float64(attempts.Total)
	}
	state := "healthy"
	if failureRate >= 10 || lifecycle.PendingAged > 0 {
		state = "unhealthy"
	} else if failureRate >= 3 || attempts.UnverifiedIdentity > 0 {
		state = "degraded"
	}
	signals := make([]string, 0, 3)
	if failureRate >= 3 {
		signals = append(signals, fmt.Sprintf("failure_rate:%.1f", failureRate))
	}
	if lifecycle.PendingAged > 0 {
		signals = append(signals, fmt.Sprintf("pending_aged:%d", lifecycle.PendingAged))
	}
	if attempts.UnverifiedIdentity > 0 {
		signals = append(signals, fmt.Sprintf("legacy_unknown:%d", attempts.UnverifiedIdentity))
	}
	return state, signals
}

func (s *service) RecordReceipt(ctx context.Context, instanceID string, providerMessageIDs []string, state string, occurredAt time.Time) error {
	return s.repository.RecordReceipt(ctx, instanceID, providerMessageIDs, state, occurredAt)
}

func (s *service) ReconcilePendingAged(ctx context.Context) (int64, error) {
	return s.repository.ReconcilePendingAged(ctx, s.now().UTC().Add(-s.pendingAge))
}

func (s *service) ListLifecycleEvents(ctx context.Context, filters argo_model.LifecycleFilters) ([]argo_model.MessageLifecycleEvent, error) {
	return s.repository.ListLifecycleEvents(ctx, filters)
}

func (s *service) LifecycleSummary(ctx context.Context, filters argo_model.LifecycleFilters) (*argo_model.LifecycleSummary, error) {
	events, err := s.repository.LifecycleEvents(ctx, filters)
	if err != nil {
		return nil, err
	}
	summary := summarizeLifecycle(events, filters.From, filters.To)
	summary.PendingAgeMinutes = int64(s.pendingAge.Minutes())
	return summary, nil
}

func (s *service) BackfillLifecycle(ctx context.Context, input LifecycleBackfillInput) (*argo_model.LifecycleBackfillReport, error) {
	options, err := lifecycleBackfillOptions(input)
	if err != nil {
		return nil, err
	}
	return s.repository.BackfillLifecycle(ctx, options)
}

func lifecycleBackfillOptions(input LifecycleBackfillInput) (argo_model.LifecycleBackfillOptions, error) {
	from, to := input.From.UTC(), input.To.UTC()
	if from.IsZero() || to.IsZero() || !from.Before(to) {
		return argo_model.LifecycleBackfillOptions{}, errors.New("from must be before to")
	}
	if to.Sub(from) > 31*24*time.Hour {
		return argo_model.LifecycleBackfillOptions{}, errors.New("backfill period cannot exceed 31 days")
	}
	limit := input.Limit
	if limit == 0 {
		limit = 1000
	}
	if limit < 1 || limit > 5000 {
		return argo_model.LifecycleBackfillOptions{}, errors.New("limit must be between 1 and 5000")
	}
	return argo_model.LifecycleBackfillOptions{
		From: from, To: to, Limit: limit, Execute: input.Execute,
	}, nil
}

type lifecycleTimes struct {
	received  *time.Time
	sent      *time.Time
	delivered *time.Time
	read      *time.Time
}

func summarizeLifecycle(events []argo_model.MessageLifecycleEvent, from, to time.Time) *argo_model.LifecycleSummary {
	summary := &argo_model.LifecycleSummary{From: from, To: to, Failures: []argo_model.LifecycleFailure{}}
	seen := map[string]map[string]bool{}
	times := map[string]*lifecycleTimes{}
	failures := map[string]int64{}
	for _, event := range events {
		key := lifecycleMessageKey(event)
		if seen[event.State] == nil {
			seen[event.State] = map[string]bool{}
		}
		seen[event.State][key] = true
		item := times[key]
		if item == nil {
			item = &lifecycleTimes{}
			times[key] = item
		}
		at := event.OccurredAt
		switch event.State {
		case "received":
			item.received = earliest(item.received, at)
		case "sent":
			item.sent = earliest(item.sent, at)
		case "delivered":
			item.delivered = earliest(item.delivered, at)
		case "read":
			item.read = earliest(item.read, at)
		case "failed":
			failureKey := event.FailureCategory + "\x00" + event.FailureCode
			failures[failureKey]++
		}
	}
	summary.Received = int64(len(seen["received"]))
	summary.Validated = int64(len(seen["validated"]))
	summary.Accepted = int64(len(seen["accepted"]))
	summary.Sent = int64(len(seen["sent"]))
	summary.Delivered = int64(len(seen["delivered"]))
	summary.Read = int64(len(seen["read"]))
	summary.Failed = int64(len(seen["failed"]))
	summary.PendingAged = int64(len(seen["pending_aged"]))
	summary.AcceptanceRate = rate(summary.Accepted, summary.Received)
	summary.SendRate = rate(summary.Sent, summary.Accepted)
	summary.DeliveryRate = rate(summary.Delivered, summary.Sent)
	summary.ReadRate = rate(summary.Read, summary.Delivered)

	var sendDurations, deliveryDurations, readDurations []float64
	for _, item := range times {
		if item.received != nil && item.sent != nil {
			sendDurations = appendNonNegativeDuration(sendDurations, *item.received, *item.sent)
		}
		if item.sent != nil && item.delivered != nil {
			deliveryDurations = appendNonNegativeDuration(deliveryDurations, *item.sent, *item.delivered)
		}
		if item.delivered != nil && item.read != nil {
			readDurations = appendNonNegativeDuration(readDurations, *item.delivered, *item.read)
		}
	}
	summary.SendLatency = latencyPercentiles(sendDurations)
	summary.DeliveryLatency = latencyPercentiles(deliveryDurations)
	summary.ReadLatency = latencyPercentiles(readDurations)
	for key, count := range failures {
		parts := strings.SplitN(key, "\x00", 2)
		summary.Failures = append(summary.Failures, argo_model.LifecycleFailure{Category: parts[0], Code: parts[1], Count: count})
	}
	sort.Slice(summary.Failures, func(i, j int) bool { return summary.Failures[i].Count > summary.Failures[j].Count })
	return summary
}

func lifecycleMessageKey(event argo_model.MessageLifecycleEvent) string {
	if event.AttemptID != nil && *event.AttemptID != "" {
		return "attempt:" + *event.AttemptID
	}
	instanceID := ""
	if event.InstanceID != nil {
		instanceID = *event.InstanceID
	}
	return "provider:" + instanceID + ":" + event.ProviderMessageID
}

func earliest(current *time.Time, candidate time.Time) *time.Time {
	if current == nil || candidate.Before(*current) {
		value := candidate
		return &value
	}
	return current
}

func rate(numerator, denominator int64) float64 {
	if denominator == 0 {
		return 0
	}
	return float64(numerator) * 100 / float64(denominator)
}

func appendNonNegativeDuration(values []float64, start, end time.Time) []float64 {
	if end.Before(start) {
		return values
	}
	return append(values, float64(end.Sub(start).Milliseconds()))
}

func latencyPercentiles(values []float64) argo_model.LifecycleLatency {
	if len(values) == 0 {
		return argo_model.LifecycleLatency{}
	}
	sort.Float64s(values)
	return argo_model.LifecycleLatency{
		P50: percentile(values, .50),
		P90: percentile(values, .90),
		P95: percentile(values, .95),
		P99: percentile(values, .99),
	}
}

func percentile(values []float64, quantile float64) float64 {
	position := quantile * float64(len(values)-1)
	lower := int(position)
	upper := lower + 1
	if upper >= len(values) {
		return values[lower]
	}
	weight := position - float64(lower)
	return values[lower]*(1-weight) + values[upper]*weight
}

func pendingAgeFromEnvironment() time.Duration {
	minutes, err := strconv.Atoi(strings.TrimSpace(os.Getenv("ARGO_MESSAGE_PENDING_AGE_MINUTES")))
	if err != nil || minutes < 1 || minutes > 10080 {
		minutes = 15
	}
	return time.Duration(minutes) * time.Minute
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

func NewService(repository argo_repository.Repository, runtime ...argo_model.GatewayRuntime) Service {
	runtimeInfo := argo_model.GatewayRuntime{Version: "0.0.0", StartedAt: time.Now().UTC()}
	if len(runtime) > 0 {
		runtimeInfo = runtime[0]
		if runtimeInfo.StartedAt.IsZero() {
			runtimeInfo.StartedAt = time.Now().UTC()
		}
	}
	return &service{
		repository:    repository,
		now:           time.Now,
		pendingAge:    pendingAgeFromEnvironment(),
		mediaMaxBytes: messageMediaMaxBytesFromEnvironment(),
		runtime:       runtimeInfo,
	}
}

func messageMediaMaxBytesFromEnvironment() int64 {
	const defaultLimit = int64(25 * 1024 * 1024)
	value := strings.TrimSpace(os.Getenv("ARGO_MESSAGE_MEDIA_MAX_BYTES"))
	if value == "" {
		return defaultLimit
	}
	parsed, err := strconv.ParseInt(value, 10, 64)
	if err != nil || parsed < 1024 || parsed > 100*1024*1024 {
		return defaultLimit
	}
	return parsed
}
