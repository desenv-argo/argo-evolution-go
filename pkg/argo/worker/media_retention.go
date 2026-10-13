package argo_worker

import (
	"context"
	"os"
	"strconv"
	"time"
)

// Captured attachments live in the same PostgreSQL server as the Argo ERP. On
// 2026-10-13 a 30-day window let argo_message_media reach 24 GB, the server ran
// out of storage and went read-only, taking the ERP down with it. The window is
// now short (3 days) and a size budget removes the oldest attachments as soon as
// the table goes over it, whatever their age.
const (
	defaultMediaRetention        = 3 * 24 * time.Hour
	defaultMediaCleanupInterval  = 15 * time.Minute
	defaultMediaCleanupTimeout   = 30 * time.Second
	defaultMediaCleanupBatch     = 500
	defaultMediaMaxTotalBytes    = int64(2 * 1024 * 1024 * 1024)
	minimumMediaMaxTotalBytes    = int64(64 * 1024 * 1024)
	maximumMediaMaxTotalBytes    = int64(100 * 1024 * 1024 * 1024)
	maxMediaCleanupBatchesPerRun = 100
)

type MessageMediaCleaner interface {
	DeleteMessageMediaBefore(ctx context.Context, cutoff time.Time, limit int) (int64, error)
	DeleteMessageMediaOverBudget(ctx context.Context, maxTotalBytes int64, limit int) (int64, error)
}

type MediaRetentionConfig struct {
	Retention     time.Duration
	Interval      time.Duration
	Timeout       time.Duration
	BatchSize     int
	MaxTotalBytes int64
}

type MediaRetentionLogger struct {
	Info  func(format string, args ...interface{})
	Error func(format string, args ...interface{})
}

type MediaRetentionWorker struct {
	cleaner MessageMediaCleaner
	config  MediaRetentionConfig
	logger  MediaRetentionLogger
}

func NewMediaRetentionWorker(cleaner MessageMediaCleaner, config MediaRetentionConfig, logger MediaRetentionLogger) *MediaRetentionWorker {
	if config.Retention <= 0 {
		config.Retention = defaultMediaRetention
	}
	if config.Interval <= 0 {
		config.Interval = defaultMediaCleanupInterval
	}
	if config.Timeout <= 0 {
		config.Timeout = defaultMediaCleanupTimeout
	}
	if config.BatchSize <= 0 || config.BatchSize > 5000 {
		config.BatchSize = defaultMediaCleanupBatch
	}
	if config.MaxTotalBytes <= 0 {
		config.MaxTotalBytes = defaultMediaMaxTotalBytes
	}
	return &MediaRetentionWorker{cleaner: cleaner, config: config, logger: logger}
}

func MediaRetentionConfigFromEnvironment() MediaRetentionConfig {
	return MediaRetentionConfig{
		Retention:     durationFromDays("ARGO_MESSAGE_MEDIA_RETENTION_DAYS", defaultMediaRetention, 24*time.Hour, 365*24*time.Hour),
		Interval:      durationFromSeconds("ARGO_MESSAGE_MEDIA_CLEANUP_SECONDS", defaultMediaCleanupInterval, time.Minute, 24*time.Hour),
		Timeout:       durationFromSeconds("ARGO_MESSAGE_MEDIA_CLEANUP_TIMEOUT_SECONDS", defaultMediaCleanupTimeout, time.Second, 5*time.Minute),
		BatchSize:     integerFromEnvironment("ARGO_MESSAGE_MEDIA_CLEANUP_BATCH", defaultMediaCleanupBatch, 1, 5000),
		MaxTotalBytes: int64FromEnvironment("ARGO_MESSAGE_MEDIA_MAX_TOTAL_BYTES", defaultMediaMaxTotalBytes, minimumMediaMaxTotalBytes, maximumMediaMaxTotalBytes),
	}
}

func (w *MediaRetentionWorker) Run(ctx context.Context) {
	w.runOnce(ctx)
	ticker := time.NewTicker(w.config.Interval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			w.runOnce(ctx)
		}
	}
}

// runOnce removes attachments older than the window and then whatever is over the
// size budget. Each step repeats full batches until a short one shows nothing is
// left, bounded by the cycle timeout.
func (w *MediaRetentionWorker) runOnce(parent context.Context) {
	startedAt := time.Now().UTC()
	ctx, cancel := context.WithTimeout(parent, w.config.Timeout)
	defer cancel()

	cutoff := startedAt.Add(-w.config.Retention)
	expired, err := w.drain(ctx, func(ctx context.Context) (int64, error) {
		return w.cleaner.DeleteMessageMediaBefore(ctx, cutoff, w.config.BatchSize)
	})
	if err != nil {
		w.logError("[ARGO_MEDIA] retention cleanup failed after %s (deleted=%d): %v", time.Since(startedAt), expired, err)
		return
	}
	overBudget, err := w.drain(ctx, func(ctx context.Context) (int64, error) {
		return w.cleaner.DeleteMessageMediaOverBudget(ctx, w.config.MaxTotalBytes, w.config.BatchSize)
	})
	if err != nil {
		w.logError("[ARGO_MEDIA] budget cleanup failed after %s (expired=%d, over_budget=%d): %v", time.Since(startedAt), expired, overBudget, err)
		return
	}
	if expired+overBudget > 0 && w.logger.Info != nil {
		w.logger.Info("[ARGO_MEDIA] cleanup completed: expired=%d over_budget=%d duration=%s", expired, overBudget, time.Since(startedAt))
	}
}

func (w *MediaRetentionWorker) drain(ctx context.Context, deleteBatch func(context.Context) (int64, error)) (int64, error) {
	var total int64
	for batch := 0; batch < maxMediaCleanupBatchesPerRun; batch++ {
		deleted, err := deleteBatch(ctx)
		total += deleted
		if err != nil {
			return total, err
		}
		if deleted < int64(w.config.BatchSize) || ctx.Err() != nil {
			break
		}
	}
	return total, nil
}

func (w *MediaRetentionWorker) logError(format string, args ...interface{}) {
	if w.logger.Error != nil {
		w.logger.Error(format, args...)
	}
}

func durationFromDays(name string, fallback, minimum, maximum time.Duration) time.Duration {
	days, err := strconv.Atoi(os.Getenv(name))
	if err != nil {
		return fallback
	}
	value := time.Duration(days) * 24 * time.Hour
	if value < minimum || value > maximum {
		return fallback
	}
	return value
}

func integerFromEnvironment(name string, fallback, minimum, maximum int) int {
	value, err := strconv.Atoi(os.Getenv(name))
	if err != nil || value < minimum || value > maximum {
		return fallback
	}
	return value
}

func int64FromEnvironment(name string, fallback, minimum, maximum int64) int64 {
	value, err := strconv.ParseInt(os.Getenv(name), 10, 64)
	if err != nil || value < minimum || value > maximum {
		return fallback
	}
	return value
}
