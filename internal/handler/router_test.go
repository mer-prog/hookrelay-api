package handler

import (
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"

	"github.com/mer-prog/hookrelay-api/internal/circuit"
	"github.com/mer-prog/hookrelay-api/internal/config"
	"github.com/mer-prog/hookrelay-api/internal/delivery"
	"github.com/mer-prog/hookrelay-api/internal/pubsub"
	"github.com/mer-prog/hookrelay-api/internal/ws"
)

func init() {
	gin.SetMode(gin.TestMode)
}

func TestSetupRouter(t *testing.T) {
	cfg := &config.Config{
		JWTSecret:   "test-secret",
		FrontendURL: "http://localhost:3000",
	}

	ps := pubsub.NewInMemory()
	defer ps.Close()

	bm := circuit.NewBreakerManager()
	engine := delivery.NewEngine(nil, ps, bm, 1)
	hub := ws.NewHub()

	assert.NotPanics(t, func() {
		r := SetupRouter(cfg, nil, engine, hub)
		assert.NotNil(t, r)
	})
}

func TestSetupRouter_HealthEndpoint(t *testing.T) {
	cfg := &config.Config{
		JWTSecret:   "test-secret",
		FrontendURL: "http://localhost:3000",
	}

	ps := pubsub.NewInMemory()
	defer ps.Close()

	bm := circuit.NewBreakerManager()
	engine := delivery.NewEngine(nil, ps, bm, 1)
	hub := ws.NewHub()

	r := SetupRouter(cfg, nil, engine, hub)
	routes := r.Routes()

	var found bool
	for _, route := range routes {
		if route.Path == "/health" && route.Method == "GET" {
			found = true
			break
		}
	}
	assert.True(t, found, "/health route should be registered")
}
