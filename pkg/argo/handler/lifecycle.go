package argo_handler

import (
	"errors"
	"net/http"
	"strconv"
	"time"

	argo_model "github.com/evolution-foundation/evolution-go/pkg/argo/model"
	"github.com/gin-gonic/gin"
)

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
