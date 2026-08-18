package argo_service

import (
	"testing"
	"time"

	argo_model "github.com/evolution-foundation/evolution-go/pkg/argo/model"
)

func TestValidateInput(t *testing.T) {
	tests := []struct {
		name    string
		input   ApplicationInput
		wantErr bool
	}{
		{name: "valid", input: ApplicationInput{Slug: "argo-erp", Name: "Argo ERP", HealthURL: "https://erp.argo.app.br/health"}},
		{name: "invalid slug", input: ApplicationInput{Slug: "ERP!", Name: "ERP"}, wantErr: true},
		{name: "missing name", input: ApplicationInput{Slug: "argo-erp"}, wantErr: true},
		{name: "invalid health URL", input: ApplicationInput{Slug: "argo-erp", Name: "ERP", HealthURL: "javascript:alert(1)"}, wantErr: true},
		{name: "negative heartbeat", input: ApplicationInput{Slug: "argo-erp", Name: "ERP", ExpectedHeartbeatSeconds: -1}, wantErr: true},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := validateInput(test.input, true)
			if (err != nil) != test.wantErr {
				t.Fatalf("validateInput() error = %v, wantErr %v", err, test.wantErr)
			}
		})
	}
}

func TestCredentialHashIsStableAndDoesNotExposeCredential(t *testing.T) {
	credential, hash, err := newCredential()
	if err != nil {
		t.Fatal(err)
	}
	if credential == "" || hash == "" {
		t.Fatal("credential and hash must be generated")
	}
	if credential == hash {
		t.Fatal("stored hash must not equal the plaintext credential")
	}
	if got := hashCredential(credential); got != hash {
		t.Fatalf("hashCredential() = %q, want %q", got, hash)
	}
}

func TestNormalizeHeartbeat(t *testing.T) {
	if got := normalizeHeartbeat(0); got != 300 {
		t.Fatalf("normalizeHeartbeat(0) = %d", got)
	}
	if got := normalizeHeartbeat(5); got != 30 {
		t.Fatalf("normalizeHeartbeat(5) = %d", got)
	}
	if got := normalizeHeartbeat(100000); got != 86400 {
		t.Fatalf("normalizeHeartbeat(100000) = %d", got)
	}
}

func TestLifecycleBackfillOptionsAreBoundedAndSafeByDefault(t *testing.T) {
	from := time.Date(2026, time.August, 1, 0, 0, 0, 0, time.UTC)
	options, err := lifecycleBackfillOptions(LifecycleBackfillInput{From: from, To: from.Add(24 * time.Hour)})
	if err != nil {
		t.Fatal(err)
	}
	if options.Limit != 1000 || options.Execute {
		t.Fatalf("unexpected safe defaults: %#v", options)
	}
	if _, err := lifecycleBackfillOptions(LifecycleBackfillInput{From: from, To: from.Add(32 * 24 * time.Hour)}); err == nil {
		t.Fatal("expected periods over 31 days to be rejected")
	}
	if _, err := lifecycleBackfillOptions(LifecycleBackfillInput{From: from, To: from.Add(time.Hour), Limit: 5001}); err == nil {
		t.Fatal("expected limits over 5000 to be rejected")
	}
}

func TestMessageMediaSafetyNormalization(t *testing.T) {
	if got := sanitizeMediaFileName(`../financeiro\boleto.pdf`, "message-1"); got != ".._financeiro_boleto.pdf" {
		t.Fatalf("unexpected sanitized filename: %q", got)
	}
	if got := sanitizeMediaFileName("", "message-1"); got != "message-1" {
		t.Fatalf("unexpected fallback filename: %q", got)
	}
	if got := normalizeMediaType("application/pdf; charset=binary"); got != "application/pdf" {
		t.Fatalf("unexpected normalized media type: %q", got)
	}
	if got := normalizeMediaType("application/pdf\r\nX-Test: unsafe"); got != "application/octet-stream" {
		t.Fatalf("unsafe media type was accepted: %q", got)
	}
}

func TestMessageMediaMaxBytesUsesSafeBounds(t *testing.T) {
	t.Setenv("ARGO_MESSAGE_MEDIA_MAX_BYTES", "1048576")
	if got := messageMediaMaxBytesFromEnvironment(); got != 1048576 {
		t.Fatalf("unexpected configured limit: %d", got)
	}
	t.Setenv("ARGO_MESSAGE_MEDIA_MAX_BYTES", "999999999")
	if got := messageMediaMaxBytesFromEnvironment(); got != 25*1024*1024 {
		t.Fatalf("invalid limit must use default: %d", got)
	}
}

func TestGatewayOperationalState(t *testing.T) {
	tests := []struct {
		name      string
		attempts  argo_model.AttemptSummary
		lifecycle argo_model.LifecycleSummary
		want      string
	}{
		{name: "healthy", attempts: argo_model.AttemptSummary{Total: 100, Failed: 1}, want: "healthy"},
		{name: "degraded by failures", attempts: argo_model.AttemptSummary{Total: 100, Failed: 4}, want: "degraded"},
		{name: "degraded by legacy traffic", attempts: argo_model.AttemptSummary{Total: 100, UnverifiedIdentity: 1}, want: "degraded"},
		{name: "unhealthy by failures", attempts: argo_model.AttemptSummary{Total: 100, Failed: 10}, want: "unhealthy"},
		{name: "unhealthy by aged messages", attempts: argo_model.AttemptSummary{Total: 100}, lifecycle: argo_model.LifecycleSummary{PendingAged: 1}, want: "unhealthy"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got, _ := gatewayOperationalState(&test.attempts, &test.lifecycle)
			if got != test.want {
				t.Fatalf("state = %q, want %q", got, test.want)
			}
		})
	}
}

func TestApplicationHealth(t *testing.T) {
	now := time.Date(2026, time.August, 18, 18, 0, 0, 0, time.UTC)
	recent := now.Add(-45 * time.Second)
	stale := now.Add(-121 * time.Second)
	tests := []struct {
		name        string
		application argo_model.Application
		wantState   string
		wantAge     int64
	}{
		{name: "disabled", application: argo_model.Application{Active: false}, wantState: "disabled"},
		{name: "unknown", application: argo_model.Application{Active: true}, wantState: "unknown"},
		{name: "healthy", application: argo_model.Application{Active: true, ExpectedHeartbeatSeconds: 60, LastHeartbeatAt: &recent, LastHeartbeatStatus: "healthy"}, wantState: "healthy", wantAge: 45},
		{name: "reported degradation", application: argo_model.Application{Active: true, ExpectedHeartbeatSeconds: 60, LastHeartbeatAt: &recent, LastHeartbeatStatus: "degraded"}, wantState: "degraded", wantAge: 45},
		{name: "stale", application: argo_model.Application{Active: true, ExpectedHeartbeatSeconds: 60, LastHeartbeatAt: &stale, LastHeartbeatStatus: "healthy"}, wantState: "offline", wantAge: 121},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			state, age := applicationHealth(&test.application, now)
			if state != test.wantState || age != test.wantAge {
				t.Fatalf("applicationHealth() = (%q, %d), want (%q, %d)", state, age, test.wantState, test.wantAge)
			}
		})
	}
}

func TestCleanValue(t *testing.T) {
	if got := cleanValue("  version\r\nwith-noise  ", 12); got != "version  wit" {
		t.Fatalf("cleanValue() = %q", got)
	}
}

func TestSummarizeLifecycleBuildsFunnelAndPercentiles(t *testing.T) {
	start := time.Date(2026, time.August, 18, 12, 0, 0, 0, time.UTC)
	attemptOne := "attempt-1"
	attemptTwo := "attempt-2"
	events := []argo_model.MessageLifecycleEvent{
		{AttemptID: &attemptOne, State: "received", OccurredAt: start},
		{AttemptID: &attemptOne, State: "validated", OccurredAt: start.Add(10 * time.Millisecond)},
		{AttemptID: &attemptOne, State: "accepted", OccurredAt: start.Add(20 * time.Millisecond)},
		{AttemptID: &attemptOne, State: "sent", OccurredAt: start.Add(100 * time.Millisecond)},
		{AttemptID: &attemptOne, State: "delivered", OccurredAt: start.Add(1100 * time.Millisecond)},
		{AttemptID: &attemptOne, State: "read", OccurredAt: start.Add(3100 * time.Millisecond)},
		{AttemptID: &attemptTwo, State: "received", OccurredAt: start},
		{AttemptID: &attemptTwo, State: "failed", FailureCategory: "validation", FailureCode: "INVALID_RECIPIENT", OccurredAt: start.Add(50 * time.Millisecond)},
	}

	summary := summarizeLifecycle(events, start, start.Add(time.Hour))
	if summary.Received != 2 || summary.Accepted != 1 || summary.Sent != 1 || summary.Delivered != 1 || summary.Read != 1 || summary.Failed != 1 {
		t.Fatalf("unexpected funnel summary: %#v", summary)
	}
	if summary.AcceptanceRate != 50 || summary.SendRate != 100 || summary.DeliveryRate != 100 || summary.ReadRate != 100 {
		t.Fatalf("unexpected rates: %#v", summary)
	}
	if summary.SendLatency.P50 != 100 || summary.DeliveryLatency.P95 != 1000 || summary.ReadLatency.P99 != 2000 {
		t.Fatalf("unexpected latency percentiles: %#v", summary)
	}
	if len(summary.Failures) != 1 || summary.Failures[0].Code != "INVALID_RECIPIENT" {
		t.Fatalf("unexpected failure breakdown: %#v", summary.Failures)
	}
}
