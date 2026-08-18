package argo_handler

import (
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
)

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

func TestLifecycleFiltersRejectUnknownState(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := &handler{now: time.Now}
	ctx, _ := gin.CreateTestContext(httptest.NewRecorder())
	ctx.Request = httptest.NewRequest("GET", "/argo/v1/messages/lifecycle?state=deleted", nil)
	if _, err := h.lifecycleFilters(ctx); err == nil {
		t.Fatal("expected invalid lifecycle state to be rejected")
	}
}
