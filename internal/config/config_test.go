package config

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// setEnv sets environment variables from a map and returns a cleanup function.
func setEnv(t *testing.T, vars map[string]string) {
	t.Helper()
	for k, v := range vars {
		t.Setenv(k, v)
	}
}

// requiredEnv returns the minimum env vars needed for a valid config.
func requiredEnv() map[string]string {
	return map[string]string{
		"DATABASE_URL": "postgres://localhost:5432/test",
		"JWT_SECRET":   "test-secret",
	}
}

func TestLoad(t *testing.T) {
	tests := []struct {
		name    string
		env     map[string]string
		assert  func(t *testing.T, cfg *Config)
		wantErr string
	}{
		{
			name: "required fields only with defaults",
			env:  requiredEnv(),
			assert: func(t *testing.T, cfg *Config) {
				assert.Equal(t, "postgres://localhost:5432/test", cfg.DatabaseURL)
				assert.Equal(t, "test-secret", cfg.JWTSecret)
				assert.Equal(t, "8080", cfg.Port)
				assert.Equal(t, "memory", cfg.PubSubDriver)
				assert.Equal(t, 10, cfg.WorkerPoolSize)
				assert.Equal(t, "http://localhost:3000", cfg.FrontendURL)
				assert.Empty(t, cfg.GoogleClientID)
				assert.Empty(t, cfg.GitHubClientID)
				assert.Empty(t, cfg.RedisURL)
			},
		},
		{
			name: "all fields set",
			env: map[string]string{
				"DATABASE_URL":         "postgres://db:5432/prod",
				"JWT_SECRET":           "prod-secret",
				"PORT":                 "9090",
				"GOOGLE_CLIENT_ID":     "google-id",
				"GOOGLE_CLIENT_SECRET": "google-secret",
				"GITHUB_CLIENT_ID":     "github-id",
				"GITHUB_CLIENT_SECRET": "github-secret",
				"PUBSUB_DRIVER":        "redis",
				"REDIS_URL":            "redis://localhost:6379",
				"WORKER_POOL_SIZE":     "20",
				"FRONTEND_URL":         "https://app.example.com",
			},
			assert: func(t *testing.T, cfg *Config) {
				assert.Equal(t, "postgres://db:5432/prod", cfg.DatabaseURL)
				assert.Equal(t, "prod-secret", cfg.JWTSecret)
				assert.Equal(t, "9090", cfg.Port)
				assert.Equal(t, "google-id", cfg.GoogleClientID)
				assert.Equal(t, "google-secret", cfg.GoogleClientSecret)
				assert.Equal(t, "github-id", cfg.GitHubClientID)
				assert.Equal(t, "github-secret", cfg.GitHubClientSecret)
				assert.Equal(t, "redis", cfg.PubSubDriver)
				assert.Equal(t, "redis://localhost:6379", cfg.RedisURL)
				assert.Equal(t, 20, cfg.WorkerPoolSize)
				assert.Equal(t, "https://app.example.com", cfg.FrontendURL)
			},
		},
		{
			name:    "missing DATABASE_URL",
			env:     map[string]string{"JWT_SECRET": "s"},
			wantErr: "DATABASE_URL is required",
		},
		{
			name:    "missing JWT_SECRET",
			env:     map[string]string{"DATABASE_URL": "postgres://localhost/db"},
			wantErr: "JWT_SECRET is required",
		},
		{
			name: "invalid PUBSUB_DRIVER",
			env: map[string]string{
				"DATABASE_URL":  "postgres://localhost/db",
				"JWT_SECRET":    "s",
				"PUBSUB_DRIVER": "kafka",
			},
			wantErr: `PUBSUB_DRIVER must be "memory" or "redis"`,
		},
		{
			name: "redis driver without REDIS_URL",
			env: map[string]string{
				"DATABASE_URL":  "postgres://localhost/db",
				"JWT_SECRET":    "s",
				"PUBSUB_DRIVER": "redis",
			},
			wantErr: `REDIS_URL is required when PUBSUB_DRIVER is "redis"`,
		},
		{
			name: "invalid WORKER_POOL_SIZE",
			env: map[string]string{
				"DATABASE_URL":     "postgres://localhost/db",
				"JWT_SECRET":       "s",
				"WORKER_POOL_SIZE": "abc",
			},
			wantErr: "parsing WORKER_POOL_SIZE",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clear all config-related env vars so tests are isolated.
			for _, key := range []string{
				"DATABASE_URL", "JWT_SECRET", "PORT",
				"GOOGLE_CLIENT_ID", "GOOGLE_CLIENT_SECRET",
				"GITHUB_CLIENT_ID", "GITHUB_CLIENT_SECRET",
				"PUBSUB_DRIVER", "REDIS_URL",
				"WORKER_POOL_SIZE", "FRONTEND_URL",
			} {
				t.Setenv(key, "")
				os.Unsetenv(key)
			}

			setEnv(t, tt.env)

			cfg, err := Load()

			if tt.wantErr != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.wantErr)
				assert.Nil(t, cfg)
				return
			}

			require.NoError(t, err)
			require.NotNil(t, cfg)
			tt.assert(t, cfg)
		})
	}
}
