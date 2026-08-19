package argo_worker

import (
	"context"
	"testing"
	"time"

	argo_model "github.com/evolution-foundation/evolution-go/pkg/argo/model"
)

type fakeUpstreamChecker struct{ called chan struct{} }

func (f fakeUpstreamChecker) Check(context.Context) *argo_model.UpstreamSnapshot {
	select {
	case f.called <- struct{}{}:
	default:
	}
	return &argo_model.UpstreamSnapshot{Status: "up_to_date", CheckedAt: time.Now()}
}

type fakeUpstreamRepository struct {
	saved chan *argo_model.UpstreamSnapshot
}

func (f fakeUpstreamRepository) SaveUpstreamSnapshot(_ context.Context, snapshot *argo_model.UpstreamSnapshot) error {
	f.saved <- snapshot
	return nil
}

func TestUpstreamMonitorRunsImmediatelyAndStops(t *testing.T) {
	called := make(chan struct{}, 1)
	saved := make(chan *argo_model.UpstreamSnapshot, 1)
	monitor := NewUpstreamMonitor(fakeUpstreamChecker{called}, fakeUpstreamRepository{saved}, UpstreamMonitorConfig{Interval: time.Hour, Timeout: time.Second}, UpstreamMonitorLogger{})
	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan struct{})
	go func() { monitor.Run(ctx); close(done) }()
	select {
	case snapshot := <-saved:
		if snapshot.Status != "up_to_date" {
			t.Fatalf("unexpected snapshot: %+v", snapshot)
		}
	case <-time.After(time.Second):
		t.Fatal("snapshot was not saved")
	}
	cancel()
	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("monitor did not stop")
	}
}

func TestUpstreamMonitorConfigFromEnvironment(t *testing.T) {
	t.Setenv("ARGO_UPSTREAM_CHECK_SECONDS", "3600")
	t.Setenv("ARGO_UPSTREAM_CHECK_TIMEOUT_SECONDS", "20")
	config := UpstreamMonitorConfigFromEnvironment()
	if config.Interval != time.Hour || config.Timeout != 20*time.Second {
		t.Fatalf("unexpected config: %+v", config)
	}
}
