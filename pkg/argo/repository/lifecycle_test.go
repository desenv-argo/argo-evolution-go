package argo_repository

import (
	"testing"
	"time"

	argo_model "github.com/evolution-foundation/evolution-go/pkg/argo/model"
)

func TestAttemptLifecycleEventsForSuccessfulLegacySend(t *testing.T) {
	started := time.Date(2026, time.August, 18, 12, 0, 0, 0, time.UTC)
	completed := started.Add(125 * time.Millisecond)
	instanceID := "instance-1"
	attempt := &argo_model.MessageAttempt{
		ID: "attempt-1", ApplicationSlug: "legacy/unknown", InstanceID: &instanceID,
		CorrelationID: "correlation-1", ProviderMessageID: "message-1", Endpoint: "/send/text",
		Outcome: "succeeded", StartedAt: started, CompletedAt: completed,
	}

	events := attemptLifecycleEvents(attempt)
	wantStates := []string{"received", "validated", "accepted", "sent"}
	if len(events) != len(wantStates) {
		t.Fatalf("attemptLifecycleEvents() returned %d events, want %d", len(events), len(wantStates))
	}
	for index, state := range wantStates {
		if events[index].State != state || events[index].ApplicationSlug != "legacy/unknown" {
			t.Fatalf("event %d = %#v, want state %q and legacy identity", index, events[index], state)
		}
		if events[index].EventKey == "" || events[index].MessageType != "text" {
			t.Fatalf("event %d did not preserve correlation attributes: %#v", index, events[index])
		}
	}
}

func TestAttemptLifecycleEventsClassifiesFailure(t *testing.T) {
	now := time.Now().UTC()
	attempt := &argo_model.MessageAttempt{
		ID: "attempt-2", ApplicationSlug: "argo-app", Endpoint: "/send/media",
		Outcome: "failed", ErrorCode: "MEDIA_FAILED", ErrorDetail: "upload failed",
		StartedAt: now, CompletedAt: now.Add(time.Second),
	}
	events := attemptLifecycleEvents(attempt)
	if len(events) != 4 || events[3].State != "failed" {
		t.Fatalf("unexpected failure lifecycle: %#v", events)
	}
	if events[1].State != "validated" || events[2].State != "accepted" {
		t.Fatalf("post-validation failure must preserve accepted evidence: %#v", events)
	}
	if events[3].FailureCategory != "media" || events[3].FailureCode != "MEDIA_FAILED" {
		t.Fatalf("failure taxonomy not preserved: %#v", events[3])
	}
}

func TestAttemptLifecycleEventsDoesNotAcceptValidationFailure(t *testing.T) {
	now := time.Now().UTC()
	attempt := &argo_model.MessageAttempt{
		ID: "attempt-3", ApplicationSlug: "argo-app", Endpoint: "/send/text",
		Outcome: "failed", ErrorCode: "INVALID_RECIPIENT", StartedAt: now, CompletedAt: now,
	}
	events := attemptLifecycleEvents(attempt)
	if len(events) != 2 || events[0].State != "received" || events[1].State != "failed" {
		t.Fatalf("validation failure must not be counted as accepted: %#v", events)
	}
}
