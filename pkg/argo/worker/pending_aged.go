package argo_worker

import (
	"context"
	"os"
	"strconv"
	"sync"
	"time"
)

const (
	defaultReconcileInterval = time.Minute
	defaultReconcileTimeout  = 10 * time.Second
)

type PendingAgedReconciler interface {
	ReconcilePendingAged(ctx context.Context) (int64, error)
}

type PendingAgedConfig struct {
	Interval time.Duration
	Timeout  time.Duration
}

type PendingAgedLogger struct {
	Info  func(format string, args ...interface{})
	Error func(format string, args ...interface{})
}

type PendingAgedStatus struct {
	Running         bool
	Runs            int64
	Failures        int64
	Created         int64
	LastCreated     int64
	LastStartedAt   time.Time
	LastCompletedAt time.Time
	LastSuccessAt   time.Time
	LastError       string
}

type PendingAgedWorker struct {
	reconciler PendingAgedReconciler
	config     PendingAgedConfig
	logger     PendingAgedLogger
	mu         sync.RWMutex
	status     PendingAgedStatus
}

func NewPendingAgedWorker(reconciler PendingAgedReconciler, config PendingAgedConfig, logger PendingAgedLogger) *PendingAgedWorker {
	if config.Interval <= 0 {
		config.Interval = defaultReconcileInterval
	}
	if config.Timeout <= 0 {
		config.Timeout = defaultReconcileTimeout
	}
	return &PendingAgedWorker{reconciler: reconciler, config: config, logger: logger}
}

func PendingAgedConfigFromEnvironment() PendingAgedConfig {
	return PendingAgedConfig{
		Interval: durationFromSeconds("ARGO_MESSAGE_PENDING_RECONCILE_SECONDS", defaultReconcileInterval, time.Second, time.Hour),
		Timeout:  durationFromSeconds("ARGO_MESSAGE_PENDING_RECONCILE_TIMEOUT_SECONDS", defaultReconcileTimeout, time.Second, 5*time.Minute),
	}
}

func (w *PendingAgedWorker) Run(ctx context.Context) {
	w.mu.Lock()
	w.status.Running = true
	w.mu.Unlock()
	defer func() {
		w.mu.Lock()
		w.status.Running = false
		w.mu.Unlock()
	}()

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

func (w *PendingAgedWorker) Status() PendingAgedStatus {
	w.mu.RLock()
	defer w.mu.RUnlock()
	return w.status
}

func (w *PendingAgedWorker) runOnce(parent context.Context) {
	startedAt := time.Now().UTC()
	w.mu.Lock()
	w.status.Runs++
	w.status.LastStartedAt = startedAt
	w.mu.Unlock()

	ctx, cancel := context.WithTimeout(parent, w.config.Timeout)
	created, err := w.reconciler.ReconcilePendingAged(ctx)
	cancel()
	completedAt := time.Now().UTC()

	w.mu.Lock()
	w.status.LastCompletedAt = completedAt
	if err != nil {
		w.status.Failures++
		w.status.LastError = err.Error()
	} else {
		w.status.Created += created
		w.status.LastCreated = created
		w.status.LastSuccessAt = completedAt
		w.status.LastError = ""
	}
	w.mu.Unlock()

	if err != nil {
		if w.logger.Error != nil {
			w.logger.Error("[ARGO_LIFECYCLE] pending_aged reconciliation failed after %s: %v", completedAt.Sub(startedAt), err)
		}
		return
	}
	if created > 0 && w.logger.Info != nil {
		w.logger.Info("[ARGO_LIFECYCLE] pending_aged reconciliation completed: created=%d duration=%s", created, completedAt.Sub(startedAt))
	}
}

func durationFromSeconds(name string, fallback, minimum, maximum time.Duration) time.Duration {
	seconds, err := strconv.Atoi(os.Getenv(name))
	if err != nil {
		return fallback
	}
	value := time.Duration(seconds) * time.Second
	if value < minimum || value > maximum {
		return fallback
	}
	return value
}
