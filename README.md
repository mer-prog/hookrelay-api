# HookRelay API

Real-time webhook delivery platform with live monitoring dashboard. This is the Go API backend — the Next.js frontend lives in a separate repository.

## Tech Stack

- **Language:** Go 1.24+
- **Framework:** Gin
- **Database:** PostgreSQL (Neon Serverless) via pgx v5
- **WebSocket:** gorilla/websocket
- **Auth:** JWT (HS256) + OAuth 2.0 (Google, GitHub)
- **Migrations:** golang-migrate
- **PubSub:** In-memory (default) / Redis (production)

## Architecture

```
┌──────────────────────────────────────────────────────┐
│                    Gin HTTP Router                    │
│  ┌────────┐ ┌───────────┐ ┌──────┐ ┌─────────────┐  │
│  │  Auth  │ │ Endpoints │ │Events│ │  Delivery   │  │
│  │Handler │ │  Handler  │ │Hndlr │ │ Log Handler │  │
│  └────────┘ └───────────┘ └──┬───┘ └─────────────┘  │
│                              │                       │
│  ┌───────────────────────────▼────────────────────┐  │
│  │            Delivery Engine (Worker Pool)        │  │
│  │  ┌────────┐ ┌────────┐ ┌────────┐ ┌────────┐  │  │
│  │  │Worker 0│ │Worker 1│ │Worker 2│ │Worker N│  │  │
│  │  └───┬────┘ └───┬────┘ └───┬────┘ └───┬────┘  │  │
│  │      │     HTTP POST + HMAC-SHA256     │       │  │
│  └──────┼──────────┼──────────┼───────────┼───────┘  │
│         │          │          │           │          │
│  ┌──────▼──────────▼──────────▼───────────▼───────┐  │
│  │              Circuit Breaker Manager            │  │
│  └─────────────────────┬──────────────────────────┘  │
│                        │                             │
│  ┌─────────────────────▼──────────────────────────┐  │
│  │           PubSub (Memory / Redis)              │  │
│  └─────────────────────┬──────────────────────────┘  │
│                        │                             │
│  ┌─────────────────────▼──────────────────────────┐  │
│  │         WebSocket Hub → Live Dashboard          │  │
│  └────────────────────────────────────────────────┘  │
│                                                      │
│  ┌────────────────────────────────────────────────┐  │
│  │            PostgreSQL (pgx v5 pool)             │  │
│  └────────────────────────────────────────────────┘  │
└──────────────────────────────────────────────────────┘
```

## Quick Start

### Prerequisites

- Go 1.24+
- PostgreSQL 15+ (or Docker)
- Make

### Using Docker Compose

```bash
cp .env.example .env
docker compose up -d
```

### Local Development

```bash
# 1. Start PostgreSQL (or use Docker)
docker compose up -d db

# 2. Configure environment
cp .env.example .env
# Edit .env with your DATABASE_URL and JWT_SECRET

# 3. Run migrations and start
make run
```

### Seed Test Data

```bash
make seed
```

## Commands

| Command              | Description                         |
|----------------------|-------------------------------------|
| `make run`           | Start dev server                    |
| `make build`         | Build production binary             |
| `make test`          | Run all tests                       |
| `make test-coverage` | Tests with HTML coverage report     |
| `make lint`          | Run golangci-lint                   |
| `make migrate-up`    | Apply database migrations           |
| `make migrate-down`  | Rollback last migration             |
| `make seed`          | Populate test data                  |
| `make docker-build`  | Build Docker image                  |

## API Endpoints

### Authentication (no auth required)

| Method | Path                            | Description           |
|--------|---------------------------------|-----------------------|
| GET    | `/api/v1/auth/google`           | Google OAuth start    |
| GET    | `/api/v1/auth/google/callback`  | Google OAuth callback |
| GET    | `/api/v1/auth/github`           | GitHub OAuth start    |
| GET    | `/api/v1/auth/github/callback`  | GitHub OAuth callback |
| POST   | `/api/v1/auth/refresh`          | Refresh access token  |
| POST   | `/api/v1/auth/logout`           | Logout (stateless)    |

### Event Ingestion (API Key auth via `X-API-Key` header)

| Method | Path              | Description    |
|--------|-------------------|----------------|
| POST   | `/api/v1/events`  | Ingest event   |

### Dashboard (JWT auth via `Authorization: Bearer <token>`)

| Method | Path                             | Description               |
|--------|----------------------------------|---------------------------|
| GET    | `/api/v1/events/:id`             | Event details             |
| GET    | `/api/v1/events/:id/deliveries`  | Event delivery logs       |
| POST   | `/api/v1/endpoints`              | Create endpoint           |
| GET    | `/api/v1/endpoints`              | List endpoints            |
| GET    | `/api/v1/endpoints/:id`          | Endpoint details          |
| PUT    | `/api/v1/endpoints/:id`          | Update endpoint           |
| DELETE | `/api/v1/endpoints/:id`          | Deactivate endpoint       |
| GET    | `/api/v1/logs`                   | List delivery logs        |
| GET    | `/api/v1/logs/:id`               | Delivery log details      |
| POST   | `/api/v1/keys`                   | Create API key            |
| GET    | `/api/v1/keys`                   | List API keys             |
| DELETE | `/api/v1/keys/:id`               | Revoke API key            |
| GET    | `/api/v1/stats`                  | Dashboard statistics      |

### WebSocket

| Path  | Description                 |
|-------|-----------------------------|
| `/ws` | Live delivery feed (JWT)    |

### Health

| Method | Path      | Description  |
|--------|-----------|--------------|
| GET    | `/health` | Health check |

## Environment Variables

| Variable               | Required | Default              | Description                   |
|------------------------|----------|----------------------|-------------------------------|
| `DATABASE_URL`         | Yes      | —                    | PostgreSQL connection string  |
| `JWT_SECRET`           | Yes      | —                    | HMAC signing key for JWTs     |
| `PORT`                 | No       | `8080`               | HTTP server port              |
| `FRONTEND_URL`         | No       | `http://localhost:3000` | CORS allowed origin        |
| `GOOGLE_CLIENT_ID`     | No       | —                    | Google OAuth client ID        |
| `GOOGLE_CLIENT_SECRET` | No       | —                    | Google OAuth client secret    |
| `GITHUB_CLIENT_ID`     | No       | —                    | GitHub OAuth client ID        |
| `GITHUB_CLIENT_SECRET` | No       | —                    | GitHub OAuth client secret    |
| `PUBSUB_DRIVER`        | No       | `memory`             | `memory` or `redis`           |
| `REDIS_URL`            | No       | —                    | Redis URL (if driver=redis)   |
| `WORKER_POOL_SIZE`     | No       | `10`                 | Concurrent delivery workers   |

## License

MIT
