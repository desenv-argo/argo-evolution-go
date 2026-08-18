package argo_worker

import (
	"context"
	"testing"
	"time"
)

type fakeMediaCleaner struct {
	cutoff time.Time
	limit  int
	called chan struct{}
}

func (f *fakeMediaCleaner) DeleteMessageMediaBefore(_ context.Context, cutoff time.Time, limit int) (int64, error) {
	f.cutoff, f.limit = cutoff, limit
	select {
	case f.called <- struct{}{}:
	default:
	}
	return 3, nil
}

func TestMediaRetentionWorkerRunsImmediately(t *testing.T) {
	cleaner := &fakeMediaCleaner{called: make(chan struct{}, 1)}
	worker := NewMediaRetentionWorker(cleaner, MediaRetentionConfig{Retention: 48 * time.Hour, Interval: time.Hour, Timeout: time.Second, BatchSize: 25}, MediaRetentionLogger{})
	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan struct{})
	go func() { worker.Run(ctx); close(done) }()
	select {
	case <-cleaner.called:
	case <-time.After(time.Second):
		t.Fatal("worker did not run on startup")
	}
	cancel()
	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("worker did not stop")
	}
	if cleaner.limit != 25 {
		t.Fatalf("limit = %d, want 25", cleaner.limit)
	}
	if age := time.Since(cleaner.cutoff); age < 47*time.Hour || age > 49*time.Hour {
		t.Fatalf("unexpected cutoff age: %s", age)
	}
}

func TestMediaRetentionConfigFromEnvironment(t *testing.T) {
	t.Setenv("ARGO_MESSAGE_MEDIA_RETENTION_DAYS", "45")
	t.Setenv("ARGO_MESSAGE_MEDIA_CLEANUP_SECONDS", "3600")
	t.Setenv("ARGO_MESSAGE_MEDIA_CLEANUP_TIMEOUT_SECONDS", "20")
	t.Setenv("ARGO_MESSAGE_MEDIA_CLEANUP_BATCH", "750")
	config := MediaRetentionConfigFromEnvironment()
	if config.Retention != 45*24*time.Hour || config.Interval != time.Hour || config.Timeout != 20*time.Second || config.BatchSize != 750 {
		t.Fatalf("unexpected config: %+v", config)
	}
}

func TestMediaRetentionConfigRejectsUnsafeValues(t *testing.T) {
	t.Setenv("ARGO_MESSAGE_MEDIA_RETENTION_DAYS", "0")
	t.Setenv("ARGO_MESSAGE_MEDIA_CLEANUP_SECONDS", "10")
	t.Setenv("ARGO_MESSAGE_MEDIA_CLEANUP_BATCH", "99999")
	config := MediaRetentionConfigFromEnvironment()
	if config.Retention != defaultMediaRetention || config.Interval != defaultMediaCleanupInterval || config.BatchSize != defaultMediaCleanupBatch {
		t.Fatalf("unexpected fallback config: %+v", config)
	}
}
