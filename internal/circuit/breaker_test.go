package circuit

import (
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestCircuitBreaker(t *testing.T) {
	tests := []struct {
		name      string
		actions   func(cb *CircuitBreaker)
		wantState State
		wantFails int
	}{
		{
			name:      "starts CLOSED with zero failures",
			actions:   func(cb *CircuitBreaker) {},
			wantState: StateClosed,
			wantFails: 0,
		},
		{
			name: "stays CLOSED below threshold",
			actions: func(cb *CircuitBreaker) {
				for i := 0; i < 9; i++ {
					cb.RecordFailure()
				}
			},
			wantState: StateClosed,
			wantFails: 9,
		},
		{
			name: "transitions to OPEN at threshold",
			actions: func(cb *CircuitBreaker) {
				for i := 0; i < 10; i++ {
					cb.RecordFailure()
				}
			},
			wantState: StateOpen,
			wantFails: 10,
		},
		{
			name: "success resets to CLOSED",
			actions: func(cb *CircuitBreaker) {
				for i := 0; i < 10; i++ {
					cb.RecordFailure()
				}
				cb.RecordSuccess()
			},
			wantState: StateClosed,
			wantFails: 0,
		},
		{
			name: "Reset transitions OPEN to HALF_OPEN",
			actions: func(cb *CircuitBreaker) {
				for i := 0; i < 10; i++ {
					cb.RecordFailure()
				}
				cb.Reset()
			},
			wantState: StateHalfOpen,
			wantFails: 10,
		},
		{
			name: "HALF_OPEN success goes to CLOSED",
			actions: func(cb *CircuitBreaker) {
				for i := 0; i < 10; i++ {
					cb.RecordFailure()
				}
				cb.Reset()
				cb.RecordSuccess()
			},
			wantState: StateClosed,
			wantFails: 0,
		},
		{
			name: "HALF_OPEN failure goes back to OPEN",
			actions: func(cb *CircuitBreaker) {
				for i := 0; i < 10; i++ {
					cb.RecordFailure()
				}
				cb.Reset() // HALF_OPEN, failures still 10
				cb.RecordFailure()
			},
			wantState: StateOpen,
			wantFails: 11,
		},
		{
			name: "Reset on CLOSED is no-op",
			actions: func(cb *CircuitBreaker) {
				cb.Reset()
			},
			wantState: StateClosed,
			wantFails: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cb := NewBreaker(uuid.New(), DefaultThreshold)
			tt.actions(cb)
			assert.Equal(t, tt.wantState, cb.State())
			assert.Equal(t, tt.wantFails, cb.ConsecutiveFailures())
		})
	}
}

func TestBreakerManager_GetBreaker(t *testing.T) {
	bm := NewBreakerManager()
	id := uuid.New()

	cb1 := bm.GetBreaker(id)
	cb2 := bm.GetBreaker(id)
	assert.Same(t, cb1, cb2, "should return the same breaker instance")

	other := bm.GetBreaker(uuid.New())
	assert.NotSame(t, cb1, other, "different endpoint should get different breaker")
}
