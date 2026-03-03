package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type StatsHandler struct {
	Pool *pgxpool.Pool
}

type Stats struct {
	TotalEvents       int     `json:"total_events"`
	TotalDeliveries   int     `json:"total_deliveries"`
	SuccessRate       float64 `json:"success_rate"`
	AvgLatencyMs      float64 `json:"avg_latency_ms"`
	ActiveEndpoints   int     `json:"active_endpoints"`
	PendingDeliveries int     `json:"pending_deliveries"`
}

// GetStats handles GET /api/v1/stats
func (h *StatsHandler) GetStats(c *gin.Context) {
	userID, err := uuid.Parse(c.GetString("userID"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user ID"})
		return
	}

	ctx := c.Request.Context()
	var stats Stats

	err = h.Pool.QueryRow(ctx,
		`SELECT COUNT(*) FROM events WHERE user_id = $1`, userID,
	).Scan(&stats.TotalEvents)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get stats"})
		return
	}

	err = h.Pool.QueryRow(ctx, `
		SELECT
			COUNT(*),
			COALESCE(AVG(CASE WHEN status = 'SUCCESS' THEN 1.0 ELSE 0.0 END) * 100, 0),
			COALESCE(AVG(latency_ms), 0)
		FROM delivery_logs dl
		JOIN endpoints ep ON dl.endpoint_id = ep.id
		WHERE ep.user_id = $1`, userID,
	).Scan(&stats.TotalDeliveries, &stats.SuccessRate, &stats.AvgLatencyMs)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get delivery stats"})
		return
	}

	err = h.Pool.QueryRow(ctx,
		`SELECT COUNT(*) FROM endpoints WHERE user_id = $1 AND is_active = true`, userID,
	).Scan(&stats.ActiveEndpoints)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get endpoint stats"})
		return
	}

	err = h.Pool.QueryRow(ctx, `
		SELECT COUNT(*)
		FROM delivery_logs dl
		JOIN endpoints ep ON dl.endpoint_id = ep.id
		WHERE ep.user_id = $1 AND dl.status = 'PENDING'`, userID,
	).Scan(&stats.PendingDeliveries)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get pending stats"})
		return
	}

	c.JSON(http.StatusOK, stats)
}
