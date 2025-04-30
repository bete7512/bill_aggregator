// pkg/util/resilience/circuitbreaker.go
package resilience

import (
	"errors"
	"sync"
	"time"
)

// CircuitBreaker states
const (
	StateClosed   = "CLOSED"
	StateOpen     = "OPEN"
	StateHalfOpen = "HALF_OPEN"
)

// CircuitBreaker implements the circuit breaker pattern
type CircuitBreaker struct {
	state              string
	failureThreshold   int
	resetTimeout       time.Duration
	maxConcurrentCalls int
	
	failures           int
	lastFailureTime    time.Time
	activeCalls        int
	
	mutex              sync.RWMutex
}

// NewCircuitBreaker creates a new circuit breaker
func NewCircuitBreaker(
	failureThreshold int,
	resetTimeout time.Duration,
	maxConcurrentCalls int,
) *CircuitBreaker {
	return &CircuitBreaker{
		state:              StateClosed,
		failureThreshold:   failureThreshold,
		resetTimeout:       resetTimeout,
		maxConcurrentCalls: maxConcurrentCalls,
		failures:           0,
		lastFailureTime:    time.Time{},
		activeCalls:        0,
	}
}

// Execute executes a function with circuit breaker protection
func (cb *CircuitBreaker) Execute(fn func() error) error {
	// Check if circuit is open
	cb.mutex.RLock()
	if cb.state == StateOpen {
		// Check if reset timeout has elapsed
		if time.Since(cb.lastFailureTime) > cb.resetTimeout {
			cb.mutex.RUnlock()
			cb.mutex.Lock()
			cb.state = StateHalfOpen
			cb.mutex.Unlock()
		} else {
			cb.mutex.RUnlock()
			return errors.New("circuit breaker is open")
		}
	} else {
		cb.mutex.RUnlock()
	}

	// Acquire permission to execute
	cb.mutex.Lock()
	if cb.activeCalls >= cb.maxConcurrentCalls {
		cb.mutex.Unlock()
		return errors.New("too many concurrent calls")
	}
	cb.activeCalls++
	cb.mutex.Unlock()

	// Ensure we decrement activeCalls when done
	defer func() {
		cb.mutex.Lock()
		cb.activeCalls--
		cb.mutex.Unlock()
	}()

	// Execute the function
	err := fn()

	// Handle the result
	if err != nil {
		cb.recordFailure()
		return err
	}

	// If successful and in half-open state, reset the circuit
	if cb.state == StateHalfOpen {
		cb.reset()
	}

	return nil
}

// recordFailure records a failure and potentially opens the circuit
func (cb *CircuitBreaker) recordFailure() {
	cb.mutex.Lock()
	defer cb.mutex.Unlock()

	cb.failures++
	cb.lastFailureTime = time.Now()

	// If failures exceed threshold, open the circuit
	if cb.failures >= cb.failureThreshold {
		cb.state = StateOpen
	}
}

// reset resets the circuit breaker to closed state
func (cb *CircuitBreaker) reset() {
	cb.mutex.Lock()
	defer cb.mutex.Unlock()

	cb.failures = 0
	cb.state = StateClosed
}

// GetState returns the current state of the circuit breaker
func (cb *CircuitBreaker) GetState() string {
	cb.mutex.RLock()
	defer cb.mutex.RUnlock()
	return cb.state
}