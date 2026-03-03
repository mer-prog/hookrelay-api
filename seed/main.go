package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"math/rand/v2"
	"os"
	"time"

	"github.com/google/uuid"
	"github.com/joho/godotenv"

	"github.com/mer-prog/hookrelay-api/internal/database"
	"github.com/mer-prog/hookrelay-api/internal/model"
)

func main() {
	_ = godotenv.Load()

	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		slog.Error("DATABASE_URL is required")
		os.Exit(1)
	}

	ctx := context.Background()
	db, err := database.New(ctx, dbURL)
	if err != nil {
		slog.Error("failed to connect to database", "error", err)
		os.Exit(1)
	}
	defer db.Close()

	if err := database.RunMigrations(dbURL, "migrations"); err != nil {
		slog.Error("failed to run migrations", "error", err)
		os.Exit(1)
	}

	slog.Info("seeding database...")

	// ---- Users ----
	users := []*model.User{
		{Email: "admin@hookrelay.dev", Name: "Admin User", Provider: "google", ProviderID: "g-admin-001", Role: "admin"},
		{Email: "dev@hookrelay.dev", Name: "Developer User", Provider: "github", ProviderID: "gh-dev-001", Role: "developer"},
		{Email: "viewer@hookrelay.dev", Name: "Viewer User", Provider: "github", ProviderID: "gh-viewer-001", Role: "viewer"},
	}
	for _, u := range users {
		if err := model.CreateUser(ctx, db.Pool, u); err != nil {
			slog.Error("failed to create user", "email", u.Email, "error", err)
			os.Exit(1)
		}
		slog.Info("created user", "email", u.Email, "id", u.ID)
	}

	// ---- Endpoints ----
	endpointDefs := []struct {
		userIdx    int
		url        string
		eventTypes []string
	}{
		{0, "https://httpbin.org/post", []string{"order.created", "payment.completed"}},
		{0, "https://httpbin.org/status/500", []string{"order.created"}},
		{1, "https://httpbin.org/post", []string{"user.registered", "notification.sent"}},
		{1, "https://httpbin.org/delay/5", []string{"invoice.generated"}},
		{2, "https://httpbin.org/post", []string{"order.created", "user.registered", "notification.sent"}},
	}

	endpoints := make([]*model.Endpoint, len(endpointDefs))
	for i, def := range endpointDefs {
		ep := &model.Endpoint{
			UserID:     users[def.userIdx].ID,
			URL:        def.url,
			Secret:     fmt.Sprintf("whsec_seed_%d_%s", i, uuid.New().String()[:8]),
			EventTypes: def.eventTypes,
			IsActive:   true,
			MaxRetries: 5,
			TimeoutMs:  30000,
		}
		if err := model.CreateEndpoint(ctx, db.Pool, ep); err != nil {
			slog.Error("failed to create endpoint", "url", ep.URL, "error", err)
			os.Exit(1)
		}
		endpoints[i] = ep
		slog.Info("created endpoint", "url", ep.URL, "id", ep.ID)
	}

	// ---- Events ----
	eventTypes := []string{"order.created", "payment.completed", "user.registered", "notification.sent", "invoice.generated"}
	payloads := []map[string]any{
		{"order_id": "ord-001", "amount": 99.99, "currency": "USD"},
		{"payment_id": "pay-001", "method": "credit_card", "amount": 99.99},
		{"user_id": "usr-001", "email": "new@example.com", "plan": "pro"},
		{"notification_id": "ntf-001", "channel": "email", "template": "welcome"},
		{"invoice_id": "inv-001", "total": 149.99, "due_date": "2026-04-01"},
	}

	events := make([]*model.Event, 50)
	for i := range 50 {
		etIdx := rand.IntN(len(eventTypes))
		userIdx := rand.IntN(len(users))

		payload := make(map[string]any)
		for k, v := range payloads[etIdx] {
			payload[k] = v
		}
		payload["seed_index"] = i
		payload["timestamp"] = time.Now().UTC().Format(time.RFC3339)
		payloadJSON, _ := json.Marshal(payload)

		ev := &model.Event{
			UserID:    users[userIdx].ID,
			EventType: eventTypes[etIdx],
			Payload:   payloadJSON,
		}
		if err := model.CreateEvent(ctx, db.Pool, ev); err != nil {
			slog.Error("failed to create event", "error", err)
			os.Exit(1)
		}
		events[i] = ev
	}
	slog.Info("created events", "count", len(events))

	// ---- Delivery Logs ----
	// Distribution: 60 SUCCESS, 20 FAILED, 10 RETRYING, 60 PENDING
	statuses := make([]string, 0, 150)
	for range 60 {
		statuses = append(statuses, "SUCCESS")
	}
	for range 20 {
		statuses = append(statuses, "FAILED")
	}
	for range 10 {
		statuses = append(statuses, "RETRYING")
	}
	for range 60 {
		statuses = append(statuses, "PENDING")
	}
	// Shuffle
	rand.Shuffle(len(statuses), func(i, j int) { statuses[i], statuses[j] = statuses[j], statuses[i] })

	for i, status := range statuses {
		ev := events[rand.IntN(len(events))]
		ep := endpoints[rand.IntN(len(endpoints))]
		latency := 50 + rand.IntN(2951) // 50–3000ms

		dl := &model.DeliveryLog{
			EventID:       ev.ID,
			EndpointID:    ep.ID,
			Status:        status,
			AttemptNumber: 1 + rand.IntN(5),
			LatencyMs:     &latency,
		}

		if status == "SUCCESS" {
			code := 200
			dl.ResponseStatus = &code
			body := `{"status":"ok"}`
			dl.ResponseBody = &body
		} else if status == "FAILED" {
			code := []int{400, 401, 403, 404, 500}[rand.IntN(5)]
			dl.ResponseStatus = &code
			errMsg := fmt.Sprintf("HTTP %d response", code)
			dl.ErrorMessage = &errMsg
		} else if status == "RETRYING" || status == "PENDING" {
			nextRetry := time.Now().UTC().Add(time.Duration(rand.IntN(300)) * time.Second)
			dl.NextRetryAt = &nextRetry
			if status == "RETRYING" {
				code := 503
				dl.ResponseStatus = &code
				errMsg := "HTTP 503 response"
				dl.ErrorMessage = &errMsg
			}
		}

		reqHeaders, _ := json.Marshal(map[string]string{
			"Content-Type":       "application/json",
			"X-HookRelay-Event": ev.EventType,
			"X-HookRelay-ID":    ev.ID.String(),
		})
		dl.RequestHeaders = reqHeaders

		if err := model.CreateDeliveryLog(ctx, db.Pool, dl); err != nil {
			slog.Error("failed to create delivery log", "index", i, "error", err)
			os.Exit(1)
		}
	}
	slog.Info("created delivery logs", "count", len(statuses))

	slog.Info("seeding complete!")
}
