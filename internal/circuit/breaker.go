package circuit

import (
	"sync"

	"github.com/google/uuid"
)

type State string

const (
	StateClosed   State = "CLOSED"
	StateOpen     State = "OPEN"
	StateHalfOpen State = "HALF_OPEN"

	DefaultThreshold = 10
)

// CircuitBreaker tracks consecutive failures for a single endpoint.
type CircuitBreaker struct {
	EndpointID          uuid.UUID
	state               State
	consecutiveFailures int
	threshold           int
	mu                  sync.Mutex
}

func NewBreaker(endpointID uuid.UUID, threshold int) *CircuitBreaker {
	if threshold <= 0 {
		threshold = DefaultThreshold
	}
	return &CircuitBreaker{
		EndpointID: endpointID,
		state:      StateClosed,
		threshold:  threshold,
	}
}

func (cb *CircuitBreaker) RecordSuccess() {
	cb.mu.Lock()
	defer cb.mu.Unlock()
	cb.consecutiveFailures = 0
	cb.state = StateClosed
}

func (cb *CircuitBreaker) RecordFailure() {
	cb.mu.Lock()
	defer cb.mu.Unlock()
	cb.consecutiveFailures++
	if cb.consecutiveFailures >= cb.threshold {
		cb.state = StateOpen
	}
}

func (cb *CircuitBreaker) IsOpen() bool {
	cb.mu.Lock()
	defer cb.mu.Unlock()
	return cb.state == StateOpen
}

// Reset transitions from OPEN to HALF_OPEN, allowing a single probe request.
func (cb *CircuitBreaker) Reset() {
	cb.mu.Lock()
	defer cb.mu.Unlock()
	if cb.state == StateOpen {
		cb.state = StateHalfOpen
	}
}

func (cb *CircuitBreaker) State() State {
	cb.mu.Lock()
	defer cb.mu.Unlock()
	return cb.state
}

func (cb *CircuitBreaker) ConsecutiveFailures() int {
	cb.mu.Lock()
	defer cb.mu.Unlock()
	return cb.consecutiveFailures
}

// BreakerManager holds circuit breakers keyed by endpoint ID.
type BreakerManager struct {
	breakers map[uuid.UUID]*CircuitBreaker
	mu       sync.Mutex
}

func NewBreakerManager() *BreakerManager {
	return &BreakerManager{
		breakers: make(map[uuid.UUID]*CircuitBreaker),
	}
}

// GetBreaker returns the breaker for the given endpoint, creating one if needed.
func (bm *BreakerManager) GetBreaker(endpointID uuid.UUID) *CircuitBreaker {
	bm.mu.Lock()
	defer bm.mu.Unlock()

	if cb, ok := bm.breakers[endpointID]; ok {
		return cb
	}
	cb := NewBreaker(endpointID, DefaultThreshold)
	bm.breakers[endpointID] = cb
	return cb
}
