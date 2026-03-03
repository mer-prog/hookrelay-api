package model

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Endpoint struct {
	ID                  uuid.UUID `json:"id"`
	UserID              uuid.UUID `json:"user_id"`
	URL                 string    `json:"url"`
	Secret              string    `json:"-"`
	EventTypes          []string  `json:"event_types"`
	IsActive            bool      `json:"is_active"`
	MaxRetries          int       `json:"max_retries"`
	TimeoutMs           int       `json:"timeout_ms"`
	CircuitState        string    `json:"circuit_state"`
	ConsecutiveFailures int       `json:"consecutive_failures"`
	CreatedAt           time.Time `json:"created_at"`
	UpdatedAt           time.Time `json:"updated_at"`
}

func CreateEndpoint(ctx context.Context, pool *pgxpool.Pool, e *Endpoint) error {
	query := `
		INSERT INTO endpoints (user_id, url, secret, event_types, is_active, max_retries, timeout_ms)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id, circuit_state, consecutive_failures, created_at, updated_at`

	err := pool.QueryRow(ctx, query,
		e.UserID, e.URL, e.Secret, e.EventTypes, e.IsActive, e.MaxRetries, e.TimeoutMs,
	).Scan(&e.ID, &e.CircuitState, &e.ConsecutiveFailures, &e.CreatedAt, &e.UpdatedAt)
	if err != nil {
		return fmt.Errorf("CreateEndpoint: %w", err)
	}
	return nil
}

func GetEndpointByID(ctx context.Context, pool *pgxpool.Pool, id uuid.UUID) (*Endpoint, error) {
	query := `
		SELECT id, user_id, url, secret, event_types, is_active, max_retries, timeout_ms,
		       circuit_state, consecutive_failures, created_at, updated_at
		FROM endpoints WHERE id = $1`

	e := &Endpoint{}
	err := pool.QueryRow(ctx, query, id).Scan(
		&e.ID, &e.UserID, &e.URL, &e.Secret, &e.EventTypes, &e.IsActive,
		&e.MaxRetries, &e.TimeoutMs, &e.CircuitState, &e.ConsecutiveFailures,
		&e.CreatedAt, &e.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("GetEndpointByID: %w", err)
	}
	return e, nil
}

func ListEndpointsByUserID(ctx context.Context, pool *pgxpool.Pool, userID uuid.UUID) ([]Endpoint, error) {
	query := `
		SELECT id, user_id, url, secret, event_types, is_active, max_retries, timeout_ms,
		       circuit_state, consecutive_failures, created_at, updated_at
		FROM endpoints WHERE user_id = $1
		ORDER BY created_at DESC`

	rows, err := pool.Query(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("ListEndpointsByUserID: %w", err)
	}
	defer rows.Close()

	var endpoints []Endpoint
	for rows.Next() {
		var e Endpoint
		if err := rows.Scan(
			&e.ID, &e.UserID, &e.URL, &e.Secret, &e.EventTypes, &e.IsActive,
			&e.MaxRetries, &e.TimeoutMs, &e.CircuitState, &e.ConsecutiveFailures,
			&e.CreatedAt, &e.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("ListEndpointsByUserID scan: %w", err)
		}
		endpoints = append(endpoints, e)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("ListEndpointsByUserID rows: %w", err)
	}
	return endpoints, nil
}

func UpdateEndpoint(ctx context.Context, pool *pgxpool.Pool, e *Endpoint) error {
	query := `
		UPDATE endpoints
		SET url = $2, secret = $3, event_types = $4, is_active = $5,
		    max_retries = $6, timeout_ms = $7, circuit_state = $8,
		    consecutive_failures = $9, updated_at = now()
		WHERE id = $1
		RETURNING updated_at`

	err := pool.QueryRow(ctx, query,
		e.ID, e.URL, e.Secret, e.EventTypes, e.IsActive,
		e.MaxRetries, e.TimeoutMs, e.CircuitState, e.ConsecutiveFailures,
	).Scan(&e.UpdatedAt)
	if err != nil {
		return fmt.Errorf("UpdateEndpoint: %w", err)
	}
	return nil
}

func DeleteEndpoint(ctx context.Context, pool *pgxpool.Pool, id uuid.UUID) error {
	query := `DELETE FROM endpoints WHERE id = $1`
	result, err := pool.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("DeleteEndpoint: %w", err)
	}
	if result.RowsAffected() == 0 {
		return fmt.Errorf("DeleteEndpoint: endpoint not found")
	}
	return nil
}

func ListActiveEndpointsByEventType(ctx context.Context, pool *pgxpool.Pool, userID uuid.UUID, eventType string) ([]Endpoint, error) {
	query := `
		SELECT id, user_id, url, secret, event_types, is_active, max_retries, timeout_ms,
		       circuit_state, consecutive_failures, created_at, updated_at
		FROM endpoints
		WHERE user_id = $1 AND is_active = true AND $2 = ANY(event_types)
		ORDER BY created_at DESC`

	rows, err := pool.Query(ctx, query, userID, eventType)
	if err != nil {
		return nil, fmt.Errorf("ListActiveEndpointsByEventType: %w", err)
	}
	defer rows.Close()

	var endpoints []Endpoint
	for rows.Next() {
		var e Endpoint
		if err := rows.Scan(
			&e.ID, &e.UserID, &e.URL, &e.Secret, &e.EventTypes, &e.IsActive,
			&e.MaxRetries, &e.TimeoutMs, &e.CircuitState, &e.ConsecutiveFailures,
			&e.CreatedAt, &e.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("ListActiveEndpointsByEventType scan: %w", err)
		}
		endpoints = append(endpoints, e)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("ListActiveEndpointsByEventType rows: %w", err)
	}
	return endpoints, nil
}
