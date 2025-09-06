package util

import (
	"context"
	"log/slog"
	"time"
)

type RetryConfig struct {
	MaxRetries   int           `json:"maxRetries"`
	InitialDelay time.Duration `json:"initialDelay"`
	MaxDelay     time.Duration `json:"maxDelay"`
	Multiplier   float64       `json:"multiplier"`
}

func DefaultRetryConfig() RetryConfig {
	return RetryConfig{
		MaxRetries:   3,
		InitialDelay: time.Second,
		MaxDelay:     8 * time.Second,
		Multiplier:   2.0,
	}
}

func WithRetry(ctx context.Context, cfg RetryConfig, operation func() error) error {
	var lastErr error
	delay := cfg.InitialDelay

	for attempt := 0; attempt <= cfg.MaxRetries; attempt++ {
		if attempt > 0 {
			slog.Debug("retrying operation",
				"attempt", attempt,
				"delay", delay,
				"last_error", lastErr.Error())

			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(delay):
			}

			delay = time.Duration(float64(delay) * cfg.Multiplier)
			if delay > cfg.MaxDelay {
				delay = cfg.MaxDelay
			}
		}

		if err := operation(); err != nil {
			lastErr = err

			if attempt == cfg.MaxRetries {
				slog.Error("operation failed after all retries",
					"attempts", attempt+1,
					"final_error", err)
			}
			continue
		}

		if attempt > 0 {
			slog.Info("operation succeeded after retry",
				"attempts", attempt+1)
		}
		return nil
	}

	return lastErr
}
