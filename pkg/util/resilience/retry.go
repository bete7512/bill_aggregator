// pkg/util/resilience/retry.go
package resilience

import (
	"fmt"
	"math"
	mathrand "math/rand"
	"time"
)

// RetryPolicy defines a retry policy
type RetryPolicy struct {
	MaxRetries      int
	InitialDelay    time.Duration
	MaxDelay        time.Duration
	BackoffFactor   float64
	RetryableStatus []int
}

// NewRetryPolicy creates a new retry policy
func NewRetryPolicy(
	maxRetries int,
	initialDelay time.Duration,
	maxDelay time.Duration,
	backoffFactor float64,
	retryableStatus []int,
) *RetryPolicy {
	return &RetryPolicy{
		MaxRetries:      maxRetries,
		InitialDelay:    initialDelay,
		MaxDelay:        maxDelay,
		BackoffFactor:   backoffFactor,
		RetryableStatus: retryableStatus,
	}
}

// Execute executes a function with retry
func (p *RetryPolicy) Execute(fn func() error) error {
	var err error

	for attempt := 0; attempt <= p.MaxRetries; attempt++ {
		// Execute the function
		err = fn()
		if err == nil {
			return nil
		}

		// Check if we've exhausted retries
		if attempt >= p.MaxRetries {
			return fmt.Errorf("max retries exceeded: %w", err)
		}

		// Calculate delay using exponential backoff
		delay := p.calculateDelay(attempt)

		// Wait before the next attempt
		time.Sleep(delay)
	}

	return err
}

// calculateDelay calculates the delay for a retry attempt
func (p *RetryPolicy) calculateDelay(attempt int) time.Duration {
	// Calculate delay using exponential backoff
	delay := float64(p.InitialDelay) * math.Pow(p.BackoffFactor, float64(attempt))

	// Ensure delay doesn't exceed max delay
	if delay > float64(p.MaxDelay) {
		delay = float64(p.MaxDelay)
	}

	// Add jitter (±20%) using math/rand
	// Note: For truly random behavior across runs, seed math/rand once at application start (e.g., rand.Seed(time.Now().UnixNano()))
	// Go 1.20+ seeds the global generator automatically.
	jitterRange := delay * 0.2
	jitter := jitterRange * (2.0*mathrand.Float64() - 1.0) // Generates a value between -jitterRange and +jitterRange
	delay = delay + jitter

	// Ensure delay is not negative after jitter
	if delay < 0 {
		delay = 0
	}

	return time.Duration(delay)
}

// IsRetryableStatus checks if a status code is retryable
func (p *RetryPolicy) IsRetryableStatus(statusCode int) bool {
	for _, code := range p.RetryableStatus {
		if statusCode == code {
			return true
		}
	}
	return false
}
