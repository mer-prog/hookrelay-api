package model

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type User struct {
	ID         uuid.UUID `json:"id"`
	Email      string    `json:"email"`
	Name       string    `json:"name"`
	AvatarURL  *string   `json:"avatar_url"`
	Provider   string    `json:"provider"`
	ProviderID string    `json:"provider_id"`
	Role       string    `json:"role"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

func CreateUser(ctx context.Context, pool *pgxpool.Pool, u *User) error {
	query := `
		INSERT INTO users (email, name, avatar_url, provider, provider_id, role)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id, created_at, updated_at`

	err := pool.QueryRow(ctx, query,
		u.Email, u.Name, u.AvatarURL, u.Provider, u.ProviderID, u.Role,
	).Scan(&u.ID, &u.CreatedAt, &u.UpdatedAt)
	if err != nil {
		return fmt.Errorf("CreateUser: %w", err)
	}
	return nil
}

func GetUserByID(ctx context.Context, pool *pgxpool.Pool, id uuid.UUID) (*User, error) {
	query := `
		SELECT id, email, name, avatar_url, provider, provider_id, role, created_at, updated_at
		FROM users WHERE id = $1`

	u := &User{}
	err := pool.QueryRow(ctx, query, id).Scan(
		&u.ID, &u.Email, &u.Name, &u.AvatarURL, &u.Provider, &u.ProviderID,
		&u.Role, &u.CreatedAt, &u.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("GetUserByID: %w", err)
	}
	return u, nil
}

func GetUserByEmail(ctx context.Context, pool *pgxpool.Pool, email string) (*User, error) {
	query := `
		SELECT id, email, name, avatar_url, provider, provider_id, role, created_at, updated_at
		FROM users WHERE email = $1`

	u := &User{}
	err := pool.QueryRow(ctx, query, email).Scan(
		&u.ID, &u.Email, &u.Name, &u.AvatarURL, &u.Provider, &u.ProviderID,
		&u.Role, &u.CreatedAt, &u.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("GetUserByEmail: %w", err)
	}
	return u, nil
}

func GetUserByProvider(ctx context.Context, pool *pgxpool.Pool, provider, providerID string) (*User, error) {
	query := `
		SELECT id, email, name, avatar_url, provider, provider_id, role, created_at, updated_at
		FROM users WHERE provider = $1 AND provider_id = $2`

	u := &User{}
	err := pool.QueryRow(ctx, query, provider, providerID).Scan(
		&u.ID, &u.Email, &u.Name, &u.AvatarURL, &u.Provider, &u.ProviderID,
		&u.Role, &u.CreatedAt, &u.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("GetUserByProvider: %w", err)
	}
	return u, nil
}
