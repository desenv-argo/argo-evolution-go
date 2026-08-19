package argo_repository

import (
	"context"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	argo_model "github.com/evolution-foundation/evolution-go/pkg/argo/model"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
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

func TestBackfillLifecycleDryRunContract(t *testing.T) {
	sqlDB, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer sqlDB.Close()
	db, err := gorm.Open(postgres.New(postgres.Config{Conn: sqlDB}), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	start := time.Date(2026, time.August, 1, 12, 0, 0, 0, time.UTC)
	instanceID := "4d7ec72c-9fc0-4af9-aad8-573f4d0a89db"
	deliveredAt, readAt := start.Add(time.Second), start.Add(2*time.Second)
	mock.ExpectQuery(`(?s)SELECT \* FROM "argo_message_attempts".*started_at >= \$1.*started_at <= \$2.*ORDER BY started_at ASC, id ASC LIMIT \$3`).
		WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), 100).
		WillReturnRows(sqlmock.NewRows([]string{"id", "application_slug", "instance_id", "correlation_id", "provider_message_id", "endpoint", "method", "http_status", "outcome", "started_at", "completed_at"}).
			AddRow("2cb567f7-474d-47e3-b9b0-686d787b995e", "legacy/unknown", instanceID, "correlation-1", "message-1", "/send/text", "POST", 200, "succeeded", start, start.Add(100*time.Millisecond)))
	mock.ExpectQuery(`(?s)SELECT \* FROM "messages".*sent_at >= \$1.*sent_at <= \$2.*delivered_at IS NOT NULL OR read_at IS NOT NULL.*ORDER BY sent_at ASC, id ASC LIMIT \$5`).
		WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), "outbound", true, 100).
		WillReturnRows(sqlmock.NewRows([]string{"instance_id", "message_id", "direction", "from_me", "message_type", "sent_at", "delivered_at", "read_at"}).
			AddRow(instanceID, "message-1", "outbound", true, "text", start, deliveredAt, readAt))
	mock.ExpectQuery(`(?s)SELECT "event_key" FROM "argo_message_lifecycle_events" WHERE event_key IN`).
		WillReturnRows(sqlmock.NewRows([]string{"event_key"}))
	repository := NewRepository(db)
	options := argo_model.LifecycleBackfillOptions{From: start.Add(-time.Minute), To: start.Add(time.Minute), Limit: 100}

	dryRun, err := repository.BackfillLifecycle(context.Background(), options)
	if err != nil {
		t.Fatal(err)
	}
	if dryRun.CandidateEvents != 6 || dryRun.PendingEvents != 6 || dryRun.EventsCreated != 0 || dryRun.AttemptsScanned != 1 || dryRun.MessagesScanned != 1 {
		t.Fatalf("unexpected dry-run report: %#v", dryRun)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
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

func TestListLifecycleFeedScopesApplicationAndUsesStableCursor(t *testing.T) {
	sqlDB, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer sqlDB.Close()
	db, err := gorm.Open(postgres.New(postgres.Config{Conn: sqlDB}), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	createdAt := time.Date(2026, time.August, 19, 14, 0, 0, 0, time.UTC)
	mock.ExpectQuery(`(?s)SELECT \* FROM "argo_message_lifecycle_events".*application_id = \$1 AND identity_verified = TRUE.*created_at > \$2.*created_at = \$3 AND id > \$4.*ORDER BY created_at ASC, id ASC LIMIT \$5`).
		WithArgs("app-id", createdAt, createdAt, "event-1", 501).
		WillReturnRows(sqlmock.NewRows([]string{"id", "application_id", "application_slug", "state", "occurred_at", "created_at"}).
			AddRow("event-2", "app-id", "argo-erp", "sent", createdAt, createdAt.Add(time.Second)))

	repository := NewRepository(db)
	events, err := repository.ListLifecycleFeed(context.Background(), "app-id", argo_model.LifecycleFeedCursor{
		CreatedAt: createdAt,
		ID:        "event-1",
	}, 501)
	if err != nil {
		t.Fatal(err)
	}
	if len(events) != 1 || events[0].ID != "event-2" {
		t.Fatalf("unexpected feed events: %#v", events)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}
