package delivery

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestCalculateBackoff(t *testing.T) {
	tests := []struct {
		name    string
		attempt int
		minWait time.Duration
		maxWait time.Duration
	}{
		{
			name:    "attempt 0 → ~1s",
			attempt: 0,
			minWait: 1 * time.Second,
			maxWait: 1500 * time.Millisecond,
		},
		{
			name:    "attempt 1 → ~2s",
			attempt: 1,
			minWait: 2 * time.Second,
			maxWait: 3 * time.Second,
		},
		{
			name:    "attempt 2 → ~4s",
			attempt: 2,
			minWait: 4 * time.Second,
			maxWait: 6 * time.Second,
		},
		{
			name:    "attempt 3 → ~8s",
			attempt: 3,
			minWait: 8 * time.Second,
			maxWait: 12 * time.Second,
		},
		{
			name:    "attempt 4 → ~16s",
			attempt: 4,
			minWait: 16 * time.Second,
			maxWait: 24 * time.Second,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := CalculateBackoff(tt.attempt)
			assert.GreaterOrEqual(t, d, tt.minWait)
			assert.LessOrEqual(t, d, tt.maxWait)
		})
	}
}

func TestShouldRetry(t *testing.T) {
	tests := []struct {
		name       string
		statusCode int
		want       bool
	}{
		{"200 OK", 200, false},
		{"201 Created", 201, false},
		{"204 No Content", 204, false},
		{"400 Bad Request", 400, false},
		{"401 Unauthorized", 401, false},
		{"403 Forbidden", 403, false},
		{"404 Not Found", 404, false},
		{"408 Request Timeout", 408, true},
		{"429 Too Many Requests", 429, true},
		{"500 Internal Server Error", 500, true},
		{"502 Bad Gateway", 502, true},
		{"503 Service Unavailable", 503, true},
		{"504 Gateway Timeout", 504, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, ShouldRetry(tt.statusCode))
		})
	}
}
