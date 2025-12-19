package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/kurushqosimi/x5-intern-hiring/internal/models"
)

//curl -X POST http://localhost:8080/api/v1/applications/crm/queue -H "Content-Type: application/json" -d '{"application_ids":["", ""]}'

func (h *Handler) QueueApplicationsToCRM(ctx *gin.Context) {
	var req models.BulkCRMActionRequest
	if err := ctx.ShouldBindJSON(&req); err != nil || len(req.ApplicationIDs) == 0 {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "application_ids обязателен"})
		return
	}

	res, err := h.service.QueueToCRM(ctx.Request.Context(), req)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusAccepted, res)
}
