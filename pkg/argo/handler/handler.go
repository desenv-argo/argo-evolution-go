package argo_handler

import (
	"errors"
	"net/http"
	"strconv"
	"time"

	argo_model "github.com/evolution-foundation/evolution-go/pkg/argo/model"
	argo_service "github.com/evolution-foundation/evolution-go/pkg/argo/service"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type Handler interface {
	CreateApplication(ctx *gin.Context)
	UpdateApplication(ctx *gin.Context)
	RotateCredential(ctx *gin.Context)
	ListApplications(ctx *gin.Context)
	ListAttempts(ctx *gin.Context)
	AttemptSummary(ctx *gin.Context)
}

type handler struct {
	service argo_service.Service
	now     func() time.Time
}

func (h *handler) CreateApplication(ctx *gin.Context) {
	var input argo_service.ApplicationInput
	if err := ctx.ShouldBindJSON(&input); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	result, err := h.service.CreateApplication(ctx.Request.Context(), input)
	if err != nil {
		if errors.Is(err, gorm.ErrDuplicatedKey) {
			ctx.JSON(http.StatusConflict, gin.H{"error": "application slug already exists"})
			return
		}
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusCreated, gin.H{"data": result})
}

func (h *handler) UpdateApplication(ctx *gin.Context) {
	var input argo_service.ApplicationInput
	if err := ctx.ShouldBindJSON(&input); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	application, err := h.service.UpdateApplication(ctx.Request.Context(), ctx.Param("applicationId"), input)
	if err != nil {
		status := http.StatusBadRequest
		if err.Error() == "application not found" {
			status = http.StatusNotFound
		}
		ctx.JSON(status, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"data": application})
}

func (h *handler) RotateCredential(ctx *gin.Context) {
	result, err := h.service.RotateCredential(ctx.Request.Context(), ctx.Param("applicationId"))
	if err != nil {
		status := http.StatusInternalServerError
		if err.Error() == "application not found" {
			status = http.StatusNotFound
		}
		ctx.JSON(status, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"data": result})
}

func (h *handler) ListApplications(ctx *gin.Context) {
	applications, err := h.service.ListApplications(ctx.Request.Context())
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list applications"})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"data": applications})
}

func (h *handler) ListAttempts(ctx *gin.Context) {
	filters, err := h.filters(ctx)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	attempts, err := h.service.ListAttempts(ctx.Request.Context(), filters)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list message attempts"})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"data": attempts})
}

func (h *handler) AttemptSummary(ctx *gin.Context) {
	filters, err := h.filters(ctx)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	summary, err := h.service.AttemptSummary(ctx.Request.Context(), filters)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to load attempt metrics"})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"data": summary})
}

func (h *handler) filters(ctx *gin.Context) (argo_model.AttemptFilters, error) {
	to := h.now().UTC()
	from := to.Add(-24 * time.Hour)
	var err error
	if value := ctx.Query("from"); value != "" {
		from, err = time.Parse(time.RFC3339, value)
		if err != nil {
			return argo_model.AttemptFilters{}, errors.New("from must use RFC3339 format")
		}
	}
	if value := ctx.Query("to"); value != "" {
		to, err = time.Parse(time.RFC3339, value)
		if err != nil {
			return argo_model.AttemptFilters{}, errors.New("to must use RFC3339 format")
		}
	}
	if !from.Before(to) {
		return argo_model.AttemptFilters{}, errors.New("from must be before to")
	}
	if to.Sub(from) > 366*24*time.Hour {
		return argo_model.AttemptFilters{}, errors.New("period cannot exceed 366 days")
	}
	limit := 100
	if value := ctx.Query("limit"); value != "" {
		limit, err = strconv.Atoi(value)
		if err != nil || limit < 1 || limit > 200 {
			return argo_model.AttemptFilters{}, errors.New("limit must be between 1 and 200")
		}
	}
	outcome := ctx.Query("outcome")
	if outcome != "" && outcome != "succeeded" && outcome != "failed" {
		return argo_model.AttemptFilters{}, errors.New("outcome must be succeeded or failed")
	}
	return argo_model.AttemptFilters{
		From:            from,
		To:              to,
		ApplicationSlug: ctx.Query("application"),
		InstanceID:      ctx.Query("instanceId"),
		Outcome:         outcome,
		ErrorCode:       ctx.Query("errorCode"),
		Limit:           limit,
	}, nil
}

func NewHandler(service argo_service.Service) Handler {
	return &handler{service: service, now: time.Now}
}

