package model

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Event struct {
	ID        uuid.UUID       `json:"id"`
	UserID    uuid.UUID       `json:"user_id"`
	EventType string          `json:"event_type"`
	Payload   json.RawMessage `json:"payload"`
	CreatedAt time.Time       `json:"created_at"`
}

func CreateEvent(ctx context.Context, pool *pgxpool.Pool, e *Event) error {
	query := `
		INSERT INTO events (user_id, event_type, payload)
		VALUES ($1, $2, $3)
		RETURNING id, created_at`

	err := pool.QueryRow(ctx, query,
		e.UserID, e.EventType, e.Payload,
	).Scan(&e.ID, &e.CreatedAt)
	if err != nil {
		return fmt.Errorf("CreateEvent: %w", err)
	}
	return nil
}

func GetEventByID(ctx context.Context, pool *pgxpool.Pool, id uuid.UUID) (*Event, error) {
	query := `
		SELECT id, user_id, event_type, payload, created_at
		FROM events WHERE id = $1`

	e := &Event{}
	err := pool.QueryRow(ctx, query, id).Scan(
		&e.ID, &e.UserID, &e.EventType, &e.Payload, &e.CreatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("GetEventByID: %w", err)
	}
	return e, nil
}

func ListEventsByUserID(ctx context.Context, pool *pgxpool.Pool, userID uuid.UUID, limit, offset int) ([]Event, error) {
	query := `
		SELECT id, user_id, event_type, payload, created_at
		FROM events WHERE user_id = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3`

	rows, err := pool.Query(ctx, query, userID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("ListEventsByUserID: %w", err)
	}
	defer rows.Close()

	var events []Event
	for rows.Next() {
		var e Event
		if err := rows.Scan(
			&e.ID, &e.UserID, &e.EventType, &e.Payload, &e.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("ListEventsByUserID scan: %w", err)
		}
		events = append(events, e)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("ListEventsByUserID rows: %w", err)
	}
	return events, nil
}
