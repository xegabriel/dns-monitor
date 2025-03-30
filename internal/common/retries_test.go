package common

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"
)

func TestRetryWithExponentialBackoff_SuccessOnFirstAttempt(t *testing.T) {
	var attempt int
	operation := func() error {
		attempt++
		return nil
	}
	ctx := context.Background()
	err := RetryWithExponentialBackoff(ctx, 3, 10*time.Millisecond, operation)
	if err != nil {
		t.Fatalf("Expected nil error, got %v", err)
	}
	if attempt != 1 {
		t.Fatalf("Expected 1 attempt, got %d", attempt)
	}
}

func TestRetryWithExponentialBackoff_EventualSuccess(t *testing.T) {
	var attempt int
	operation := func() error {
		attempt++
		if attempt < 3 {
			return errors.New("failure")
		}
		return nil
	}
	ctx := context.Background()
	start := time.Now()
	err := RetryWithExponentialBackoff(ctx, 5, 10*time.Millisecond, operation)
	if err != nil {
		t.Fatalf("Expected nil error, got %v", err)
	}
	if attempt != 3 {
		t.Fatalf("Expected 3 attempts, got %d", attempt)
	}
	elapsed := time.Since(start)
	// At minimum, the delay for the second attempt (10ms) and third attempt (20ms) should add up.
	if elapsed < 30*time.Millisecond {
		t.Fatalf("Expected elapsed time to be at least 30ms, got %v", elapsed)
	}
}

func TestRetryWithExponentialBackoff_AllFailures(t *testing.T) {
	var attempt int
	expectedErr := errors.New("permanent failure")
	operation := func() error {
		attempt++
		return expectedErr
	}
	ctx := context.Background()
	err := RetryWithExponentialBackoff(ctx, 3, 10*time.Millisecond, operation)
	if err == nil {
		t.Fatalf("Expected error, got nil")
	}
	if attempt != 3 {
		t.Fatalf("Expected 3 attempts, got %d", attempt)
	}
	expectedMsg := fmt.Sprintf("operation failed after %d attempts: %v", 3, expectedErr)
	if err.Error() != expectedMsg {
		t.Fatalf("Expected error message %q, got %q", expectedMsg, err.Error())
	}
}
