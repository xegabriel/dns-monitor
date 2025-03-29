package common

import (
	"context"
	"fmt"
	"log"
	"time"
)

// RetryWithExponentialBackoff retries the provided operation with an exponential backoff.
// It stops retrying when the operation succeeds or when the context is done.
func RetryWithExponentialBackoff(ctx context.Context, attempts int, initialDelay time.Duration, operation func() error) error {
	delay := initialDelay
	var lastErr error

	for i := 0; i < attempts; i++ {
		err := operation()
		if err == nil {
			if i != 0 { // Exclude the first attempt from logging
				log.Printf("Operation succeeded on attempt %d", i+1)
			}
			return nil
		}

		lastErr = err
		log.Printf("Attempt %d failed: %v. Retrying in %s...", i+1, err, delay)

		// Wait for the delay or return if context is cancelled.
		select {
		case <-ctx.Done():
			log.Printf("Context cancelled during retry wait: %v", ctx.Err())
			return ctx.Err()
		case <-time.After(delay):
		}

		delay *= 2 // Exponential backoff: double the delay.
	}

	log.Printf("Operation failed after %d attempts: %v", attempts, lastErr)
	return fmt.Errorf("operation failed after %d attempts: %w", attempts, lastErr)
}
