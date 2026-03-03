package handler

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/mer-prog/hookrelay-api/internal/auth"
	"github.com/mer-prog/hookrelay-api/internal/config"
	"github.com/mer-prog/hookrelay-api/internal/delivery"
	"github.com/mer-prog/hookrelay-api/internal/middleware"
	"github.com/mer-prog/hookrelay-api/internal/ws"
)

func SetupRouter(cfg *config.Config, pool *pgxpool.Pool, engine *delivery.Engine, hub *ws.Hub) *gin.Engine {
	r := gin.New()

	// Global middleware
	r.Use(gin.Recovery())
	r.Use(middleware.RequestLogger())
	r.Use(middleware.CORSMiddleware(cfg.FrontendURL))
	r.Use(middleware.RateLimiter(100, time.Minute))

	// Health check
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	v1 := r.Group("/api/v1")

	// ---- Auth routes (no auth required) ----
	authHandler := &AuthHandler{
		Pool:      pool,
		Config:    cfg,
		GoogleCfg: auth.GoogleOAuthConfig(cfg),
		GitHubCfg: auth.GitHubOAuthConfig(cfg),
	}
	authGroup := v1.Group("/auth")
	{
		authGroup.GET("/google", authHandler.GoogleLogin)
		authGroup.GET("/google/callback", authHandler.GoogleCallback)
		authGroup.GET("/github", authHandler.GitHubLogin)
		authGroup.GET("/github/callback", authHandler.GitHubCallback)
		authGroup.POST("/refresh", authHandler.RefreshToken)
		authGroup.POST("/logout", authHandler.Logout)
	}

	// ---- Event ingestion (API Key auth) ----
	eventHandler := &EventHandler{Pool: pool, Engine: engine}
	ingest := v1.Group("/events")
	ingest.Use(auth.APIKeyMiddleware(pool))
	{
		ingest.POST("", eventHandler.CreateEvent)
	}

	// ---- Dashboard routes (JWT auth) ----
	jwtAuth := auth.AuthMiddleware(cfg.JWTSecret)

	// Events (read-only via JWT)
	events := v1.Group("/events")
	events.Use(jwtAuth)
	{
		events.GET("/:id", eventHandler.GetEvent)
		events.GET("/:id/deliveries", eventHandler.ListEventDeliveries)
	}

	// Endpoints
	endpointHandler := &EndpointHandler{Pool: pool}
	endpoints := v1.Group("/endpoints")
	endpoints.Use(jwtAuth)
	{
		endpoints.POST("", endpointHandler.CreateEndpoint)
		endpoints.GET("", endpointHandler.ListEndpoints)
		endpoints.GET("/:id", endpointHandler.GetEndpoint)
		endpoints.PUT("/:id", endpointHandler.UpdateEndpoint)
		endpoints.DELETE("/:id", endpointHandler.DeleteEndpoint)
	}

	// Delivery logs
	logHandler := &DeliveryLogHandler{Pool: pool}
	logs := v1.Group("/logs")
	logs.Use(jwtAuth)
	{
		logs.GET("", logHandler.ListDeliveryLogs)
		logs.GET("/:id", logHandler.GetDeliveryLog)
	}

	// API keys
	keyHandler := &APIKeyHandler{Pool: pool}
	keys := v1.Group("/keys")
	keys.Use(jwtAuth)
	{
		keys.POST("", keyHandler.CreateAPIKey)
		keys.GET("", keyHandler.ListAPIKeys)
		keys.DELETE("/:id", keyHandler.RevokeAPIKey)
	}

	// Stats
	statsHandler := &StatsHandler{Pool: pool}
	stats := v1.Group("/stats")
	stats.Use(jwtAuth)
	{
		stats.GET("", statsHandler.GetStats)
	}

	// WebSocket
	r.GET("/ws", jwtAuth, func(c *gin.Context) {
		ws.ServeWs(hub, c.Writer, c.Request, c.GetString("userID"))
	})

	return r
}
