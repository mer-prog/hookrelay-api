package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/mer-prog/hookrelay-api/internal/circuit"
	"github.com/mer-prog/hookrelay-api/internal/config"
	"github.com/mer-prog/hookrelay-api/internal/database"
	"github.com/mer-prog/hookrelay-api/internal/delivery"
	"github.com/mer-prog/hookrelay-api/internal/handler"
	"github.com/mer-prog/hookrelay-api/internal/pubsub"
	"github.com/mer-prog/hookrelay-api/internal/ws"
)

func main() {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		slog.Error("failed to load config", "error", err)
		os.Exit(1)
	}

	// Connect to database
	ctx := context.Background()
	db, err := database.New(ctx, cfg.DatabaseURL)
	if err != nil {
		slog.Error("failed to connect to database", "error", err)
		os.Exit(1)
	}
	defer db.Close()

	// Run migrations
	if err := database.RunMigrations(cfg.DatabaseURL, "migrations"); err != nil {
		slog.Error("failed to run migrations", "error", err)
		os.Exit(1)
	}

	// Initialize PubSub
	ps := pubsub.NewInMemory()
	defer ps.Close()

	// Initialize delivery engine
	breakerMgr := circuit.NewBreakerManager()
	engine := delivery.NewEngine(db.Pool, ps, breakerMgr, cfg.WorkerPoolSize)

	// Start WebSocket hub
	hub := ws.NewHub()
	hubCtx, hubCancel := context.WithCancel(ctx)
	defer hubCancel()
	go hub.Run(hubCtx)

	// Start delivery engine
	engine.Start(ctx)
	defer engine.Stop()

	// Setup router
	router := handler.SetupRouter(cfg, db.Pool, engine, hub)

	// Start HTTP server
	srv := &http.Server{
		Addr:         ":" + cfg.Port,
		Handler:      router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	go func() {
		slog.Info("server starting", "port", cfg.Port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			slog.Error("server error", "error", err)
			os.Exit(1)
		}
	}()

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	slog.Info("shutting down server...")
	shutdownCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		slog.Error("server forced to shutdown", "error", err)
	}

	slog.Info("server stopped")
}
