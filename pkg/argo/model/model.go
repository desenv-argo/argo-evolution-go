package argo_model

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Application identifies a system that consumes the Evolution API.
// Credentials are stored only as hashes and are never returned by this model.
type Application struct {
	ID                       string     `json:"id" gorm:"type:uuid;primaryKey"`
	Slug                     string     `json:"slug" gorm:"size:100;uniqueIndex;not null"`
	Name                     string     `json:"name" gorm:"size:160;not null"`
	Environment              string     `json:"environment" gorm:"size:32;not null;default:'production'"`
	Owner                    string     `json:"owner,omitempty" gorm:"size:160"`
	BaseURL                  string     `json:"base_url,omitempty" gorm:"type:text"`
	HealthURL                string     `json:"health_url,omitempty" gorm:"type:text"`
	CredentialHash           string     `json:"-" gorm:"size:64;not null"`
	Active                   bool       `json:"active" gorm:"not null;default:true"`
	ExpectedHeartbeatSeconds int        `json:"expected_heartbeat_seconds" gorm:"not null;default:300"`
	LastSeenAt               *time.Time `json:"last_seen_at,omitempty"`
	LastHeartbeatAt          *time.Time `json:"last_heartbeat_at,omitempty" gorm:"index"`
	LastHeartbeatStatus      string     `json:"last_heartbeat_status,omitempty" gorm:"size:24"`
	LastHeartbeatLatencyMS   int64      `json:"last_heartbeat_latency_ms,omitempty"`
	LastHeartbeatVersion     string     `json:"last_heartbeat_version,omitempty" gorm:"size:80"`
	HealthState              string     `json:"health_state" gorm:"-"`
	HeartbeatAgeSeconds      int64      `json:"heartbeat_age_seconds,omitempty" gorm:"-"`
	CreatedAt                time.Time  `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt                time.Time  `json:"updated_at" gorm:"autoUpdateTime"`
}

// IntegrationHeartbeat is an immutable health signal emitted by an
// authenticated application. Keeping it in the Argo namespace avoids coupling
// integration availability to upstream WhatsApp transport models.
type IntegrationHeartbeat struct {
	ID            string       `json:"id" gorm:"type:uuid;primaryKey"`
	ApplicationID string       `json:"application_id" gorm:"type:uuid;index:idx_argo_heartbeat_app_received,priority:1;not null"`
	Status        string       `json:"status" gorm:"size:24;index;not null"`
	LatencyMS     int64        `json:"latency_ms" gorm:"not null;default:0"`
	Version       string       `json:"version,omitempty" gorm:"size:80"`
	Component     string       `json:"component,omitempty" gorm:"size:100"`
	Message       string       `json:"message,omitempty" gorm:"size:500"`
	ReceivedAt    time.Time    `json:"received_at" gorm:"index:idx_argo_heartbeat_app_received,priority:2;index;not null"`
	CreatedAt     time.Time    `json:"created_at" gorm:"autoCreateTime"`
	Application   *Application `json:"application,omitempty" gorm:"foreignKey:ApplicationID"`
}

func (IntegrationHeartbeat) TableName() string { return "argo_integration_heartbeats" }

func (h *IntegrationHeartbeat) BeforeCreate(tx *gorm.DB) error {
	if h.ID == "" {
		h.ID = uuid.NewString()
	}
	return nil
}

type HeartbeatFilters struct {
	From            time.Time
	To              time.Time
	ApplicationSlug string
	Status          string
	Limit           int
}

type HealthSummary struct {
	From             time.Time `json:"from"`
	To               time.Time `json:"to"`
	Applications     int64     `json:"applications"`
	Healthy          int64     `json:"healthy"`
	Degraded         int64     `json:"degraded"`
	Offline          int64     `json:"offline"`
	Unknown          int64     `json:"unknown"`
	HeartbeatEvents  int64     `json:"heartbeat_events"`
	UnhealthyEvents  int64     `json:"unhealthy_events"`
	AverageLatencyMS float64   `json:"average_latency_ms"`
}

func (Application) TableName() string { return "argo_applications" }

func (a *Application) BeforeCreate(tx *gorm.DB) error {
	if a.ID == "" {
		a.ID = uuid.NewString()
	}
	return nil
}

// MessageAttempt records every request made to a send endpoint, including
// failures that never produced a WhatsApp message.
type MessageAttempt struct {
	ID                string       `json:"id" gorm:"type:uuid;primaryKey"`
	ApplicationID     *string      `json:"application_id,omitempty" gorm:"type:uuid;index:idx_argo_attempt_app_started,priority:1"`
	ApplicationSlug   string       `json:"application_slug" gorm:"size:100;index"`
	IdentityVerified  bool         `json:"identity_verified" gorm:"not null;default:false"`
	InstanceID        *string      `json:"instance_id,omitempty" gorm:"type:uuid;index:idx_argo_attempt_instance_started,priority:1"`
	CorrelationID     string       `json:"correlation_id" gorm:"size:64;index;not null"`
	IdempotencyKey    string       `json:"idempotency_key,omitempty" gorm:"size:128;index"`
	ProviderMessageID string       `json:"provider_message_id,omitempty" gorm:"size:128;index"`
	Endpoint          string       `json:"endpoint" gorm:"size:160;index;not null"`
	Method            string       `json:"method" gorm:"size:12;not null"`
	HTTPStatus        int          `json:"http_status" gorm:"index;not null"`
	Outcome           string       `json:"outcome" gorm:"size:24;index;not null"`
	ErrorCode         string       `json:"error_code,omitempty" gorm:"size:64;index"`
	ErrorDetail       string       `json:"error_detail,omitempty" gorm:"size:500"`
	DurationMS        int64        `json:"duration_ms" gorm:"not null"`
	StartedAt         time.Time    `json:"started_at" gorm:"index:idx_argo_attempt_app_started,priority:2;index:idx_argo_attempt_instance_started,priority:2;index"`
	CompletedAt       time.Time    `json:"completed_at"`
	CreatedAt         time.Time    `json:"created_at" gorm:"autoCreateTime"`
	Application       *Application `json:"application,omitempty" gorm:"foreignKey:ApplicationID"`
}

func (MessageAttempt) TableName() string { return "argo_message_attempts" }

func (a *MessageAttempt) BeforeCreate(tx *gorm.DB) error {
	if a.ID == "" {
		a.ID = uuid.NewString()
	}
	return nil
}

type AttemptFilters struct {
	From            time.Time
	To              time.Time
	ApplicationSlug string
	InstanceID      string
	Outcome         string
	ErrorCode       string
	Limit           int
}

type ErrorBreakdown struct {
	ErrorCode string `json:"error_code"`
	Count     int64  `json:"count"`
}

type AttemptSummary struct {
	From               time.Time        `json:"from"`
	To                 time.Time        `json:"to"`
	Total              int64            `json:"total"`
	Succeeded          int64            `json:"succeeded"`
	Failed             int64            `json:"failed"`
	UnverifiedIdentity int64            `json:"unverified_identity"`
	SuccessRate        float64          `json:"success_rate"`
	AverageDurationMS  float64          `json:"average_duration_ms"`
	Errors             []ErrorBreakdown `json:"errors"`
}

// MessageLifecycleEvent is an immutable fact in the operational lifecycle of
// a message. It deliberately duplicates correlation attributes from the
// attempt so historical evidence remains queryable even when an application
// or upstream message changes later.
type MessageLifecycleEvent struct {
	ID                string    `json:"id" gorm:"type:uuid;primaryKey"`
	EventKey          string    `json:"-" gorm:"size:320;uniqueIndex;not null"`
	AttemptID         *string   `json:"attempt_id,omitempty" gorm:"type:uuid;index"`
	ApplicationID     *string   `json:"application_id,omitempty" gorm:"type:uuid;index"`
	ApplicationSlug   string    `json:"application_slug" gorm:"size:100;index;not null"`
	IdentityVerified  bool      `json:"identity_verified" gorm:"not null;default:false"`
	InstanceID        *string   `json:"instance_id,omitempty" gorm:"type:uuid;index:idx_argo_lifecycle_instance_message,priority:1"`
	ProviderMessageID string    `json:"provider_message_id,omitempty" gorm:"size:128;index:idx_argo_lifecycle_instance_message,priority:2;index"`
	CorrelationID     string    `json:"correlation_id,omitempty" gorm:"size:64;index"`
	IdempotencyKey    string    `json:"idempotency_key,omitempty" gorm:"size:128;index"`
	MessageType       string    `json:"message_type,omitempty" gorm:"size:48;index"`
	State             string    `json:"state" gorm:"size:24;index;not null"`
	FailureCategory   string    `json:"failure_category,omitempty" gorm:"size:48;index"`
	FailureCode       string    `json:"failure_code,omitempty" gorm:"size:64;index"`
	FailureDetail     string    `json:"failure_detail,omitempty" gorm:"size:500"`
	OccurredAt        time.Time `json:"occurred_at" gorm:"index;not null"`
	CreatedAt         time.Time `json:"created_at" gorm:"autoCreateTime"`
}

// MessageMedia stores a captured attachment independently from the upstream
// message model. The binary is served only through the authenticated Argo API.
type MessageMedia struct {
	ID                string    `json:"id" gorm:"type:uuid;primaryKey"`
	InstanceID        string    `json:"instance_id" gorm:"type:uuid;uniqueIndex:idx_argo_media_instance_message,priority:1;not null"`
	ProviderMessageID string    `json:"provider_message_id" gorm:"size:128;uniqueIndex:idx_argo_media_instance_message,priority:2;not null"`
	FileName          string    `json:"file_name" gorm:"size:255"`
	MimeType          string    `json:"mime_type" gorm:"size:160;not null"`
	SizeBytes         int64     `json:"size_bytes" gorm:"not null"`
	Content           []byte    `json:"-" gorm:"type:bytea;not null"`
	CreatedAt         time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt         time.Time `json:"updated_at" gorm:"autoUpdateTime"`
}

func (MessageMedia) TableName() string { return "argo_message_media" }

func (m *MessageMedia) BeforeCreate(tx *gorm.DB) error {
	if m.ID == "" {
		m.ID = uuid.NewString()
	}
	return nil
}

func (MessageLifecycleEvent) TableName() string { return "argo_message_lifecycle_events" }

func (e *MessageLifecycleEvent) BeforeCreate(tx *gorm.DB) error {
	if e.ID == "" {
		e.ID = uuid.NewString()
	}
	return nil
}

type LifecycleFilters struct {
	From              time.Time
	To                time.Time
	ApplicationSlug   string
	InstanceID        string
	MessageType       string
	State             string
	ProviderMessageID string
	CorrelationID     string
	IdempotencyKey    string
	Limit             int
}

// LifecycleBackfillOptions bounds an explicit reconstruction of lifecycle
// evidence that predates the Argo lifecycle table.
type LifecycleBackfillOptions struct {
	From    time.Time
	To      time.Time
	Limit   int
	Execute bool
}

type LifecycleBackfillReport struct {
	From            time.Time `json:"from"`
	To              time.Time `json:"to"`
	Execute         bool      `json:"execute"`
	AttemptsScanned int       `json:"attempts_scanned"`
	MessagesScanned int       `json:"messages_scanned"`
	CandidateEvents int       `json:"candidate_events"`
	ExistingEvents  int       `json:"existing_events"`
	PendingEvents   int       `json:"pending_events"`
	EventsCreated   int64     `json:"events_created"`
}

type LifecycleLatency struct {
	P50 float64 `json:"p50_ms"`
	P90 float64 `json:"p90_ms"`
	P95 float64 `json:"p95_ms"`
	P99 float64 `json:"p99_ms"`
}

type LifecycleFailure struct {
	Category string `json:"category"`
	Code     string `json:"code"`
	Count    int64  `json:"count"`
}

type LifecycleSummary struct {
	From              time.Time          `json:"from"`
	To                time.Time          `json:"to"`
	Received          int64              `json:"received"`
	Validated         int64              `json:"validated"`
	Accepted          int64              `json:"accepted"`
	Sent              int64              `json:"sent"`
	Delivered         int64              `json:"delivered"`
	Read              int64              `json:"read"`
	Failed            int64              `json:"failed"`
	PendingAged       int64              `json:"pending_aged"`
	AcceptanceRate    float64            `json:"acceptance_rate"`
	SendRate          float64            `json:"send_rate"`
	DeliveryRate      float64            `json:"delivery_rate"`
	ReadRate          float64            `json:"read_rate"`
	SendLatency       LifecycleLatency   `json:"send_latency"`
	DeliveryLatency   LifecycleLatency   `json:"delivery_latency"`
	ReadLatency       LifecycleLatency   `json:"read_latency"`
	Failures          []LifecycleFailure `json:"failures"`
	PendingAgeMinutes int64              `json:"pending_age_minutes"`
}
