package config

import (
	"fmt"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

type Config struct {
	DatabaseURL        string
	JWTSecret          string
	Port               string
	GoogleClientID     string
	GoogleClientSecret string
	GitHubClientID     string
	GitHubClientSecret string
	PubSubDriver       string
	RedisURL           string
	WorkerPoolSize     int
	FrontendURL        string
}

func Load() (*Config, error) {
	// Best-effort .env loading; ignore if file is missing.
	_ = godotenv.Load()

	cfg := &Config{
		Port:           envOrDefault("PORT", "8080"),
		PubSubDriver:   envOrDefault("PUBSUB_DRIVER", "memory"),
		FrontendURL:    envOrDefault("FRONTEND_URL", "http://localhost:3000"),
		DatabaseURL:    os.Getenv("DATABASE_URL"),
		JWTSecret:      os.Getenv("JWT_SECRET"),
		GoogleClientID: os.Getenv("GOOGLE_CLIENT_ID"),
		GoogleClientSecret: os.Getenv("GOOGLE_CLIENT_SECRET"),
		GitHubClientID:     os.Getenv("GITHUB_CLIENT_ID"),
		GitHubClientSecret: os.Getenv("GITHUB_CLIENT_SECRET"),
		RedisURL:           os.Getenv("REDIS_URL"),
	}

	poolSize, err := strconv.Atoi(envOrDefault("WORKER_POOL_SIZE", "10"))
	if err != nil {
		return nil, fmt.Errorf("parsing WORKER_POOL_SIZE: %w", err)
	}
	cfg.WorkerPoolSize = poolSize

	if cfg.DatabaseURL == "" {
		return nil, fmt.Errorf("DATABASE_URL is required")
	}
	if cfg.JWTSecret == "" {
		return nil, fmt.Errorf("JWT_SECRET is required")
	}

	if cfg.PubSubDriver != "memory" && cfg.PubSubDriver != "redis" {
		return nil, fmt.Errorf("PUBSUB_DRIVER must be \"memory\" or \"redis\", got %q", cfg.PubSubDriver)
	}
	if cfg.PubSubDriver == "redis" && cfg.RedisURL == "" {
		return nil, fmt.Errorf("REDIS_URL is required when PUBSUB_DRIVER is \"redis\"")
	}

	return cfg, nil
}

func envOrDefault(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
