package analytics_handler

import (
	"net/http"
	"strconv"
	"strings"
	"time"

	analytics_model "github.com/evolution-foundation/evolution-go/pkg/analytics/model"
	analytics_repository "github.com/evolution-foundation/evolution-go/pkg/analytics/repository"
	analytics_settings "github.com/evolution-foundation/evolution-go/pkg/analytics/settings"
	"github.com/gin-gonic/gin"
)

type AnalyticsHandler interface {
	Dashboard(ctx *gin.Context)
	Conversations(ctx *gin.Context)
	Messages(ctx *gin.Context)
	Settings(ctx *gin.Context)
	UpdateSettings(ctx *gin.Context)
}

type analyticsHandler struct {
	repository  analytics_repository.AnalyticsRepository
	captureGate *analytics_settings.CaptureGate
	now         func() time.Time
}

// Dashboard returns operational instance and message metrics for the Manager.
// @Summary Manager dashboard metrics
// @Tags Analytics
// @Produce json
// @Param from query string false "RFC3339 timestamp or YYYY-MM-DD"
// @Param to query string false "RFC3339 timestamp or YYYY-MM-DD"
// @Param instanceId query string false "Instance UUID"
// @Success 200 {object} gin.H
// @Router /analytics/dashboard [get]
func (h *analyticsHandler) Dashboard(ctx *gin.Context) {
	filters, err := h.parseFilters(ctx)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	result, err := h.repository.Dashboard(ctx.Request.Context(), filters)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to load dashboard metrics"})
		return
	}
	result.MessageCaptureEnabled = h.captureGate.Enabled()
	ctx.JSON(http.StatusOK, gin.H{"message": "success", "data": result})
}

func (h *analyticsHandler) Settings(ctx *gin.Context) {
	ctx.JSON(http.StatusOK, gin.H{"message": "success", "data": h.captureGate.Settings()})
}

func (h *analyticsHandler) UpdateSettings(ctx *gin.Context) {
	var input struct {
		MessageCaptureEnabled *bool `json:"message_capture_enabled"`
	}
	if err := ctx.ShouldBindJSON(&input); err != nil || input.MessageCaptureEnabled == nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "message_capture_enabled is required"})
		return
	}
	settings, err := h.captureGate.Update(*input.MessageCaptureEnabled)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update analytics settings"})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"message": "success", "data": settings})
}

// Conversations returns one row per WhatsApp chat, ordered by last activity.
// @Summary List structured conversations
// @Tags Analytics
// @Produce json
// @Success 200 {object} gin.H
// @Router /analytics/conversations [get]
func (h *analyticsHandler) Conversations(ctx *gin.Context) {
	filters, err := h.parseFilters(ctx)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	result, err := h.repository.ListConversations(ctx.Request.Context(), filters)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to load conversations"})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"message": "success", "data": result})
}

// Messages returns the timeline for an instance/chat pair.
// @Summary List messages from a conversation
// @Tags Analytics
// @Produce json
// @Param instanceId query string true "Instance UUID"
// @Param chatJid query string true "WhatsApp chat JID"
// @Success 200 {object} gin.H
// @Router /analytics/messages [get]
func (h *analyticsHandler) Messages(ctx *gin.Context) {
	filters, err := h.parseFilters(ctx)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	chatJID := strings.TrimSpace(ctx.Query("chatJid"))
	if filters.InstanceID == "" || chatJID == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "instanceId and chatJid are required"})
		return
	}

	result, err := h.repository.ListMessages(ctx.Request.Context(), filters, chatJID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to load messages"})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"message": "success", "data": result})
}

func (h *analyticsHandler) parseFilters(ctx *gin.Context) (analytics_model.Filters, error) {
	now := h.now().UTC()
	to, err := parseDateTime(ctx.Query("to"), now, true)
	if err != nil {
		return analytics_model.Filters{}, err
	}
	from, err := parseDateTime(ctx.Query("from"), to.AddDate(0, 0, -7), false)
	if err != nil {
		return analytics_model.Filters{}, err
	}
	if from.After(to) {
		return analytics_model.Filters{}, &filterError{message: "from must be before to"}
	}
	if to.Sub(from) > 366*24*time.Hour {
		return analytics_model.Filters{}, &filterError{message: "period cannot exceed 366 days"}
	}

	limit := 50
	if rawLimit := ctx.Query("limit"); rawLimit != "" {
		parsed, parseErr := strconv.Atoi(rawLimit)
		if parseErr != nil || parsed <= 0 {
			return analytics_model.Filters{}, &filterError{message: "limit must be a positive integer"}
		}
		limit = parsed
	}

	var before *time.Time
	if rawBefore := ctx.Query("before"); rawBefore != "" {
		parsed, parseErr := time.Parse(time.RFC3339Nano, rawBefore)
		if parseErr != nil {
			return analytics_model.Filters{}, &filterError{message: "before must use RFC3339 format"}
		}
		parsed = parsed.UTC()
		before = &parsed
	}

	var isGroup *bool
	if rawGroup := ctx.Query("isGroup"); rawGroup != "" {
		parsed, parseErr := strconv.ParseBool(rawGroup)
		if parseErr != nil {
			return analytics_model.Filters{}, &filterError{message: "isGroup must be true or false"}
		}
		isGroup = &parsed
	}

	direction := strings.ToLower(strings.TrimSpace(ctx.Query("direction")))
	if direction != "" && direction != "inbound" && direction != "outbound" {
		return analytics_model.Filters{}, &filterError{message: "direction must be inbound or outbound"}
	}

	return analytics_model.Filters{
		From:        from,
		To:          to,
		InstanceID:  strings.TrimSpace(ctx.Query("instanceId")),
		Search:      strings.TrimSpace(ctx.Query("search")),
		Direction:   direction,
		MessageType: strings.ToLower(strings.TrimSpace(ctx.Query("messageType"))),
		Status:      strings.TrimSpace(ctx.Query("status")),
		IsGroup:     isGroup,
		Before:      before,
		Limit:       limit,
	}, nil
}

func parseDateTime(raw string, fallback time.Time, endOfDay bool) (time.Time, error) {
	if raw == "" {
		return fallback.UTC(), nil
	}
	if parsed, err := time.Parse(time.RFC3339, raw); err == nil {
		return parsed.UTC(), nil
	}
	parsed, err := time.Parse("2006-01-02", raw)
	if err != nil {
		return time.Time{}, &filterError{message: "dates must use RFC3339 or YYYY-MM-DD format"}
	}
	if endOfDay {
		parsed = parsed.Add(24*time.Hour - time.Nanosecond)
	}
	return parsed.UTC(), nil
}

type filterError struct {
	message string
}

func (e *filterError) Error() string { return e.message }

func NewAnalyticsHandler(repository analytics_repository.AnalyticsRepository, captureGate *analytics_settings.CaptureGate) AnalyticsHandler {
	return &analyticsHandler{repository: repository, captureGate: captureGate, now: time.Now}
}
