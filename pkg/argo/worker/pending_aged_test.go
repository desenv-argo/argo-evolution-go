package argo_worker

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"
)

type fakePendingAgedReconciler struct {
	mu      sync.Mutex
	calls   int
	created int64
	err     error
	called  chan struct{}
}

type blockingPendingAgedReconciler struct{}

func (blockingPendingAgedReconciler) ReconcilePendingAged(ctx context.Context) (int64, error) {
	<-ctx.Done()
	return 0, ctx.Err()
}

func (f *fakePendingAgedReconciler) ReconcilePendingAged(context.Context) (int64, error) {
	f.mu.Lock()
	f.calls++
	f.mu.Unlock()
	if f.called != nil {
		select {
		case f.called <- struct{}{}:
		default:
		}
	}
	return f.created, f.err
}

func TestPendingAgedWorkerRunsImmediatelyAndStops(t *testing.T) {
	reconciler := &fakePendingAgedReconciler{created: 2, called: make(chan struct{}, 1)}
	worker := NewPendingAgedWorker(reconciler, PendingAgedConfig{Interval: time.Hour, Timeout: time.Second}, PendingAgedLogger{})
	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan struct{})
	go func() {
		worker.Run(ctx)
		close(done)
	}()

	select {
	case <-reconciler.called:
	case <-time.After(time.Second):
		t.Fatal("worker did not reconcile on startup")
	}
	cancel()
	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("worker did not stop after cancellation")
	}

	status := worker.Status()
	if status.Running || status.Runs != 1 || status.Created != 2 || status.LastCreated != 2 || status.LastSuccessAt.IsZero() {
		t.Fatalf("unexpected worker status: %+v", status)
	}
}

func TestPendingAgedWorkerKeepsFailureInStatus(t *testing.T) {
	reconciler := &fakePendingAgedReconciler{err: errors.New("database unavailable")}
	worker := NewPendingAgedWorker(reconciler, PendingAgedConfig{}, PendingAgedLogger{})

	worker.runOnce(context.Background())

	status := worker.Status()
	if status.Runs != 1 || status.Failures != 1 || status.LastError != "database unavailable" || !status.LastSuccessAt.IsZero() {
		t.Fatalf("unexpected worker status: %+v", status)
	}
}

func TestPendingAgedWorkerAppliesExecutionTimeout(t *testing.T) {
	worker := NewPendingAgedWorker(blockingPendingAgedReconciler{}, PendingAgedConfig{Timeout: 10 * time.Millisecond}, PendingAgedLogger{})

	worker.runOnce(context.Background())

	status := worker.Status()
	if status.Failures != 1 || status.LastError != context.DeadlineExceeded.Error() {
		t.Fatalf("unexpected timeout status: %+v", status)
	}
}

func TestPendingAgedConfigFromEnvironment(t *testing.T) {
	t.Setenv("ARGO_MESSAGE_PENDING_RECONCILE_SECONDS", "30")
	t.Setenv("ARGO_MESSAGE_PENDING_RECONCILE_TIMEOUT_SECONDS", "4")

	config := PendingAgedConfigFromEnvironment()
	if config.Interval != 30*time.Second || config.Timeout != 4*time.Second {
		t.Fatalf("unexpected config: %+v", config)
	}
}

func TestPendingAgedConfigRejectsValuesOutsideLimits(t *testing.T) {
	t.Setenv("ARGO_MESSAGE_PENDING_RECONCILE_SECONDS", "0")
	t.Setenv("ARGO_MESSAGE_PENDING_RECONCILE_TIMEOUT_SECONDS", "invalid")

	config := PendingAgedConfigFromEnvironment()
	if config.Interval != defaultReconcileInterval || config.Timeout != defaultReconcileTimeout {
		t.Fatalf("unexpected fallback config: %+v", config)
	}
}
