package argo_middleware

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"strings"
	"time"

	argo_model "github.com/evolution-foundation/evolution-go/pkg/argo/model"
	argo_service "github.com/evolution-foundation/evolution-go/pkg/argo/service"
	instance_model "github.com/evolution-foundation/evolution-go/pkg/instance/model"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

const maxCapturedResponseBytes = 64 * 1024

type AttemptTracker struct {
	service argo_service.Service
}

func (t *AttemptTracker) Track() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		startedAt := time.Now().UTC()
		correlationID := cleanHeader(ctx.GetHeader("X-Correlation-Id"), 64)
		if correlationID == "" {
			correlationID = uuid.NewString()
		}
		ctx.Request.Header.Set("X-Correlation-Id", correlationID)
		ctx.Header("X-Correlation-Id", correlationID)

		identityCtx, cancelIdentity := context.WithTimeout(context.Background(), 2*time.Second)
		identity := t.service.Identify(
			identityCtx,
			ctx.GetHeader("X-Argo-Application-Id"),
			ctx.GetHeader("X-Argo-Application-Key"),
		)
		cancelIdentity()

		writer := &captureWriter{ResponseWriter: ctx.Writer}
		ctx.Writer = writer
		ctx.Next()

		completedAt := time.Now().UTC()
		status := ctx.Writer.Status()
		outcome := "succeeded"
		if status < http.StatusOK || status >= http.StatusMultipleChoices {
			outcome = "failed"
		}

		var instanceID *string
		if value, exists := ctx.Get("instance"); exists {
			if instance, ok := value.(*instance_model.Instance); ok && instance != nil {
				id := instance.Id
				instanceID = &id
			}
		}

		errorDetail := responseError(writer.body.Bytes())
		attempt := &argo_model.MessageAttempt{
			ApplicationID:     identity.ApplicationID,
			ApplicationSlug:   identity.ApplicationSlug,
			IdentityVerified:  identity.Verified,
			InstanceID:        instanceID,
			CorrelationID:     correlationID,
			IdempotencyKey:    cleanHeader(ctx.GetHeader("Idempotency-Key"), 128),
			ProviderMessageID: responseMessageID(writer.body.Bytes()),
			Endpoint:          ctx.FullPath(),
			Method:            ctx.Request.Method,
			HTTPStatus:        status,
			Outcome:           outcome,
			ErrorCode:         classifyError(status, errorDetail),
			ErrorDetail:       cleanHeader(errorDetail, 500),
			DurationMS:        completedAt.Sub(startedAt).Milliseconds(),
			StartedAt:         startedAt,
			CompletedAt:       completedAt,
		}
		if attempt.Endpoint == "" {
			attempt.Endpoint = ctx.Request.URL.Path
		}

		recordCtx, cancelRecord := context.WithTimeout(context.Background(), 3*time.Second)
		_ = t.service.RecordAttempt(recordCtx, attempt)
		cancelRecord()
	}
}

type captureWriter struct {
	gin.ResponseWriter
	body bytes.Buffer
}

func (w *captureWriter) Write(data []byte) (int, error) {
	w.capture(data)
	return w.ResponseWriter.Write(data)
}

func (w *captureWriter) WriteString(value string) (int, error) {
	w.capture([]byte(value))
	return w.ResponseWriter.WriteString(value)
}

func (w *captureWriter) capture(data []byte) {
	remaining := maxCapturedResponseBytes - w.body.Len()
	if remaining <= 0 {
		return
	}
	if len(data) > remaining {
		data = data[:remaining]
	}
	_, _ = w.body.Write(data)
}

func responseError(body []byte) string {
	var payload struct {
		Error   string `json:"error"`
		Message string `json:"message"`
	}
	if json.Unmarshal(body, &payload) != nil {
		return ""
	}
	if payload.Error != "" {
		return payload.Error
	}
	return payload.Message
}

func responseMessageID(body []byte) string {
	var payload struct {
		Data struct {
			Info struct {
				ID string `json:"ID"`
			} `json:"Info"`
		} `json:"data"`
	}
	if json.Unmarshal(body, &payload) != nil {
		return ""
	}
	return cleanHeader(payload.Data.Info.ID, 128)
}

func classifyError(status int, detail string) string {
	if status >= http.StatusOK && status < http.StatusMultipleChoices {
		return ""
	}
	message := strings.ToLower(detail)
	switch {
	case status == http.StatusUnauthorized || status == http.StatusForbidden:
		return "AUTH_FAILED"
	case status == http.StatusTooManyRequests:
		return "RATE_LIMITED"
	case strings.Contains(message, "no active session"), strings.Contains(message, "client disconnected"), strings.Contains(message, "not connected"):
		return "INSTANCE_OFFLINE"
	case strings.Contains(message, "invalid phone"), strings.Contains(message, "invalid jid"), strings.Contains(message, "user not found"):
		return "INVALID_RECIPIENT"
	case strings.Contains(message, "timeout"), strings.Contains(message, "timed out"), strings.Contains(message, "deadline exceeded"):
		return "WHATSAPP_TIMEOUT"
	case strings.Contains(message, "media"), strings.Contains(message, "upload"), strings.Contains(message, "download"):
		return "MEDIA_FAILED"
	case status >= http.StatusBadRequest && status < http.StatusInternalServerError:
		return "VALIDATION_FAILED"
	default:
		return "INTERNAL_ERROR"
	}
}

func cleanHeader(value string, limit int) string {
	value = strings.TrimSpace(strings.ReplaceAll(strings.ReplaceAll(value, "\n", " "), "\r", " "))
	if len(value) > limit {
		return value[:limit]
	}
	return value
}

func NewAttemptTracker(service argo_service.Service) *AttemptTracker {
	return &AttemptTracker{service: service}
}
