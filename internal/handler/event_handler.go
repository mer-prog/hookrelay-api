package handler

import (
	"encoding/json"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/mer-prog/hookrelay-api/internal/delivery"
	"github.com/mer-prog/hookrelay-api/internal/model"
)

type EventHandler struct {
	Pool   *pgxpool.Pool
	Engine *delivery.Engine
}

// CreateEvent handles POST /api/v1/events
func (h *EventHandler) CreateEvent(c *gin.Context) {
	userID, err := uuid.Parse(c.GetString("userID"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user ID"})
		return
	}

	var req struct {
		EventType string          `json:"event_type" binding:"required"`
		Payload   json.RawMessage `json:"payload" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	event := &model.Event{
		UserID:    userID,
		EventType: req.EventType,
		Payload:   req.Payload,
	}
	if err := model.CreateEvent(c.Request.Context(), h.Pool, event); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create event"})
		return
	}

	endpoints, err := model.ListActiveEndpointsByEventType(c.Request.Context(), h.Pool, userID, req.EventType)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list endpoints"})
		return
	}

	if len(endpoints) > 0 {
		h.Engine.Dispatch(c.Request.Context(), *event, endpoints)
	}

	c.JSON(http.StatusCreated, event)
}

// GetEvent handles GET /api/v1/events/:id
func (h *EventHandler) GetEvent(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid event ID"})
		return
	}

	event, err := model.GetEventByID(c.Request.Context(), h.Pool, id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "event not found"})
		return
	}

	c.JSON(http.StatusOK, event)
}

// ListEventDeliveries handles GET /api/v1/events/:id/deliveries
func (h *EventHandler) ListEventDeliveries(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid event ID"})
		return
	}

	logs, err := model.ListDeliveryLogsByEventID(c.Request.Context(), h.Pool, id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list deliveries"})
		return
	}

	c.JSON(http.StatusOK, logs)
}
