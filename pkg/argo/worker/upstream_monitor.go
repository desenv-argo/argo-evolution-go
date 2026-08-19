package argo_worker

import (
	"context"
	"time"

	argo_model "github.com/evolution-foundation/evolution-go/pkg/argo/model"
)

const defaultUpstreamCheckInterval = 6 * time.Hour
const defaultUpstreamCheckTimeout = 30 * time.Second

type UpstreamChecker interface {
	Check(ctx context.Context) *argo_model.UpstreamSnapshot
}
type UpstreamSnapshotRepository interface {
	SaveUpstreamSnapshot(ctx context.Context, snapshot *argo_model.UpstreamSnapshot) error
}

type UpstreamMonitorConfig struct {
	Interval time.Duration
	Timeout  time.Duration
}
type UpstreamMonitorLogger struct {
	Info  func(format string, args ...interface{})
	Error func(format string, args ...interface{})
}

type UpstreamMonitor struct {
	checker    UpstreamChecker
	repository UpstreamSnapshotRepository
	config     UpstreamMonitorConfig
	logger     UpstreamMonitorLogger
}

func NewUpstreamMonitor(checker UpstreamChecker, repository UpstreamSnapshotRepository, config UpstreamMonitorConfig, logger UpstreamMonitorLogger) *UpstreamMonitor {
	if config.Interval <= 0 {
		config.Interval = defaultUpstreamCheckInterval
	}
	if config.Timeout <= 0 {
		config.Timeout = defaultUpstreamCheckTimeout
	}
	return &UpstreamMonitor{checker: checker, repository: repository, config: config, logger: logger}
}

func UpstreamMonitorConfigFromEnvironment() UpstreamMonitorConfig {
	return UpstreamMonitorConfig{
		Interval: durationFromSeconds("ARGO_UPSTREAM_CHECK_SECONDS", defaultUpstreamCheckInterval, 5*time.Minute, 7*24*time.Hour),
		Timeout:  durationFromSeconds("ARGO_UPSTREAM_CHECK_TIMEOUT_SECONDS", defaultUpstreamCheckTimeout, 5*time.Second, 5*time.Minute),
	}
}

func (w *UpstreamMonitor) Run(ctx context.Context) {
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

func (w *UpstreamMonitor) runOnce(parent context.Context) {
	startedAt := time.Now().UTC()
	checkCtx, cancelCheck := context.WithTimeout(parent, w.config.Timeout)
	snapshot := w.checker.Check(checkCtx)
	cancelCheck()
	persistCtx, cancelPersist := context.WithTimeout(parent, w.config.Timeout)
	err := w.repository.SaveUpstreamSnapshot(persistCtx, snapshot)
	cancelPersist()
	if err != nil {
		if w.logger.Error != nil {
			w.logger.Error("[ARGO_UPSTREAM] failed to persist check: %v", err)
		}
		return
	}
	if snapshot.Status == "unavailable" {
		if w.logger.Error != nil {
			w.logger.Error("[ARGO_UPSTREAM] check unavailable: %s", snapshot.Error)
		}
		return
	}
	if w.logger.Info != nil {
		w.logger.Info("[ARGO_UPSTREAM] check completed: status=%s version=%s behind=%d duration=%s", snapshot.Status, snapshot.LatestVersion, snapshot.BehindBy, time.Since(startedAt))
	}
}
