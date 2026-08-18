package argo_repository

import (
	"context"
	"errors"

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

func NewRepository(db *gorm.DB) Repository {
	return &repository{db: db}
}
