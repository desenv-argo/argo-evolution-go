package argo_handler

import (
	"encoding/base64"
	"errors"
	"net/http"
	"strconv"
	"strings"
	"time"

	argo_model "github.com/evolution-foundation/evolution-go/pkg/argo/model"
	argo_service "github.com/evolution-foundation/evolution-go/pkg/argo/service"
	"github.com/gin-gonic/gin"
)

const lifecycleFeedCursorSeparator = "\n"

var lifecycleStates = map[string]bool{
	"received": true, "validated": true, "accepted": true, "sent": true,
	"delivered": true, "read": true, "failed": true, "pending_aged": true,
}

func (h *handler) ListLifecycleEvents(ctx *gin.Context) {
	filters, err := h.lifecycleFilters(ctx)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	events, err := h.service.ListLifecycleEvents(ctx.Request.Context(), filters)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list message lifecycle events"})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"data": events})
}

func (h *handler) ApplicationLifecycleFeed(ctx *gin.Context) {
	identity := h.service.Identify(
		ctx.Request.Context(),
		ctx.GetHeader("X-Argo-Application-Id"),
		ctx.GetHeader("X-Argo-Application-Key"),
	)
	if !identity.Verified || identity.ApplicationID == nil {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "invalid application credential"})
		return
	}

	limit, err := lifecycleFeedLimit(ctx.Query("limit"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	cursor, err := decodeLifecycleFeedCursor(ctx.Query("cursor"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid lifecycle cursor"})
		return
	}
	events, err := h.service.ListLifecycleFeed(
		ctx.Request.Context(),
		*identity.ApplicationID,
		cursor,
		limit+1,
	)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list message lifecycle feed"})
		return
	}

	hasMore := len(events) > limit
	if hasMore {
		events = events[:limit]
	}
	items := make([]argo_model.LifecycleFeedItem, 0, len(events))
	for _, event := range events {
		items = append(items, lifecycleFeedItem(event))
	}
	nextCursor := ""
	if len(events) > 0 {
		last := events[len(events)-1]
		nextCursor = encodeLifecycleFeedCursor(argo_model.LifecycleFeedCursor{
			CreatedAt: last.CreatedAt,
			ID:        last.ID,
		})
	}
	ctx.JSON(http.StatusOK, argo_model.LifecycleFeedPage{
		Data:       items,
		NextCursor: nextCursor,
		HasMore:    hasMore,
	})
}

func lifecycleFeedLimit(value string) (int, error) {
	if value == "" {
		return 200, nil
	}
	limit, err := strconv.Atoi(value)
	if err != nil || limit < 1 || limit > 500 {
		return 0, errors.New("limit must be between 1 and 500")
	}
	return limit, nil
}

func encodeLifecycleFeedCursor(cursor argo_model.LifecycleFeedCursor) string {
	payload := cursor.CreatedAt.UTC().Format(time.RFC3339Nano) + lifecycleFeedCursorSeparator + cursor.ID
	return base64.RawURLEncoding.EncodeToString([]byte(payload))
}

func decodeLifecycleFeedCursor(value string) (argo_model.LifecycleFeedCursor, error) {
	if value == "" {
		return argo_model.LifecycleFeedCursor{}, nil
	}
	payload, err := base64.RawURLEncoding.DecodeString(value)
	if err != nil {
		return argo_model.LifecycleFeedCursor{}, err
	}
	parts := strings.SplitN(string(payload), lifecycleFeedCursorSeparator, 2)
	if len(parts) != 2 || strings.TrimSpace(parts[1]) == "" {
		return argo_model.LifecycleFeedCursor{}, errors.New("invalid cursor payload")
	}
	createdAt, err := time.Parse(time.RFC3339Nano, parts[0])
	if err != nil {
		return argo_model.LifecycleFeedCursor{}, err
	}
	return argo_model.LifecycleFeedCursor{CreatedAt: createdAt.UTC(), ID: parts[1]}, nil
}

func lifecycleFeedItem(event argo_model.MessageLifecycleEvent) argo_model.LifecycleFeedItem {
	return argo_model.LifecycleFeedItem{
		EventID:           event.ID,
		Application:       event.ApplicationSlug,
		InstanceID:        event.InstanceID,
		CorrelationID:     event.CorrelationID,
		ProviderMessageID: event.ProviderMessageID,
		IdempotencyKey:    event.IdempotencyKey,
		MessageType:       event.MessageType,
		State:             event.State,
		FailureCategory:   event.FailureCategory,
		FailureCode:       event.FailureCode,
		OccurredAt:        event.OccurredAt.UTC(),
		CreatedAt:         event.CreatedAt.UTC(),
	}
}

func (h *handler) LifecycleSummary(ctx *gin.Context) {
	filters, err := h.lifecycleFilters(ctx)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	summary, err := h.service.LifecycleSummary(ctx.Request.Context(), filters)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to load message lifecycle metrics"})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"data": summary})
}

func (h *handler) BackfillLifecycle(ctx *gin.Context) {
	var input argo_service.LifecycleBackfillInput
	if err := ctx.ShouldBindJSON(&input); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	report, err := h.service.BackfillLifecycle(ctx.Request.Context(), input)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"data": report})
}

func (h *handler) lifecycleFilters(ctx *gin.Context) (argo_model.LifecycleFilters, error) {
	to := h.now().UTC()
	from := to.Add(-7 * 24 * time.Hour)
	var err error
	if value := ctx.Query("from"); value != "" {
		from, err = time.Parse(time.RFC3339, value)
		if err != nil {
			return argo_model.LifecycleFilters{}, errors.New("from must use RFC3339 format")
		}
	}
	if value := ctx.Query("to"); value != "" {
		to, err = time.Parse(time.RFC3339, value)
		if err != nil {
			return argo_model.LifecycleFilters{}, errors.New("to must use RFC3339 format")
		}
	}
	if !from.Before(to) {
		return argo_model.LifecycleFilters{}, errors.New("from must be before to")
	}
	if to.Sub(from) > 366*24*time.Hour {
		return argo_model.LifecycleFilters{}, errors.New("period cannot exceed 366 days")
	}
	limit := 200
	if value := ctx.Query("limit"); value != "" {
		limit, err = strconv.Atoi(value)
		if err != nil || limit < 1 || limit > 500 {
			return argo_model.LifecycleFilters{}, errors.New("limit must be between 1 and 500")
		}
	}
	state := ctx.Query("state")
	if state != "" && !lifecycleStates[state] {
		return argo_model.LifecycleFilters{}, errors.New("invalid lifecycle state")
	}
	return argo_model.LifecycleFilters{
		From:              from,
		To:                to,
		ApplicationSlug:   ctx.Query("application"),
		InstanceID:        ctx.Query("instanceId"),
		MessageType:       ctx.Query("type"),
		State:             state,
		ProviderMessageID: ctx.Query("providerMessageId"),
		CorrelationID:     ctx.Query("correlationId"),
		IdempotencyKey:    ctx.Query("idempotencyKey"),
		Limit:             limit,
	}, nil
}
