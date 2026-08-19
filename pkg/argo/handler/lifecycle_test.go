package argo_handler

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	argo_model "github.com/evolution-foundation/evolution-go/pkg/argo/model"
	argo_service "github.com/evolution-foundation/evolution-go/pkg/argo/service"
	"github.com/gin-gonic/gin"
)

type lifecycleFeedServiceStub struct {
	argo_service.Service
	identity      argo_service.Identity
	events        []argo_model.MessageLifecycleEvent
	applicationID string
	limit         int
}

func (s *lifecycleFeedServiceStub) Identify(context.Context, string, string) argo_service.Identity {
	return s.identity
}

func (s *lifecycleFeedServiceStub) ListLifecycleFeed(_ context.Context, applicationID string, _ argo_model.LifecycleFeedCursor, limit int) ([]argo_model.MessageLifecycleEvent, error) {
	s.applicationID = applicationID
	s.limit = limit
	return s.events, nil
}

func TestLifecycleFiltersContract(t *testing.T) {
	gin.SetMode(gin.TestMode)
	now := time.Date(2026, time.August, 18, 12, 0, 0, 0, time.UTC)
	h := &handler{now: func() time.Time { return now }}
	ctx, _ := gin.CreateTestContext(httptest.NewRecorder())
	ctx.Request = httptest.NewRequest("GET", "/argo/v1/messages/lifecycle?application=argo-app&instanceId=instance-1&type=text&state=delivered&providerMessageId=message-1&correlationId=correlation-1&idempotencyKey=command-1&limit=250", nil)

	filters, err := h.lifecycleFilters(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if filters.ApplicationSlug != "argo-app" || filters.InstanceID != "instance-1" || filters.MessageType != "text" || filters.State != "delivered" {
		t.Fatalf("unexpected lifecycle filters: %#v", filters)
	}
	if filters.ProviderMessageID != "message-1" || filters.CorrelationID != "correlation-1" || filters.IdempotencyKey != "command-1" || filters.Limit != 250 {
		t.Fatalf("correlation filters were not preserved: %#v", filters)
	}
}

func TestLifecycleFeedCursorRoundTrip(t *testing.T) {
	want := argo_model.LifecycleFeedCursor{
		CreatedAt: time.Date(2026, time.August, 19, 12, 34, 56, 789, time.UTC),
		ID:        "c578a8ef-0adb-49be-a01d-e829c7746a25",
	}
	encoded := encodeLifecycleFeedCursor(want)
	got, err := decodeLifecycleFeedCursor(encoded)
	if err != nil {
		t.Fatal(err)
	}
	if !got.CreatedAt.Equal(want.CreatedAt) || got.ID != want.ID {
		t.Fatalf("decoded cursor = %#v, want %#v", got, want)
	}
}

func TestLifecycleFeedCursorRejectsMalformedValue(t *testing.T) {
	if _, err := decodeLifecycleFeedCursor("not-base64!"); err == nil {
		t.Fatal("expected malformed cursor to be rejected")
	}
}

func TestLifecycleFeedItemUsesApplicationContract(t *testing.T) {
	instanceID := "f5821036-1cb2-40fb-8de6-b88e7ed1af2f"
	now := time.Date(2026, time.August, 19, 13, 0, 0, 0, time.UTC)
	item := lifecycleFeedItem(argo_model.MessageLifecycleEvent{
		ID: "event-1", ApplicationSlug: "argo-erp", InstanceID: &instanceID,
		CorrelationID: "correlation-1", ProviderMessageID: "provider-1",
		State: "delivered", FailureDetail: "raw recipient detail",
		OccurredAt: now.Add(-time.Second), CreatedAt: now,
	})
	payload, err := json.Marshal(item)
	if err != nil {
		t.Fatal(err)
	}
	value := string(payload)
	for _, field := range []string{"eventId", "application", "instanceId", "correlationId", "providerMessageId", "occurredAt", "createdAt"} {
		if !strings.Contains(value, "\""+field+"\"") {
			t.Fatalf("feed payload %s does not contain %s", value, field)
		}
	}
	if strings.Contains(value, "failureDetail") {
		t.Fatalf("application feed must not expose raw provider detail: %s", value)
	}
	if strings.Contains(value, "raw recipient detail") {
		t.Fatalf("application feed leaked raw provider detail: %s", value)
	}
}

func TestApplicationLifecycleFeedRequiresVerifiedApplication(t *testing.T) {
	gin.SetMode(gin.TestMode)
	service := &lifecycleFeedServiceStub{identity: argo_service.Identity{ApplicationSlug: "argo-erp"}}
	h := &handler{service: service, now: time.Now}
	response := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(response)
	ctx.Request = httptest.NewRequest(http.MethodGet, "/argo/v1/messages/lifecycle/feed", nil)

	h.ApplicationLifecycleFeed(ctx)

	if response.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusUnauthorized)
	}
}

func TestApplicationLifecycleFeedForcesCredentialApplication(t *testing.T) {
	gin.SetMode(gin.TestMode)
	applicationID := "4c91f3f8-28b3-4289-af0b-4849bcc4c8fa"
	now := time.Date(2026, time.August, 19, 15, 0, 0, 0, time.UTC)
	service := &lifecycleFeedServiceStub{
		identity: argo_service.Identity{
			ApplicationID: &applicationID, ApplicationSlug: "argo-erp", Verified: true,
		},
		events: []argo_model.MessageLifecycleEvent{{
			ID: "event-1", ApplicationSlug: "argo-erp", State: "sent",
			OccurredAt: now, CreatedAt: now,
		}},
	}
	h := &handler{service: service, now: time.Now}
	response := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(response)
	ctx.Request = httptest.NewRequest(http.MethodGet, "/argo/v1/messages/lifecycle/feed?application=another-app&limit=25", nil)
	ctx.Request.Header.Set("X-Argo-Application-Id", "argo-erp")
	ctx.Request.Header.Set("X-Argo-Application-Key", "credential")

	h.ApplicationLifecycleFeed(ctx)

	if response.Code != http.StatusOK {
		t.Fatalf("status = %d, body = %s", response.Code, response.Body.String())
	}
	if service.applicationID != applicationID || service.limit != 26 {
		t.Fatalf("feed scope = %q limit %d, want %q limit 26", service.applicationID, service.limit, applicationID)
	}
	var page argo_model.LifecycleFeedPage
	if err := json.Unmarshal(response.Body.Bytes(), &page); err != nil {
		t.Fatal(err)
	}
	if len(page.Data) != 1 || page.Data[0].Application != "argo-erp" || page.NextCursor == "" {
		t.Fatalf("unexpected feed page: %#v", page)
	}
}

func TestLifecycleFiltersRejectUnknownState(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := &handler{now: time.Now}
	ctx, _ := gin.CreateTestContext(httptest.NewRecorder())
	ctx.Request = httptest.NewRequest("GET", "/argo/v1/messages/lifecycle?state=deleted", nil)
	if _, err := h.lifecycleFilters(ctx); err == nil {
		t.Fatal("expected invalid lifecycle state to be rejected")
	}
}
