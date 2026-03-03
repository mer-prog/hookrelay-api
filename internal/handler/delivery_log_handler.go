package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/mer-prog/hookrelay-api/internal/model"
)

type DeliveryLogHandler struct {
	Pool *pgxpool.Pool
}

// ListDeliveryLogs handles GET /api/v1/logs
func (h *DeliveryLogHandler) ListDeliveryLogs(c *gin.Context) {
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "50"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))

	if limit <= 0 || limit > 100 {
		limit = 50
	}

	endpointIDStr := c.Query("endpoint_id")
	if endpointIDStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "endpoint_id query parameter is required"})
		return
	}

	endpointID, err := uuid.Parse(endpointIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid endpoint_id"})
		return
	}

	logs, err := model.ListDeliveryLogsByEndpointID(c.Request.Context(), h.Pool, endpointID, limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list delivery logs"})
		return
	}

	c.JSON(http.StatusOK, logs)
}

// GetDeliveryLog handles GET /api/v1/logs/:id
func (h *DeliveryLogHandler) GetDeliveryLog(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid delivery log ID"})
		return
	}

	log, err := model.GetDeliveryLogByID(c.Request.Context(), h.Pool, id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "delivery log not found"})
		return
	}

	c.JSON(http.StatusOK, log)
}
