package resilience

import (
	"fmt"
	"math"
	"time"
)

// Retrier implements exponential backoff retry logic
type Retrier struct {
	maxRetries int
	baseDelay  time.Duration
}

// NewRetrier creates a new retrier with default settings
func NewRetrier() *Retrier {
	return &Retrier{
		maxRetries: 3,
		baseDelay:  time.Second,
	}
}

// Do executes a function with retry logic
func (r *Retrier) Do(fn func() error) error {
	var err error
	
	for attempt := 0; attempt <= r.maxRetries; attempt++ {
		err = fn()
		if err == nil {
			return nil
		}

		// Don't sleep after the last attempt
		if attempt < r.maxRetries {
			delay := r.calculateDelay(attempt)
			time.Sleep(delay)
		}
	}

	return fmt.Errorf("failed after %d attempts: %w", r.maxRetries+1, err)
}

// calculateDelay calculates the delay for exponential backoff (2^attempt * baseDelay)
func (r *Retrier) calculateDelay(attempt int) time.Duration {
	multiplier := math.Pow(2, float64(attempt))
	delay := time.Duration(multiplier) * r.baseDelay
	
	// Cap at 30 seconds
	maxDelay := 30 * time.Second
	if delay > maxDelay {
		delay = maxDelay
	}
	
	return delay
}
