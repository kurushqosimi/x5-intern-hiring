package handlers

import (
	"github.com/google/uuid"
	"github.com/kurushqosimi/x5-intern-hiring/internal/services"
	"go.uber.org/zap"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/kurushqosimi/x5-intern-hiring/internal/models"
)

//curl -X POST http://localhost:8080/api/v1/applications/reject \
//-H "Content-Type: application/json" \
//-d '{"application_ids":["<uuid1>"],"status_reason":"Не подошли по требованиям"}'

func (h *Handler) InviteApplications(ctx *gin.Context) {
	var req models.BulkEmailActionRequest
	if err := ctx.ShouldBindJSON(&req); err != nil || len(req.ApplicationIDs) == 0 {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "application_ids обязателен"})
		return
	}

	for _, id := range req.ApplicationIDs {
		if _, err := uuid.Parse(id); err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid application_id: " + id})
			return
		}
	}

	res, err := h.service.Invite(ctx.Request.Context(), req)
	if err != nil {
		h.logger.Error("h.service.Invite: ", zap.Error(err))
		// template not found => 400
		if services.IsTemplateNotFound(err) {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "template_code не найден или не активен"})
			return
		}
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "внутренняя ошибка сервера"})
		return
	}

	ctx.JSON(http.StatusAccepted, res)
}

func (h *Handler) RejectApplications(ctx *gin.Context) {
	var req models.BulkEmailActionRequest
	if err := ctx.ShouldBindJSON(&req); err != nil || len(req.ApplicationIDs) == 0 {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "application_ids обязателен"})
		return
	}

	res, err := h.service.Reject(ctx.Request.Context(), req)
	if err != nil {
		h.logger.Error("h.service.Reject: ", zap.Error(err))
		if services.IsTemplateNotFound(err) {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "template_code не найден или не активен"})
			return
		}
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "внутренняя ошибка сервера"})
		return
	}

	ctx.JSON(http.StatusAccepted, res)
}
