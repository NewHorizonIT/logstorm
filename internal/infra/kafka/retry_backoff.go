package kafka

import (
	"context"
	"math/rand"
	"time"
)

type RetryCnf struct {
	MaxRetries int           // Number of retries
	BaseDelay  time.Duration // Time to wait before the first retry
	MaxDelay   time.Duration // Maximum delay between retries
}

// BackOffDelay calculates exponential backoff delay with jitter.
// Jitter adds ±25% randomization to prevent thundering herd.
func BackOffDelay(base time.Duration, retries int, maxDelay time.Duration) time.Duration {
	// Prevent overflow: 1<<retries overflows at retries > 62
	if retries > 30 {
		return addJitter(maxDelay)
	}

	delay := base * time.Duration(1<<retries)
	if delay > maxDelay {
		delay = maxDelay
	}

	return addJitter(delay)
}

func addJitter(delay time.Duration) time.Duration {
	// Jitter factor: [0.75, 1.25]
	jitterFactor := 0.75 + rand.Float64()*0.5
	return time.Duration(float64(delay) * jitterFactor)
}

// RetryWithBackoff executes fn with exponential backoff retry.
// Returns nil on success, or the last error after all retries exhausted.
func RetryWithBackoff(
	ctx context.Context,
	cfg RetryCnf,
	fn func() error,
) error {
	var err error

	for attempt := 0; attempt <= cfg.MaxRetries; attempt++ {
		err = fn()
		if err == nil {
			return nil
		}

		if attempt == cfg.MaxRetries {
			return err
		}

		delay := BackOffDelay(cfg.BaseDelay, attempt, cfg.MaxDelay)

		// Use NewTimer instead of time.After to avoid memory leak
		timer := time.NewTimer(delay)
		select {
		case <-timer.C:
			// Continue to next retry
		case <-ctx.Done():
			timer.Stop()
			return ctx.Err()
		}
	}

	return err
}
