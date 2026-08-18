package argo_handler

import (
	"mime"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

func (h *handler) GetMessageMedia(ctx *gin.Context) {
	media, err := h.service.GetMessageMedia(
		ctx.Request.Context(), ctx.Query("instanceId"), ctx.Param("messageId"),
	)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if media == nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "message media not found"})
		return
	}
	disposition := "inline"
	if strings.EqualFold(ctx.Query("download"), "true") {
		disposition = "attachment"
	}
	ctx.Header("Content-Disposition", mime.FormatMediaType(disposition, map[string]string{"filename": media.FileName}))
	ctx.Header("Cache-Control", "private, max-age=300")
	ctx.Header("X-Content-Type-Options", "nosniff")
	ctx.Data(http.StatusOK, media.MimeType, media.Content)
}
