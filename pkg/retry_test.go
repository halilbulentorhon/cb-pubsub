package util

import (
	"context"
	"errors"
	"testing"
	"time"
)

func TestDefaultRetryConfig(t *testing.T) {
	cfg := DefaultRetryConfig()

	expected := RetryConfig{
		MaxRetries:   3,
		InitialDelay: time.Second,
		MaxDelay:     8 * time.Second,
		Multiplier:   2.0,
	}

	if cfg.MaxRetries != expected.MaxRetries {
		t.Errorf("MaxRetries = %d, want %d", cfg.MaxRetries, expected.MaxRetries)
	}
	if cfg.InitialDelay != expected.InitialDelay {
		t.Errorf("InitialDelay = %v, want %v", cfg.InitialDelay, expected.InitialDelay)
	}
	if cfg.MaxDelay != expected.MaxDelay {
		t.Errorf("MaxDelay = %v, want %v", cfg.MaxDelay, expected.MaxDelay)
	}
	if cfg.Multiplier != expected.Multiplier {
		t.Errorf("Multiplier = %f, want %f", cfg.Multiplier, expected.Multiplier)
	}
}

func TestWithRetry_Success(t *testing.T) {
	cfg := RetryConfig{
		MaxRetries:   3,
		InitialDelay: 10 * time.Millisecond,
		MaxDelay:     100 * time.Millisecond,
		Multiplier:   2.0,
	}

	callCount := 0
	operation := func() error {
		callCount++
		return nil
	}

	err := WithRetry(context.Background(), cfg, operation)
	if err != nil {
		t.Errorf("WithRetry returned error: %v", err)
	}
	if callCount != 1 {
		t.Errorf("operation called %d times, want 1", callCount)
	}
}

func TestWithRetry_SuccessAfterRetries(t *testing.T) {
	cfg := RetryConfig{
		MaxRetries:   3,
		InitialDelay: 10 * time.Millisecond,
		MaxDelay:     100 * time.Millisecond,
		Multiplier:   2.0,
	}

	callCount := 0
	operation := func() error {
		callCount++
		if callCount < 3 {
			return errors.New("temporary error")
		}
		return nil
	}

	err := WithRetry(context.Background(), cfg, operation)
	if err != nil {
		t.Errorf("WithRetry returned error: %v", err)
	}
	if callCount != 3 {
		t.Errorf("operation called %d times, want 3", callCount)
	}
}

func TestWithRetry_FailureAfterAllRetries(t *testing.T) {
	cfg := RetryConfig{
		MaxRetries:   2,
		InitialDelay: 10 * time.Millisecond,
		MaxDelay:     100 * time.Millisecond,
		Multiplier:   2.0,
	}

	expectedErr := errors.New("persistent error")
	callCount := 0
	operation := func() error {
		callCount++
		return expectedErr
	}

	err := WithRetry(context.Background(), cfg, operation)
	if err != expectedErr {
		t.Errorf("WithRetry returned error: %v, want %v", err, expectedErr)
	}
	if callCount != 3 {
		t.Errorf("operation called %d times, want 3", callCount)
	}
}

func TestWithRetry_ContextCancellation(t *testing.T) {
	cfg := RetryConfig{
		MaxRetries:   5,
		InitialDelay: 100 * time.Millisecond,
		MaxDelay:     500 * time.Millisecond,
		Multiplier:   2.0,
	}

	ctx, cancel := context.WithCancel(context.Background())
	callCount := 0
	operation := func() error {
		callCount++
		if callCount == 2 {
			cancel()
		}
		return errors.New("test error")
	}

	err := WithRetry(ctx, cfg, operation)
	if !errors.Is(err, context.Canceled) {
		t.Errorf("WithRetry returned error: %v, want context.Canceled", err)
	}
	if callCount < 2 {
		t.Errorf("operation called %d times, should be at least 2", callCount)
	}
}

func TestWithRetry_DelayCalculation(t *testing.T) {
	cfg := RetryConfig{
		MaxRetries:   3,
		InitialDelay: 10 * time.Millisecond,
		MaxDelay:     50 * time.Millisecond,
		Multiplier:   3.0,
	}

	callCount := 0
	var delays []time.Duration
	startTime := time.Now()
	var lastCallTime time.Time

	operation := func() error {
		now := time.Now()
		if callCount > 0 {
			delay := now.Sub(lastCallTime)
			delays = append(delays, delay)
		}
		lastCallTime = now
		callCount++
		return errors.New("test error")
	}

	WithRetry(context.Background(), cfg, operation)

	if len(delays) != 3 {
		t.Errorf("Expected 3 delays, got %d", len(delays))
		return
	}

	expectedDelays := []time.Duration{
		10 * time.Millisecond,
		30 * time.Millisecond,
		50 * time.Millisecond,
	}

	for i, delay := range delays {
		if delay < expectedDelays[i] || delay > expectedDelays[i]+20*time.Millisecond {
			t.Errorf("Delay %d = %v, want approximately %v", i, delay, expectedDelays[i])
		}
	}

	totalTime := time.Since(startTime)
	expectedMinTime := 10*time.Millisecond + 30*time.Millisecond + 50*time.Millisecond
	if totalTime < expectedMinTime {
		t.Errorf("Total time %v is less than expected minimum %v", totalTime, expectedMinTime)
	}
}

func TestWithRetry_ZeroRetries(t *testing.T) {
	cfg := RetryConfig{
		MaxRetries:   0,
		InitialDelay: 10 * time.Millisecond,
		MaxDelay:     100 * time.Millisecond,
		Multiplier:   2.0,
	}

	callCount := 0
	expectedErr := errors.New("test error")
	operation := func() error {
		callCount++
		return expectedErr
	}

	err := WithRetry(context.Background(), cfg, operation)
	if err != expectedErr {
		t.Errorf("WithRetry returned error: %v, want %v", err, expectedErr)
	}
	if callCount != 1 {
		t.Errorf("operation called %d times, want 1", callCount)
	}
}
