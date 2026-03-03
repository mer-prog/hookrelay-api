package model

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type DeliveryLog struct {
	ID              uuid.UUID       `json:"id"`
	EventID         uuid.UUID       `json:"event_id"`
	EndpointID      uuid.UUID       `json:"endpoint_id"`
	Status          string          `json:"status"`
	AttemptNumber   int             `json:"attempt_number"`
	RequestHeaders  json.RawMessage `json:"request_headers"`
	ResponseStatus  *int            `json:"response_status"`
	ResponseBody    *string         `json:"response_body"`
	ResponseHeaders json.RawMessage `json:"response_headers"`
	LatencyMs       *int            `json:"latency_ms"`
	ErrorMessage    *string         `json:"error_message"`
	NextRetryAt     *time.Time      `json:"next_retry_at"`
	CreatedAt       time.Time       `json:"created_at"`
}

func CreateDeliveryLog(ctx context.Context, pool *pgxpool.Pool, d *DeliveryLog) error {
	query := `
		INSERT INTO delivery_logs (event_id, endpoint_id, status, attempt_number, request_headers, latency_ms, next_retry_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id, created_at`

	err := pool.QueryRow(ctx, query,
		d.EventID, d.EndpointID, d.Status, d.AttemptNumber, d.RequestHeaders, d.LatencyMs, d.NextRetryAt,
	).Scan(&d.ID, &d.CreatedAt)
	if err != nil {
		return fmt.Errorf("CreateDeliveryLog: %w", err)
	}
	return nil
}

func UpdateDeliveryLog(ctx context.Context, pool *pgxpool.Pool, d *DeliveryLog) error {
	query := `
		UPDATE delivery_logs
		SET status = $2, attempt_number = $3, response_status = $4,
		    response_body = $5, response_headers = $6, latency_ms = $7,
		    error_message = $8, next_retry_at = $9
		WHERE id = $1`

	result, err := pool.Exec(ctx, query,
		d.ID, d.Status, d.AttemptNumber, d.ResponseStatus,
		d.ResponseBody, d.ResponseHeaders, d.LatencyMs,
		d.ErrorMessage, d.NextRetryAt,
	)
	if err != nil {
		return fmt.Errorf("UpdateDeliveryLog: %w", err)
	}
	if result.RowsAffected() == 0 {
		return fmt.Errorf("UpdateDeliveryLog: delivery log not found")
	}
	return nil
}

func GetDeliveryLogByID(ctx context.Context, pool *pgxpool.Pool, id uuid.UUID) (*DeliveryLog, error) {
	query := `
		SELECT id, event_id, endpoint_id, status, attempt_number, request_headers,
		       response_status, response_body, response_headers, latency_ms,
		       error_message, next_retry_at, created_at
		FROM delivery_logs WHERE id = $1`

	d := &DeliveryLog{}
	err := pool.QueryRow(ctx, query, id).Scan(
		&d.ID, &d.EventID, &d.EndpointID, &d.Status, &d.AttemptNumber,
		&d.RequestHeaders, &d.ResponseStatus, &d.ResponseBody, &d.ResponseHeaders,
		&d.LatencyMs, &d.ErrorMessage, &d.NextRetryAt, &d.CreatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("GetDeliveryLogByID: %w", err)
	}
	return d, nil
}

func ListDeliveryLogsByEndpointID(ctx context.Context, pool *pgxpool.Pool, endpointID uuid.UUID, limit, offset int) ([]DeliveryLog, error) {
	query := `
		SELECT id, event_id, endpoint_id, status, attempt_number, request_headers,
		       response_status, response_body, response_headers, latency_ms,
		       error_message, next_retry_at, created_at
		FROM delivery_logs WHERE endpoint_id = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3`

	rows, err := pool.Query(ctx, query, endpointID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("ListDeliveryLogsByEndpointID: %w", err)
	}
	defer rows.Close()

	var logs []DeliveryLog
	for rows.Next() {
		var d DeliveryLog
		if err := rows.Scan(
			&d.ID, &d.EventID, &d.EndpointID, &d.Status, &d.AttemptNumber,
			&d.RequestHeaders, &d.ResponseStatus, &d.ResponseBody, &d.ResponseHeaders,
			&d.LatencyMs, &d.ErrorMessage, &d.NextRetryAt, &d.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("ListDeliveryLogsByEndpointID scan: %w", err)
		}
		logs = append(logs, d)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("ListDeliveryLogsByEndpointID rows: %w", err)
	}
	return logs, nil
}

func ListDeliveryLogsByEventID(ctx context.Context, pool *pgxpool.Pool, eventID uuid.UUID) ([]DeliveryLog, error) {
	query := `
		SELECT id, event_id, endpoint_id, status, attempt_number, request_headers,
		       response_status, response_body, response_headers, latency_ms,
		       error_message, next_retry_at, created_at
		FROM delivery_logs WHERE event_id = $1
		ORDER BY created_at DESC`

	rows, err := pool.Query(ctx, query, eventID)
	if err != nil {
		return nil, fmt.Errorf("ListDeliveryLogsByEventID: %w", err)
	}
	defer rows.Close()

	var logs []DeliveryLog
	for rows.Next() {
		var d DeliveryLog
		if err := rows.Scan(
			&d.ID, &d.EventID, &d.EndpointID, &d.Status, &d.AttemptNumber,
			&d.RequestHeaders, &d.ResponseStatus, &d.ResponseBody, &d.ResponseHeaders,
			&d.LatencyMs, &d.ErrorMessage, &d.NextRetryAt, &d.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("ListDeliveryLogsByEventID scan: %w", err)
		}
		logs = append(logs, d)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("ListDeliveryLogsByEventID rows: %w", err)
	}
	return logs, nil
}

func ListPendingRetries(ctx context.Context, pool *pgxpool.Pool, now time.Time) ([]DeliveryLog, error) {
	query := `
		SELECT id, event_id, endpoint_id, status, attempt_number, request_headers,
		       response_status, response_body, response_headers, latency_ms,
		       error_message, next_retry_at, created_at
		FROM delivery_logs
		WHERE status = 'PENDING' AND next_retry_at <= $1
		ORDER BY next_retry_at ASC`

	rows, err := pool.Query(ctx, query, now)
	if err != nil {
		return nil, fmt.Errorf("ListPendingRetries: %w", err)
	}
	defer rows.Close()

	var logs []DeliveryLog
	for rows.Next() {
		var d DeliveryLog
		if err := rows.Scan(
			&d.ID, &d.EventID, &d.EndpointID, &d.Status, &d.AttemptNumber,
			&d.RequestHeaders, &d.ResponseStatus, &d.ResponseBody, &d.ResponseHeaders,
			&d.LatencyMs, &d.ErrorMessage, &d.NextRetryAt, &d.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("ListPendingRetries scan: %w", err)
		}
		logs = append(logs, d)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("ListPendingRetries rows: %w", err)
	}
	return logs, nil
}
