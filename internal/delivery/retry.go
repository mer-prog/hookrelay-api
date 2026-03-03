package delivery

import (
	"context"
	"fmt"
	"math"
	"math/rand/v2"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/mer-prog/hookrelay-api/internal/model"
)

const baseDelay = 1 * time.Second

// CalculateBackoff returns 1s × 2^attempt plus random jitter (0–50% of base delay).
func CalculateBackoff(attempt int) time.Duration {
	base := baseDelay * time.Duration(math.Pow(2, float64(attempt)))
	jitter := time.Duration(rand.Int64N(int64(base) / 2))
	return base + jitter
}

// ShouldRetry decides whether a delivery with the given status code should be retried.
// 2xx → success (no retry). 400, 401, 403, 404 → permanent client errors (no retry). All others → retry.
func ShouldRetry(statusCode int) bool {
	if statusCode >= 200 && statusCode < 300 {
		return false
	}
	switch statusCode {
	case 400, 401, 403, 404:
		return false
	}
	return true
}

// ScheduleRetry updates the delivery log with the calculated next retry time.
func ScheduleRetry(ctx context.Context, pool *pgxpool.Pool, d *model.DeliveryLog) error {
	backoff := CalculateBackoff(d.AttemptNumber)
	nextRetry := time.Now().UTC().Add(backoff)
	d.Status = "PENDING"
	d.AttemptNumber++
	d.NextRetryAt = &nextRetry

	if err := model.UpdateDeliveryLog(ctx, pool, d); err != nil {
		return fmt.Errorf("ScheduleRetry: %w", err)
	}
	return nil
}
