package retry

import (
	"math"
	"math/rand"
	"sync"
	"time"
)

// ExponentialBackoff implements an exponential backoff strategy with optional jitter.
// It is safe for concurrent use.
type ExponentialBackoff struct {
	mu     sync.Mutex
	min    time.Duration
	max    time.Duration
	factor float64
	jitter bool
}

// ExponentialBackoffConfig holds configuration for exponential backoff.
type ExponentialBackoffConfig struct {
	Min    time.Duration // Minimum delay
	Max    time.Duration // Maximum delay
	Factor float64       // Multiplier for exponential growth (typically 2.0)
	Jitter bool          // Add randomization to prevent thundering herd
}

// NewExponentialBackoff creates a new exponential backoff strategy.
// Min must be > 0, max must be >= min, and factor must be > 1.0.
func NewExponentialBackoff(config ExponentialBackoffConfig) *ExponentialBackoff {
	// Apply defaults and validation
	if config.Min <= 0 {
		config.Min = 1 * time.Second
	}
	if config.Max < config.Min {
		config.Max = config.Min
	}
	if config.Factor <= 1.0 {
		config.Factor = 2.0
	}

	return &ExponentialBackoff{
		min:    config.Min,
		max:    config.Max,
		factor: config.Factor,
		jitter: config.Jitter,
	}
}

// NewDefaultExponentialBackoff creates an exponential backoff with sensible defaults:
// - Min: 1s
// - Max: 30s
// - Factor: 2.0
// - Jitter: true
func NewDefaultExponentialBackoff() *ExponentialBackoff {
	return NewExponentialBackoff(ExponentialBackoffConfig{
		Min:    1 * time.Second,
		Max:    30 * time.Second,
		Factor: 2.0,
		Jitter: true,
	})
}

// NextDelay calculates the delay for the given attempt using exponential backoff.
// The calculation is: min * (factor ^ attempt), capped at max.
// If jitter is enabled, adds randomization: delay * [0.5, 1.5).
func (e *ExponentialBackoff) NextDelay(attempt int) time.Duration {
	e.mu.Lock()
	defer e.mu.Unlock()

	if attempt < 0 {
		attempt = 0
	}

	// Calculate exponential delay: min * (factor ^ attempt)
	delay := float64(e.min) * math.Pow(e.factor, float64(attempt))

	// Cap at maximum
	if delay > float64(e.max) {
		delay = float64(e.max)
	}

	// Apply jitter if enabled (randomize between 50% and 150% of delay)
	if e.jitter {
		//nolint:gosec // G404: math/rand acceptable for jitter (non-cryptographic use)
		delay *= (0.5 + rand.Float64())
	}

	return time.Duration(delay)
}

// Reset is a no-op for exponential backoff as it's stateless.
// Kept for interface compatibility.
func (e *ExponentialBackoff) Reset() {
	// Stateless strategy, nothing to reset
}

// ConstantBackoff implements a constant delay strategy.
// It is safe for concurrent use.
type ConstantBackoff struct {
	delay time.Duration
}

// NewConstantBackoff creates a backoff strategy with a fixed delay.
func NewConstantBackoff(delay time.Duration) *ConstantBackoff {
	if delay <= 0 {
		delay = 1 * time.Second
	}
	return &ConstantBackoff{delay: delay}
}

// NextDelay returns the constant delay regardless of attempt number.
func (c *ConstantBackoff) NextDelay(attempt int) time.Duration {
	return c.delay
}

// Reset is a no-op for constant backoff as it's stateless.
func (c *ConstantBackoff) Reset() {
	// Stateless strategy, nothing to reset
}

// LinearBackoff implements a linear backoff strategy.
// It is safe for concurrent use.
type LinearBackoff struct {
	mu        sync.Mutex
	increment time.Duration
	max       time.Duration
}

// NewLinearBackoff creates a linear backoff strategy.
// Delay increases by increment for each attempt, capped at max.
func NewLinearBackoff(increment, max time.Duration) *LinearBackoff {
	if increment <= 0 {
		increment = 1 * time.Second
	}
	if max < increment {
		max = increment
	}
	return &LinearBackoff{
		increment: increment,
		max:       max,
	}
}

// NextDelay calculates the delay as: increment * (attempt + 1), capped at max.
func (l *LinearBackoff) NextDelay(attempt int) time.Duration {
	l.mu.Lock()
	defer l.mu.Unlock()

	if attempt < 0 {
		attempt = 0
	}

	delay := l.increment * time.Duration(attempt+1)
	if delay > l.max {
		delay = l.max
	}
	return delay
}

// Reset is a no-op for linear backoff as it's stateless.
func (l *LinearBackoff) Reset() {
	// Stateless strategy, nothing to reset
}
