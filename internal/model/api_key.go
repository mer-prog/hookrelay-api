package model

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type APIKey struct {
	ID         uuid.UUID  `json:"id"`
	UserID     uuid.UUID  `json:"user_id"`
	KeyHash    string     `json:"-"`
	Name       string     `json:"name"`
	Scopes     []string   `json:"scopes"`
	IsRevoked  bool       `json:"is_revoked"`
	LastUsedAt *time.Time `json:"last_used_at"`
	CreatedAt  time.Time  `json:"created_at"`
}

func CreateAPIKey(ctx context.Context, pool *pgxpool.Pool, k *APIKey) error {
	query := `
		INSERT INTO api_keys (user_id, key_hash, name, scopes)
		VALUES ($1, $2, $3, $4)
		RETURNING id, is_revoked, created_at`

	err := pool.QueryRow(ctx, query,
		k.UserID, k.KeyHash, k.Name, k.Scopes,
	).Scan(&k.ID, &k.IsRevoked, &k.CreatedAt)
	if err != nil {
		return fmt.Errorf("CreateAPIKey: %w", err)
	}
	return nil
}

func GetAPIKeyByHash(ctx context.Context, pool *pgxpool.Pool, keyHash string) (*APIKey, error) {
	query := `
		SELECT id, user_id, key_hash, name, scopes, is_revoked, last_used_at, created_at
		FROM api_keys WHERE key_hash = $1`

	k := &APIKey{}
	err := pool.QueryRow(ctx, query, keyHash).Scan(
		&k.ID, &k.UserID, &k.KeyHash, &k.Name, &k.Scopes,
		&k.IsRevoked, &k.LastUsedAt, &k.CreatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("GetAPIKeyByHash: %w", err)
	}
	return k, nil
}

func ListAPIKeysByUserID(ctx context.Context, pool *pgxpool.Pool, userID uuid.UUID) ([]APIKey, error) {
	query := `
		SELECT id, user_id, key_hash, name, scopes, is_revoked, last_used_at, created_at
		FROM api_keys WHERE user_id = $1
		ORDER BY created_at DESC`

	rows, err := pool.Query(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("ListAPIKeysByUserID: %w", err)
	}
	defer rows.Close()

	var keys []APIKey
	for rows.Next() {
		var k APIKey
		if err := rows.Scan(
			&k.ID, &k.UserID, &k.KeyHash, &k.Name, &k.Scopes,
			&k.IsRevoked, &k.LastUsedAt, &k.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("ListAPIKeysByUserID scan: %w", err)
		}
		keys = append(keys, k)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("ListAPIKeysByUserID rows: %w", err)
	}
	return keys, nil
}

func RevokeAPIKey(ctx context.Context, pool *pgxpool.Pool, id uuid.UUID) error {
	query := `UPDATE api_keys SET is_revoked = true WHERE id = $1`
	result, err := pool.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("RevokeAPIKey: %w", err)
	}
	if result.RowsAffected() == 0 {
		return fmt.Errorf("RevokeAPIKey: api key not found")
	}
	return nil
}

func UpdateLastUsed(ctx context.Context, pool *pgxpool.Pool, id uuid.UUID) error {
	query := `UPDATE api_keys SET last_used_at = now() WHERE id = $1`
	result, err := pool.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("UpdateLastUsed: %w", err)
	}
	if result.RowsAffected() == 0 {
		return fmt.Errorf("UpdateLastUsed: api key not found")
	}
	return nil
}
