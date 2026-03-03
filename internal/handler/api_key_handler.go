package handler

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/mer-prog/hookrelay-api/internal/model"
)

type APIKeyHandler struct {
	Pool *pgxpool.Pool
}

// CreateAPIKey handles POST /api/v1/keys
func (h *APIKeyHandler) CreateAPIKey(c *gin.Context) {
	userID, err := uuid.Parse(c.GetString("userID"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user ID"})
		return
	}

	var req struct {
		Name   string   `json:"name" binding:"required"`
		Scopes []string `json:"scopes"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	rawKey := make([]byte, 32)
	if _, err := rand.Read(rawKey); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to generate key"})
		return
	}
	keyStr := "hrk_" + hex.EncodeToString(rawKey)

	hash := sha256.Sum256([]byte(keyStr))
	keyHash := hex.EncodeToString(hash[:])

	apiKey := &model.APIKey{
		UserID:  userID,
		KeyHash: keyHash,
		Name:    req.Name,
		Scopes:  req.Scopes,
	}
	if err := model.CreateAPIKey(c.Request.Context(), h.Pool, apiKey); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create API key"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"id":         apiKey.ID,
		"key":        keyStr,
		"name":       apiKey.Name,
		"scopes":     apiKey.Scopes,
		"created_at": apiKey.CreatedAt,
	})
}

// ListAPIKeys handles GET /api/v1/keys
func (h *APIKeyHandler) ListAPIKeys(c *gin.Context) {
	userID, err := uuid.Parse(c.GetString("userID"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user ID"})
		return
	}

	keys, err := model.ListAPIKeysByUserID(c.Request.Context(), h.Pool, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list API keys"})
		return
	}

	c.JSON(http.StatusOK, keys)
}

// RevokeAPIKey handles DELETE /api/v1/keys/:id
func (h *APIKeyHandler) RevokeAPIKey(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid API key ID"})
		return
	}

	if err := model.RevokeAPIKey(c.Request.Context(), h.Pool, id); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "API key not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "API key revoked"})
}
