# HookRelay API

## Overview
Real-time webhook delivery platform with live monitoring dashboard.
Go API + Next.js frontend (separate repos). This repo is the Go API.

## Architecture
- Go 1.24 + Gin framework
- PostgreSQL (Neon Serverless) via pgx v5
- WebSocket (gorilla/websocket) for real-time delivery feed
- Interface-driven PubSub (InMemory default / Redis production swap)
- golang-migrate for DB migrations

## Directory Structure
- cmd/server/main.go       → Entry point
- internal/config/          → Environment variable loading
- internal/auth/            → JWT + OAuth 2.0 (Google, GitHub)
- internal/handler/         → HTTP handlers (Gin)
- internal/model/           → DB models and queries (pgx, raw SQL)
- internal/delivery/        → Webhook delivery engine (worker pool, retry, signer)
- internal/circuit/         → Circuit breaker
- internal/pubsub/          → PubSub interface (memory.go / redis.go)
- internal/ws/              → WebSocket hub + client management
- internal/middleware/       → Rate limiter, CORS, request logger
- internal/database/        → PostgreSQL connection + migration runner
- migrations/               → SQL migration files (golang-migrate format)
- seed/                     → Test data generation

## Commands
- make run            → Start dev server (go run cmd/server/main.go)
- make build          → Build production binary to bin/hookrelay
- make test           → Run all tests
- make test-coverage  → Tests with coverage report
- make lint           → golangci-lint run
- make migrate-up     → Apply migrations
- make migrate-down   → Rollback last migration
- make seed           → Populate test data
- make docker-build   → Build Docker image

## Code Standards
- Go standard project layout (cmd/ + internal/)
- Table-driven tests with testify
- Interface-driven design for PubSub, CircuitBreaker
- pgx for raw SQL (NO ORM)
- Error wrapping: fmt.Errorf("operation: %w", err)
- Context propagation on all DB and HTTP calls
- Structured logging with slog

## Key Design Patterns
- PubSub interface: PUBSUB_DRIVER env switches memory/redis
- Worker pool: buffered channel + N goroutines for concurrent delivery
- Circuit breaker: per-endpoint, 10 consecutive failures → OPEN
- Exponential backoff: 1s × 2^attempt + jitter, max 5 retries
- HMAC-SHA256 webhook signing on all outbound requests
- JWT access (15min) + refresh token (7d) rotation

## Environment Variables
- DATABASE_URL (required)
- JWT_SECRET (required)
- GOOGLE_CLIENT_ID, GOOGLE_CLIENT_SECRET (optional)
- GITHUB_CLIENT_ID, GITHUB_CLIENT_SECRET (optional)
- PUBSUB_DRIVER=memory|redis (default: memory)
- REDIS_URL (if PUBSUB_DRIVER=redis)
- WORKER_POOL_SIZE (default: 10)
- PORT (default: 8080)
