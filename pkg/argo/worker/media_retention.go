package argo_worker

import (
	"context"
	"os"
	"strconv"
	"time"
)

const (
	defaultMediaRetention       = 30 * 24 * time.Hour
	defaultMediaCleanupInterval = 6 * time.Hour
	defaultMediaCleanupTimeout  = 30 * time.Second
	defaultMediaCleanupBatch    = 500
)

type MessageMediaCleaner interface {
	DeleteMessageMediaBefore(ctx context.Context, cutoff time.Time, limit int) (int64, error)
}

type MediaRetentionConfig struct {
	Retention time.Duration
	Interval  time.Duration
	Timeout   time.Duration
	BatchSize int
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
	return &MediaRetentionWorker{cleaner: cleaner, config: config, logger: logger}
}

func MediaRetentionConfigFromEnvironment() MediaRetentionConfig {
	return MediaRetentionConfig{
		Retention: durationFromDays("ARGO_MESSAGE_MEDIA_RETENTION_DAYS", defaultMediaRetention, 24*time.Hour, 365*24*time.Hour),
		Interval:  durationFromSeconds("ARGO_MESSAGE_MEDIA_CLEANUP_SECONDS", defaultMediaCleanupInterval, time.Minute, 24*time.Hour),
		Timeout:   durationFromSeconds("ARGO_MESSAGE_MEDIA_CLEANUP_TIMEOUT_SECONDS", defaultMediaCleanupTimeout, time.Second, 5*time.Minute),
		BatchSize: integerFromEnvironment("ARGO_MESSAGE_MEDIA_CLEANUP_BATCH", defaultMediaCleanupBatch, 1, 5000),
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

func (w *MediaRetentionWorker) runOnce(parent context.Context) {
	startedAt := time.Now().UTC()
	ctx, cancel := context.WithTimeout(parent, w.config.Timeout)
	deleted, err := w.cleaner.DeleteMessageMediaBefore(ctx, startedAt.Add(-w.config.Retention), w.config.BatchSize)
	cancel()
	if err != nil {
		if w.logger.Error != nil {
			w.logger.Error("[ARGO_MEDIA] retention cleanup failed after %s: %v", time.Since(startedAt), err)
		}
		return
	}
	if deleted > 0 && w.logger.Info != nil {
		w.logger.Info("[ARGO_MEDIA] retention cleanup completed: deleted=%d duration=%s", deleted, time.Since(startedAt))
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
