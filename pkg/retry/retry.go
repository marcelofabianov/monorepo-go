package retry

import (
	"context"
	"log/slog"
	"time"

	"github.com/marcelofabianov/fault"
)

var (
	// ErrMaxAttemptsReached is returned when the maximum number of retry attempts is exceeded.
	ErrMaxAttemptsReached = fault.New(
		"maximum retry attempts reached",
		fault.WithCode(fault.Invalid),
	)

	// ErrInvalidConfig is returned when retry configuration is invalid.
	ErrInvalidConfig = fault.New(
		"invalid retry configuration",
		fault.WithCode(fault.Invalid),
	)
)

// Strategy defines the interface for calculating retry delays.
// Implementations must be safe for concurrent use.
type Strategy interface {
	// NextDelay calculates the delay duration for the given attempt number.
	// Attempt numbers start at 0 for the first retry.
	NextDelay(attempt int) time.Duration

	// Reset resets the strategy state to initial values.
	Reset()
}

// RetryableFunc represents a function that can be retried.
// It returns an error to indicate whether the operation should be retried.
type RetryableFunc func(ctx context.Context) error

// Config holds the retry configuration.
type Config struct {
	// MaxAttempts is the maximum number of retry attempts (0 means no retries).
	MaxAttempts int

	// Strategy defines how retry delays are calculated.
	Strategy Strategy

	// OnRetry is called before each retry attempt.
	// The attempt parameter starts at 0 for the first retry.
	OnRetry func(attempt int, err error)

	// Logger for retry operations. If nil, uses slog.Default().
	Logger *slog.Logger
}

// Validate checks if the retry configuration is valid.
func (c *Config) Validate() error {
	if c.MaxAttempts < 0 {
		return fault.Wrap(ErrInvalidConfig, "max attempts must be non-negative",
			fault.WithContext("max_attempts", c.MaxAttempts),
		)
	}
	if c.Strategy == nil {
		return fault.Wrap(ErrInvalidConfig, "strategy cannot be nil")
	}
	return nil
}

// Do executes the given function with retries according to the configuration.
// It returns the last error encountered if all attempts fail.
func Do(ctx context.Context, config *Config, fn RetryableFunc) error {
	if err := config.Validate(); err != nil {
		return err
	}

	logger := config.Logger
	if logger == nil {
		logger = slog.Default()
	}

	err := fn(ctx)
	if err == nil {
		return nil
	}

	if config.MaxAttempts == 0 {
		return err
	}

	logger.Debug("Starting retry attempts",
		"max_attempts", config.MaxAttempts,
		"error", err.Error(),
	)

	for attempt := 0; attempt < config.MaxAttempts; attempt++ {
		if ctx.Err() != nil {
			return fault.Wrap(ctx.Err(), "context cancelled during retry",
				fault.WithContext("attempt", attempt),
				fault.WithContext("max_attempts", config.MaxAttempts),
			)
		}

		if config.OnRetry != nil {
			config.OnRetry(attempt, err)
		}

		delay := config.Strategy.NextDelay(attempt)

		logger.Debug("Retrying after delay",
			"attempt", attempt+1,
			"max_attempts", config.MaxAttempts,
			"delay_ms", delay.Milliseconds(),
		)

		select {
		case <-ctx.Done():
			return fault.Wrap(ctx.Err(), "context cancelled during retry delay",
				fault.WithContext("attempt", attempt),
				fault.WithContext("max_attempts", config.MaxAttempts),
			)
		case <-time.After(delay):
		}

		err = fn(ctx)
		if err == nil {
			logger.Debug("Retry succeeded",
				"attempt", attempt+1,
				"total_attempts", attempt+2,
			)
			return nil
		}
	}

	logger.Warn("All retry attempts failed",
		"max_attempts", config.MaxAttempts,
		"error", err.Error(),
	)

	return fault.Wrap(ErrMaxAttemptsReached, "all retry attempts failed",
		fault.WithContext("attempts", config.MaxAttempts),
		fault.WithWrappedErr(err),
	)
}
