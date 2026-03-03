package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/mer-prog/hookrelay-api/internal/model"
)

type EndpointHandler struct {
	Pool *pgxpool.Pool
}

// CreateEndpoint handles POST /api/v1/endpoints
func (h *EndpointHandler) CreateEndpoint(c *gin.Context) {
	userID, err := uuid.Parse(c.GetString("userID"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user ID"})
		return
	}

	var req struct {
		URL        string   `json:"url" binding:"required,url"`
		Secret     string   `json:"secret" binding:"required"`
		EventTypes []string `json:"event_types" binding:"required"`
		MaxRetries *int     `json:"max_retries"`
		TimeoutMs  *int     `json:"timeout_ms"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ep := &model.Endpoint{
		UserID:     userID,
		URL:        req.URL,
		Secret:     req.Secret,
		EventTypes: req.EventTypes,
		IsActive:   true,
		MaxRetries: 5,
		TimeoutMs:  30000,
	}
	if req.MaxRetries != nil {
		ep.MaxRetries = *req.MaxRetries
	}
	if req.TimeoutMs != nil {
		ep.TimeoutMs = *req.TimeoutMs
	}

	if err := model.CreateEndpoint(c.Request.Context(), h.Pool, ep); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create endpoint"})
		return
	}

	c.JSON(http.StatusCreated, ep)
}

// ListEndpoints handles GET /api/v1/endpoints
func (h *EndpointHandler) ListEndpoints(c *gin.Context) {
	userID, err := uuid.Parse(c.GetString("userID"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user ID"})
		return
	}

	endpoints, err := model.ListEndpointsByUserID(c.Request.Context(), h.Pool, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list endpoints"})
		return
	}

	c.JSON(http.StatusOK, endpoints)
}

// GetEndpoint handles GET /api/v1/endpoints/:id
func (h *EndpointHandler) GetEndpoint(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid endpoint ID"})
		return
	}

	ep, err := model.GetEndpointByID(c.Request.Context(), h.Pool, id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "endpoint not found"})
		return
	}

	c.JSON(http.StatusOK, ep)
}

// UpdateEndpoint handles PUT /api/v1/endpoints/:id
func (h *EndpointHandler) UpdateEndpoint(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid endpoint ID"})
		return
	}

	ep, err := model.GetEndpointByID(c.Request.Context(), h.Pool, id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "endpoint not found"})
		return
	}

	var req struct {
		URL        *string  `json:"url"`
		Secret     *string  `json:"secret"`
		EventTypes []string `json:"event_types"`
		IsActive   *bool    `json:"is_active"`
		MaxRetries *int     `json:"max_retries"`
		TimeoutMs  *int     `json:"timeout_ms"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if req.URL != nil {
		ep.URL = *req.URL
	}
	if req.Secret != nil {
		ep.Secret = *req.Secret
	}
	if req.EventTypes != nil {
		ep.EventTypes = req.EventTypes
	}
	if req.IsActive != nil {
		ep.IsActive = *req.IsActive
	}
	if req.MaxRetries != nil {
		ep.MaxRetries = *req.MaxRetries
	}
	if req.TimeoutMs != nil {
		ep.TimeoutMs = *req.TimeoutMs
	}

	if err := model.UpdateEndpoint(c.Request.Context(), h.Pool, ep); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update endpoint"})
		return
	}

	c.JSON(http.StatusOK, ep)
}

// DeleteEndpoint handles DELETE /api/v1/endpoints/:id (logical delete: is_active = false)
func (h *EndpointHandler) DeleteEndpoint(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid endpoint ID"})
		return
	}

	ep, err := model.GetEndpointByID(c.Request.Context(), h.Pool, id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "endpoint not found"})
		return
	}

	ep.IsActive = false
	if err := model.UpdateEndpoint(c.Request.Context(), h.Pool, ep); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete endpoint"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "endpoint deactivated"})
}
