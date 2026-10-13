package argo_repository

import (
	"context"
	"time"

	argo_model "github.com/evolution-foundation/evolution-go/pkg/argo/model"
	"gorm.io/gorm/clause"
)

func (r *repository) SaveMessageMedia(ctx context.Context, media *argo_model.MessageMedia) error {
	return r.db.WithContext(ctx).Clauses(clause.OnConflict{
		Columns: []clause.Column{{Name: "instance_id"}, {Name: "provider_message_id"}},
		DoUpdates: clause.AssignmentColumns([]string{
			"file_name", "mime_type", "size_bytes", "content", "updated_at",
		}),
	}).Create(media).Error
}

func (r *repository) DeleteMessageMediaBefore(ctx context.Context, cutoff time.Time, limit int) (int64, error) {
	if limit <= 0 || limit > 5000 {
		limit = 500
	}
	result := r.db.WithContext(ctx).Exec(`
		DELETE FROM argo_message_media
		WHERE id IN (
			SELECT id FROM argo_message_media
			WHERE created_at < ?
			ORDER BY created_at ASC
			LIMIT ?
		)`, cutoff, limit)
	return result.RowsAffected, result.Error
}

// DeleteMessageMediaOverBudget keeps the newest attachments whose summed size fits
// maxTotalBytes and removes the rest, oldest first, at most limit rows per call.
// The age-based retention alone let the table reach 24 GB and fill the database
// server shared with the Argo ERP (2026-10-13): the budget caps it regardless of age.
func (r *repository) DeleteMessageMediaOverBudget(ctx context.Context, maxTotalBytes int64, limit int) (int64, error) {
	if limit <= 0 || limit > 5000 {
		limit = 500
	}
	result := r.db.WithContext(ctx).Exec(`
		DELETE FROM argo_message_media
		WHERE id IN (
			SELECT id FROM (
				SELECT id, SUM(size_bytes) OVER (ORDER BY created_at DESC, id DESC) AS kept_bytes
				FROM argo_message_media
			) ranked
			WHERE kept_bytes > ?
			ORDER BY kept_bytes DESC
			LIMIT ?
		)`, maxTotalBytes, limit)
	return result.RowsAffected, result.Error
}

func (r *repository) GetMessageMedia(ctx context.Context, instanceID, providerMessageID string) (*argo_model.MessageMedia, error) {
	var media argo_model.MessageMedia
	err := r.db.WithContext(ctx).
		Where("instance_id = ? AND provider_message_id = ?", instanceID, providerMessageID).
		Limit(1).Find(&media).Error
	if err != nil {
		return nil, err
	}
	if media.ID == "" {
		return nil, nil
	}
	return &media, nil
}
