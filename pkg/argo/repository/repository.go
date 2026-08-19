package argo_repository

import (
	"context"
	"errors"
	"time"

	argo_model "github.com/evolution-foundation/evolution-go/pkg/argo/model"
	"gorm.io/gorm"
)

type Repository interface {
	CreateApplication(ctx context.Context, application *argo_model.Application) error
	SaveApplication(ctx context.Context, application *argo_model.Application) error
	GetApplicationByID(ctx context.Context, id string) (*argo_model.Application, error)
	GetApplicationBySlug(ctx context.Context, slug string) (*argo_model.Application, error)
	ListApplications(ctx context.Context) ([]argo_model.Application, error)
	RecordAttempt(ctx context.Context, attempt *argo_model.MessageAttempt) error
	ListAttempts(ctx context.Context, filters argo_model.AttemptFilters) ([]argo_model.MessageAttempt, error)
	AttemptSummary(ctx context.Context, filters argo_model.AttemptFilters) (*argo_model.AttemptSummary, error)
	GatewayUsage(ctx context.Context, filters argo_model.AttemptFilters) ([]argo_model.GatewayUsage, []argo_model.GatewayUsage, error)
	RecordHeartbeat(ctx context.Context, heartbeat *argo_model.IntegrationHeartbeat) error
	ListHeartbeats(ctx context.Context, filters argo_model.HeartbeatFilters) ([]argo_model.IntegrationHeartbeat, error)
	HeartbeatMetrics(ctx context.Context, filters argo_model.HeartbeatFilters) (int64, int64, float64, error)
	RecordReceipt(ctx context.Context, instanceID string, providerMessageIDs []string, state string, occurredAt time.Time) error
	ListLifecycleEvents(ctx context.Context, filters argo_model.LifecycleFilters) ([]argo_model.MessageLifecycleEvent, error)
	ListLifecycleFeed(ctx context.Context, applicationID string, cursor argo_model.LifecycleFeedCursor, limit int) ([]argo_model.MessageLifecycleEvent, error)
	LifecycleEvents(ctx context.Context, filters argo_model.LifecycleFilters) ([]argo_model.MessageLifecycleEvent, error)
	ReconcilePendingAged(ctx context.Context, cutoff time.Time) (int64, error)
	BackfillLifecycle(ctx context.Context, options argo_model.LifecycleBackfillOptions) (*argo_model.LifecycleBackfillReport, error)
	SaveMessageMedia(ctx context.Context, media *argo_model.MessageMedia) error
	GetMessageMedia(ctx context.Context, instanceID, providerMessageID string) (*argo_model.MessageMedia, error)
	DeleteMessageMediaBefore(ctx context.Context, cutoff time.Time, limit int) (int64, error)
	SaveUpstreamSnapshot(ctx context.Context, snapshot *argo_model.UpstreamSnapshot) error
	LatestUpstreamSnapshot(ctx context.Context) (*argo_model.UpstreamSnapshot, error)
}

func (r *repository) GatewayUsage(ctx context.Context, filters argo_model.AttemptFilters) ([]argo_model.GatewayUsage, []argo_model.GatewayUsage, error) {
	applications := make([]argo_model.GatewayUsage, 0)
	if err := r.scopedAttempts(ctx, filters).Select(`
		CASE WHEN identity_verified = TRUE AND application_slug <> '' THEN application_slug ELSE 'legacy/unknown' END AS key,
		COUNT(*) AS total,
		COALESCE(SUM(CASE WHEN outcome = 'succeeded' THEN 1 ELSE 0 END), 0) AS succeeded,
		COALESCE(SUM(CASE WHEN outcome = 'failed' THEN 1 ELSE 0 END), 0) AS failed,
		COALESCE(SUM(CASE WHEN identity_verified = FALSE THEN 1 ELSE 0 END), 0) AS unverified_identity,
		COALESCE(AVG(duration_ms), 0) AS average_duration_ms,
		MAX(completed_at) AS last_activity_at
	`).Group("key").Order("total DESC").Scan(&applications).Error; err != nil {
		return nil, nil, err
	}

	instances := make([]argo_model.GatewayUsage, 0)
	if err := r.scopedAttempts(ctx, filters).Select(`
		COALESCE(CAST(instance_id AS TEXT), 'unknown') AS key,
		COUNT(*) AS total,
		COALESCE(SUM(CASE WHEN outcome = 'succeeded' THEN 1 ELSE 0 END), 0) AS succeeded,
		COALESCE(SUM(CASE WHEN outcome = 'failed' THEN 1 ELSE 0 END), 0) AS failed,
		COALESCE(SUM(CASE WHEN identity_verified = FALSE THEN 1 ELSE 0 END), 0) AS unverified_identity,
		COALESCE(AVG(duration_ms), 0) AS average_duration_ms,
		MAX(completed_at) AS last_activity_at
	`).Group("instance_id").Order("total DESC").Scan(&instances).Error; err != nil {
		return nil, nil, err
	}
	return applications, instances, nil
}

func (r *repository) RecordHeartbeat(ctx context.Context, heartbeat *argo_model.IntegrationHeartbeat) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(heartbeat).Error; err != nil {
			return err
		}
		return tx.Model(&argo_model.Application{}).
			Where("id = ?", heartbeat.ApplicationID).
			Updates(map[string]interface{}{
				"last_heartbeat_at":         heartbeat.ReceivedAt,
				"last_heartbeat_status":     heartbeat.Status,
				"last_heartbeat_latency_ms": heartbeat.LatencyMS,
				"last_heartbeat_version":    heartbeat.Version,
			}).Error
	})
}

func (r *repository) ListHeartbeats(ctx context.Context, filters argo_model.HeartbeatFilters) ([]argo_model.IntegrationHeartbeat, error) {
	limit := filters.Limit
	if limit <= 0 || limit > 200 {
		limit = 100
	}
	heartbeats := make([]argo_model.IntegrationHeartbeat, 0)
	err := r.scopedHeartbeats(ctx, filters).
		Preload("Application").
		Order("received_at DESC, id DESC").
		Limit(limit).
		Find(&heartbeats).Error
	return heartbeats, err
}

func (r *repository) HeartbeatMetrics(ctx context.Context, filters argo_model.HeartbeatFilters) (int64, int64, float64, error) {
	var aggregate struct {
		HeartbeatEvents  int64   `gorm:"column:heartbeat_events"`
		UnhealthyEvents  int64   `gorm:"column:unhealthy_events"`
		AverageLatencyMS float64 `gorm:"column:average_latency_ms"`
	}
	err := r.scopedHeartbeats(ctx, filters).Select(`
		COUNT(*) AS heartbeat_events,
		COALESCE(SUM(CASE WHEN argo_integration_heartbeats.status <> 'healthy' THEN 1 ELSE 0 END), 0) AS unhealthy_events,
		CAST(COALESCE(AVG(argo_integration_heartbeats.latency_ms), 0) AS DOUBLE PRECISION) AS average_latency_ms
	`).Scan(&aggregate).Error
	return aggregate.HeartbeatEvents, aggregate.UnhealthyEvents, aggregate.AverageLatencyMS, err
}

type repository struct {
	db *gorm.DB
}

func (r *repository) CreateApplication(ctx context.Context, application *argo_model.Application) error {
	return r.db.WithContext(ctx).Create(application).Error
}

func (r *repository) SaveApplication(ctx context.Context, application *argo_model.Application) error {
	return r.db.WithContext(ctx).Save(application).Error
}

func (r *repository) GetApplicationByID(ctx context.Context, id string) (*argo_model.Application, error) {
	var application argo_model.Application
	err := r.db.WithContext(ctx).First(&application, "id = ?", id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	return &application, err
}

func (r *repository) GetApplicationBySlug(ctx context.Context, slug string) (*argo_model.Application, error) {
	var application argo_model.Application
	err := r.db.WithContext(ctx).First(&application, "slug = ?", slug).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	return &application, err
}

func (r *repository) ListApplications(ctx context.Context) ([]argo_model.Application, error) {
	applications := make([]argo_model.Application, 0)
	err := r.db.WithContext(ctx).Order("name ASC, environment ASC").Find(&applications).Error
	return applications, err
}

func (r *repository) RecordAttempt(ctx context.Context, attempt *argo_model.MessageAttempt) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(attempt).Error; err != nil {
			return err
		}
		if err := r.recordAttemptLifecycle(tx, attempt); err != nil {
			return err
		}
		if attempt.ApplicationID != nil && attempt.IdentityVerified {
			return tx.Model(&argo_model.Application{}).
				Where("id = ?", *attempt.ApplicationID).
				Update("last_seen_at", attempt.CompletedAt).Error
		}
		return nil
	})
}

func (r *repository) ListAttempts(ctx context.Context, filters argo_model.AttemptFilters) ([]argo_model.MessageAttempt, error) {
	limit := filters.Limit
	if limit <= 0 || limit > 200 {
		limit = 100
	}
	query := r.scopedAttempts(ctx, filters)
	attempts := make([]argo_model.MessageAttempt, 0)
	err := query.Order("started_at DESC, id DESC").Limit(limit).Find(&attempts).Error
	return attempts, err
}

func (r *repository) AttemptSummary(ctx context.Context, filters argo_model.AttemptFilters) (*argo_model.AttemptSummary, error) {
	summary := &argo_model.AttemptSummary{From: filters.From, To: filters.To, Errors: []argo_model.ErrorBreakdown{}}
	var aggregate struct {
		Total              int64
		Succeeded          int64
		Failed             int64
		UnverifiedIdentity int64
		AverageDurationMS  float64
	}
	if err := r.scopedAttempts(ctx, filters).Select(`
		COUNT(*) AS total,
		COALESCE(SUM(CASE WHEN outcome = 'succeeded' THEN 1 ELSE 0 END), 0) AS succeeded,
		COALESCE(SUM(CASE WHEN outcome = 'failed' THEN 1 ELSE 0 END), 0) AS failed,
		COALESCE(SUM(CASE WHEN identity_verified = FALSE THEN 1 ELSE 0 END), 0) AS unverified_identity,
		COALESCE(AVG(duration_ms), 0) AS average_duration_ms
	`).Scan(&aggregate).Error; err != nil {
		return nil, err
	}
	summary.Total = aggregate.Total
	summary.Succeeded = aggregate.Succeeded
	summary.Failed = aggregate.Failed
	summary.UnverifiedIdentity = aggregate.UnverifiedIdentity
	summary.AverageDurationMS = aggregate.AverageDurationMS
	if aggregate.Total > 0 {
		summary.SuccessRate = float64(aggregate.Succeeded) * 100 / float64(aggregate.Total)
	}

	if err := r.scopedAttempts(ctx, filters).
		Select("error_code, COUNT(*) AS count").
		Where("error_code <> ''").
		Group("error_code").
		Order("count DESC").
		Scan(&summary.Errors).Error; err != nil {
		return nil, err
	}
	return summary, nil
}

func (r *repository) scopedAttempts(ctx context.Context, filters argo_model.AttemptFilters) *gorm.DB {
	query := r.db.WithContext(ctx).Model(&argo_model.MessageAttempt{})
	if !filters.From.IsZero() {
		query = query.Where("started_at >= ?", filters.From)
	}
	if !filters.To.IsZero() {
		query = query.Where("started_at <= ?", filters.To)
	}
	if filters.ApplicationSlug != "" {
		query = query.Where("application_slug = ?", filters.ApplicationSlug)
	}
	if filters.InstanceID != "" {
		query = query.Where("instance_id = ?", filters.InstanceID)
	}
	if filters.Outcome != "" {
		query = query.Where("outcome = ?", filters.Outcome)
	}
	if filters.ErrorCode != "" {
		query = query.Where("error_code = ?", filters.ErrorCode)
	}
	return query
}

func (r *repository) scopedHeartbeats(ctx context.Context, filters argo_model.HeartbeatFilters) *gorm.DB {
	query := r.db.WithContext(ctx).Model(&argo_model.IntegrationHeartbeat{})
	if !filters.From.IsZero() {
		query = query.Where("received_at >= ?", filters.From)
	}
	if !filters.To.IsZero() {
		query = query.Where("received_at <= ?", filters.To)
	}
	if filters.ApplicationSlug != "" {
		query = query.Joins("JOIN argo_applications ON argo_applications.id = argo_integration_heartbeats.application_id").
			Where("argo_applications.slug = ?", filters.ApplicationSlug)
	}
	if filters.Status != "" {
		query = query.Where("argo_integration_heartbeats.status = ?", filters.Status)
	}
	return query
}

func NewRepository(db *gorm.DB) Repository {
	return &repository{db: db}
}
