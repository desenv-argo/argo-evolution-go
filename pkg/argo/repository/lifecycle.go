package argo_repository

import (
	"context"
	"fmt"
	"path"
	"strings"
	"time"

	argo_model "github.com/evolution-foundation/evolution-go/pkg/argo/model"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

const maxLifecycleSummaryEvents = 100000

func (r *repository) recordAttemptLifecycle(tx *gorm.DB, attempt *argo_model.MessageAttempt) error {
	for _, event := range attemptLifecycleEvents(attempt) {
		if err := tx.Clauses(clause.OnConflict{DoNothing: true}).Create(&event).Error; err != nil {
			return err
		}
	}
	return nil
}

func attemptLifecycleEvents(attempt *argo_model.MessageAttempt) []argo_model.MessageLifecycleEvent {
	base := argo_model.MessageLifecycleEvent{
		AttemptID:         &attempt.ID,
		ApplicationID:     attempt.ApplicationID,
		ApplicationSlug:   attempt.ApplicationSlug,
		IdentityVerified:  attempt.IdentityVerified,
		InstanceID:        attempt.InstanceID,
		ProviderMessageID: attempt.ProviderMessageID,
		CorrelationID:     attempt.CorrelationID,
		IdempotencyKey:    attempt.IdempotencyKey,
		MessageType:       strings.TrimPrefix(path.Base(attempt.Endpoint), "/"),
	}
	if base.ApplicationSlug == "" {
		base.ApplicationSlug = "legacy/unknown"
	}

	states := []struct {
		state string
		at    time.Time
	}{{state: "received", at: attempt.StartedAt}}
	if attempt.Outcome == "succeeded" {
		states = append(states,
			struct {
				state string
				at    time.Time
			}{state: "validated", at: attempt.CompletedAt},
			struct {
				state string
				at    time.Time
			}{state: "accepted", at: attempt.CompletedAt},
			struct {
				state string
				at    time.Time
			}{state: "sent", at: attempt.CompletedAt},
		)
	} else {
		if acceptedBeforeFailure(attempt.ErrorCode) {
			states = append(states,
				struct {
					state string
					at    time.Time
				}{state: "validated", at: attempt.CompletedAt},
				struct {
					state string
					at    time.Time
				}{state: "accepted", at: attempt.CompletedAt},
			)
		}
		states = append(states, struct {
			state string
			at    time.Time
		}{state: "failed", at: attempt.CompletedAt})
	}

	events := make([]argo_model.MessageLifecycleEvent, 0, len(states))
	for _, transition := range states {
		event := base
		event.EventKey = fmt.Sprintf("attempt:%s:%s", attempt.ID, transition.state)
		event.State = transition.state
		event.OccurredAt = transition.at.UTC()
		if transition.state == "failed" {
			event.FailureCode = attempt.ErrorCode
			event.FailureCategory = failureCategory(attempt.ErrorCode)
			event.FailureDetail = attempt.ErrorDetail
		}
		events = append(events, event)
	}
	return events
}

func acceptedBeforeFailure(code string) bool {
	switch code {
	case "AUTH_FAILED", "RATE_LIMITED", "INVALID_RECIPIENT", "VALIDATION_FAILED":
		return false
	default:
		return true
	}
}

func (r *repository) RecordReceipt(ctx context.Context, instanceID string, providerMessageIDs []string, state string, occurredAt time.Time) error {
	if state != "delivered" && state != "read" {
		return fmt.Errorf("unsupported receipt lifecycle state %q", state)
	}
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		for _, providerMessageID := range providerMessageIDs {
			providerMessageID = strings.TrimSpace(providerMessageID)
			if providerMessageID == "" {
				continue
			}
			var attempt argo_model.MessageAttempt
			lookup := tx.Where("instance_id = ? AND provider_message_id = ?", instanceID, providerMessageID).
				Order("completed_at DESC").Limit(1).Find(&attempt)
			if lookup.Error != nil {
				return lookup.Error
			}
			event := argo_model.MessageLifecycleEvent{
				EventKey:          fmt.Sprintf("receipt:%s:%s:%s", instanceID, providerMessageID, state),
				ApplicationSlug:   "legacy/unknown",
				InstanceID:        &instanceID,
				ProviderMessageID: providerMessageID,
				State:             state,
				OccurredAt:        occurredAt.UTC(),
			}
			if lookup.RowsAffected > 0 {
				event.AttemptID = &attempt.ID
				event.ApplicationID = attempt.ApplicationID
				event.ApplicationSlug = attempt.ApplicationSlug
				event.IdentityVerified = attempt.IdentityVerified
				event.CorrelationID = attempt.CorrelationID
				event.IdempotencyKey = attempt.IdempotencyKey
				event.MessageType = strings.TrimPrefix(path.Base(attempt.Endpoint), "/")
			}
			if err := tx.Clauses(clause.OnConflict{DoNothing: true}).Create(&event).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

func (r *repository) ReconcilePendingAged(ctx context.Context, cutoff time.Time) (int64, error) {
	var sent []argo_model.MessageLifecycleEvent
	err := r.db.WithContext(ctx).
		Where("state = ? AND occurred_at <= ?", "sent", cutoff.UTC()).
		Where(`NOT EXISTS (
			SELECT 1 FROM argo_message_lifecycle_events terminal
			WHERE terminal.provider_message_id = argo_message_lifecycle_events.provider_message_id
			AND terminal.instance_id = argo_message_lifecycle_events.instance_id
			AND terminal.state IN ('delivered', 'read', 'failed', 'pending_aged')
		)`).
		Limit(2000).
		Find(&sent).Error
	if err != nil {
		return 0, err
	}
	var created int64
	for _, source := range sent {
		event := source
		event.ID = ""
		event.EventKey = "pending:" + source.EventKey
		event.State = "pending_aged"
		event.OccurredAt = cutoff.UTC()
		event.CreatedAt = time.Time{}
		result := r.db.WithContext(ctx).Clauses(clause.OnConflict{DoNothing: true}).Create(&event)
		if result.Error != nil {
			return created, result.Error
		}
		created += result.RowsAffected
	}
	return created, nil
}

func (r *repository) ListLifecycleEvents(ctx context.Context, filters argo_model.LifecycleFilters) ([]argo_model.MessageLifecycleEvent, error) {
	limit := filters.Limit
	if limit <= 0 || limit > 500 {
		limit = 200
	}
	events := make([]argo_model.MessageLifecycleEvent, 0)
	err := r.scopedLifecycleEvents(ctx, filters).
		Order("occurred_at DESC, id DESC").
		Limit(limit).
		Find(&events).Error
	return events, err
}

func (r *repository) LifecycleEvents(ctx context.Context, filters argo_model.LifecycleFilters) ([]argo_model.MessageLifecycleEvent, error) {
	events := make([]argo_model.MessageLifecycleEvent, 0)
	err := r.scopedLifecycleEvents(ctx, filters).
		Order("occurred_at ASC, id ASC").
		Limit(maxLifecycleSummaryEvents).
		Find(&events).Error
	return events, err
}

func (r *repository) scopedLifecycleEvents(ctx context.Context, filters argo_model.LifecycleFilters) *gorm.DB {
	query := r.db.WithContext(ctx).Model(&argo_model.MessageLifecycleEvent{})
	if !filters.From.IsZero() {
		query = query.Where("occurred_at >= ?", filters.From)
	}
	if !filters.To.IsZero() {
		query = query.Where("occurred_at <= ?", filters.To)
	}
	if filters.ApplicationSlug != "" {
		query = query.Where("application_slug = ?", filters.ApplicationSlug)
	}
	if filters.InstanceID != "" {
		query = query.Where("instance_id = ?", filters.InstanceID)
	}
	if filters.MessageType != "" {
		query = query.Where("message_type = ?", filters.MessageType)
	}
	if filters.State != "" {
		query = query.Where("state = ?", filters.State)
	}
	if filters.ProviderMessageID != "" {
		query = query.Where("provider_message_id = ?", filters.ProviderMessageID)
	}
	if filters.CorrelationID != "" {
		query = query.Where("correlation_id = ?", filters.CorrelationID)
	}
	if filters.IdempotencyKey != "" {
		query = query.Where("idempotency_key = ?", filters.IdempotencyKey)
	}
	return query
}

func failureCategory(code string) string {
	switch code {
	case "AUTH_FAILED":
		return "authentication"
	case "RATE_LIMITED":
		return "throttling"
	case "INSTANCE_OFFLINE":
		return "availability"
	case "INVALID_RECIPIENT", "VALIDATION_FAILED":
		return "validation"
	case "WHATSAPP_TIMEOUT":
		return "transport"
	case "MEDIA_FAILED":
		return "media"
	default:
		return "internal"
	}
}
