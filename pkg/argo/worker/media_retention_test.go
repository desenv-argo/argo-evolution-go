package argo_worker

import (
	"context"
	"sync"
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

func (f *fakeMediaCleaner) DeleteMessageMediaOverBudget(_ context.Context, _ int64, _ int) (int64, error) {
	return 0, nil
}

// scriptedMediaCleaner answers each call with the next scripted count and records
// the order of the calls.
type scriptedMediaCleaner struct {
	mu         sync.Mutex
	expired    []int64
	overBudget []int64
	calls      []string
	budgets    []int64
}

func (s *scriptedMediaCleaner) DeleteMessageMediaBefore(_ context.Context, _ time.Time, _ int) (int64, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.calls = append(s.calls, "expired")
	return next(&s.expired), nil
}

func (s *scriptedMediaCleaner) DeleteMessageMediaOverBudget(_ context.Context, maxTotalBytes int64, _ int) (int64, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.calls = append(s.calls, "budget")
	s.budgets = append(s.budgets, maxTotalBytes)
	return next(&s.overBudget), nil
}

func next(script *[]int64) int64 {
	if len(*script) == 0 {
		return 0
	}
	value := (*script)[0]
	*script = (*script)[1:]
	return value
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

// One cycle removes everything it can: full batches repeat until a short one, first
// the expired attachments, then what is over the size budget — even when nothing
// expired, which is what happens when a burst of media arrives within the window.
func TestMediaRetentionCycleDrainsExpiredThenEnforcesTheBudget(t *testing.T) {
	cleaner := &scriptedMediaCleaner{expired: []int64{25, 25, 4}, overBudget: []int64{25, 0}}
	worker := NewMediaRetentionWorker(cleaner, MediaRetentionConfig{Retention: 72 * time.Hour, Interval: time.Hour, Timeout: time.Second, BatchSize: 25, MaxTotalBytes: 128 * 1024 * 1024}, MediaRetentionLogger{})

	worker.runOnce(context.Background())

	want := []string{"expired", "expired", "expired", "budget", "budget"}
	if len(cleaner.calls) != len(want) {
		t.Fatalf("calls = %v, want %v", cleaner.calls, want)
	}
	for i := range want {
		if cleaner.calls[i] != want[i] {
			t.Fatalf("calls = %v, want %v", cleaner.calls, want)
		}
	}
	for _, budget := range cleaner.budgets {
		if budget != 128*1024*1024 {
			t.Fatalf("budget = %d, want %d", budget, 128*1024*1024)
		}
	}
}

func TestMediaRetentionCycleEnforcesTheBudgetWhenNothingExpired(t *testing.T) {
	cleaner := &scriptedMediaCleaner{overBudget: []int64{7}}
	worker := NewMediaRetentionWorker(cleaner, MediaRetentionConfig{BatchSize: 25}, MediaRetentionLogger{})

	worker.runOnce(context.Background())

	if len(cleaner.calls) != 2 || cleaner.calls[0] != "expired" || cleaner.calls[1] != "budget" {
		t.Fatalf("calls = %v, want [expired budget]", cleaner.calls)
	}
}

func TestMediaRetentionDefaultsAreThreeDaysAndTwoGiB(t *testing.T) {
	worker := NewMediaRetentionWorker(&scriptedMediaCleaner{}, MediaRetentionConfig{}, MediaRetentionLogger{})
	if worker.config.Retention != 3*24*time.Hour {
		t.Fatalf("retention = %s, want 72h", worker.config.Retention)
	}
	if worker.config.MaxTotalBytes != 2*1024*1024*1024 {
		t.Fatalf("budget = %d, want 2 GiB", worker.config.MaxTotalBytes)
	}
	if worker.config.Interval != 15*time.Minute {
		t.Fatalf("interval = %s, want 15m", worker.config.Interval)
	}
	if config := MediaRetentionConfigFromEnvironment(); config.Retention != 3*24*time.Hour || config.MaxTotalBytes != 2*1024*1024*1024 {
		t.Fatalf("environment defaults = %+v", config)
	}
}

func TestMediaRetentionConfigFromEnvironment(t *testing.T) {
	t.Setenv("ARGO_MESSAGE_MEDIA_RETENTION_DAYS", "45")
	t.Setenv("ARGO_MESSAGE_MEDIA_CLEANUP_SECONDS", "3600")
	t.Setenv("ARGO_MESSAGE_MEDIA_CLEANUP_TIMEOUT_SECONDS", "20")
	t.Setenv("ARGO_MESSAGE_MEDIA_CLEANUP_BATCH", "750")
	t.Setenv("ARGO_MESSAGE_MEDIA_MAX_TOTAL_BYTES", "536870912")
	config := MediaRetentionConfigFromEnvironment()
	if config.Retention != 45*24*time.Hour || config.Interval != time.Hour || config.Timeout != 20*time.Second || config.BatchSize != 750 || config.MaxTotalBytes != 512*1024*1024 {
		t.Fatalf("unexpected config: %+v", config)
	}
}

func TestMediaRetentionConfigRejectsUnsafeValues(t *testing.T) {
	t.Setenv("ARGO_MESSAGE_MEDIA_RETENTION_DAYS", "0")
	t.Setenv("ARGO_MESSAGE_MEDIA_CLEANUP_SECONDS", "10")
	t.Setenv("ARGO_MESSAGE_MEDIA_CLEANUP_BATCH", "99999")
	t.Setenv("ARGO_MESSAGE_MEDIA_MAX_TOTAL_BYTES", "1024")
	config := MediaRetentionConfigFromEnvironment()
	if config.Retention != defaultMediaRetention || config.Interval != defaultMediaCleanupInterval || config.BatchSize != defaultMediaCleanupBatch || config.MaxTotalBytes != defaultMediaMaxTotalBytes {
		t.Fatalf("unexpected fallback config: %+v", config)
	}
}
