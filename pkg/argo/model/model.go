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
	CreatedAt                time.Time  `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt                time.Time  `json:"updated_at" gorm:"autoUpdateTime"`
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
