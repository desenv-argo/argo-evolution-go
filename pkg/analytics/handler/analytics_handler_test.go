package analytics_handler

import (
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
)

func TestParseFiltersUsesSevenDayDefault(t *testing.T) {
	gin.SetMode(gin.TestMode)
	now := time.Date(2026, time.August, 17, 15, 0, 0, 0, time.UTC)
	handler := &analyticsHandler{now: func() time.Time { return now }}
	ctx, _ := gin.CreateTestContext(httptest.NewRecorder())
	ctx.Request = httptest.NewRequest("GET", "/analytics/dashboard", nil)

	filters, err := handler.parseFilters(ctx)
	if err != nil {
		t.Fatalf("parse filters: %v", err)
	}
	if !filters.To.Equal(now) || !filters.From.Equal(now.AddDate(0, 0, -7)) {
		t.Fatalf("unexpected default period: %s - %s", filters.From, filters.To)
	}
}

func TestParseFiltersRejectsInvalidDirection(t *testing.T) {
	gin.SetMode(gin.TestMode)
	handler := &analyticsHandler{now: time.Now}
	ctx, _ := gin.CreateTestContext(httptest.NewRecorder())
	ctx.Request = httptest.NewRequest("GET", "/analytics/conversations?direction=sideways", nil)

	if _, err := handler.parseFilters(ctx); err == nil {
		t.Fatal("expected invalid direction to be rejected")
	}
}
