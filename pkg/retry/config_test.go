package retry

import (
	"os"
	"testing"
	"time"
)

func TestLoadConfig(t *testing.T) {
	origMaxAttempts := os.Getenv("RETRY_MAX_ATTEMPTS")
	origBackoffType := os.Getenv("RETRY_BACKOFF_TYPE")
	defer func() {
		os.Setenv("RETRY_MAX_ATTEMPTS", origMaxAttempts)
		os.Setenv("RETRY_BACKOFF_TYPE", origBackoffType)
	}()

	t.Run("loads defaults when no env vars set", func(t *testing.T) {
		os.Unsetenv("RETRY_MAX_ATTEMPTS")
		os.Unsetenv("RETRY_BACKOFF_TYPE")

		cfg := LoadConfig()

		if cfg.MaxAttempts != 3 {
			t.Errorf("expected max attempts 3, got %d", cfg.MaxAttempts)
		}
		if cfg.Backoff.Type != "exponential" {
			t.Errorf("expected backoff type 'exponential', got %s", cfg.Backoff.Type)
		}
		if cfg.Backoff.Min != 1*time.Second {
			t.Errorf("expected min 1s, got %v", cfg.Backoff.Min)
		}
		if cfg.Backoff.Max != 30*time.Second {
			t.Errorf("expected max 30s, got %v", cfg.Backoff.Max)
		}
		if cfg.Backoff.Factor != 2.0 {
			t.Errorf("expected factor 2.0, got %f", cfg.Backoff.Factor)
		}
		if !cfg.Backoff.Jitter {
			t.Error("expected jitter to be true")
		}
	})

	t.Run("loads from environment variables", func(t *testing.T) {
		os.Setenv("RETRY_MAX_ATTEMPTS", "5")
		os.Setenv("RETRY_BACKOFF_TYPE", "constant")
		os.Setenv("RETRY_BACKOFF_DELAY", "2s")

		cfg := LoadConfig()

		if cfg.MaxAttempts != 5 {
			t.Errorf("expected max attempts 5, got %d", cfg.MaxAttempts)
		}
		if cfg.Backoff.Type != "constant" {
			t.Errorf("expected backoff type 'constant', got %s", cfg.Backoff.Type)
		}
		if cfg.Backoff.Delay != 2*time.Second {
			t.Errorf("expected delay 2s, got %v", cfg.Backoff.Delay)
		}
	})
}

func TestBackoffConfig_CreateStrategy(t *testing.T) {
	t.Run("creates exponential backoff", func(t *testing.T) {
		bc := BackoffConfig{
			Type:   "exponential",
			Min:    500 * time.Millisecond,
			Max:    10 * time.Second,
			Factor: 2.5,
			Jitter: false,
		}

		strategy, err := bc.CreateStrategy()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if strategy == nil {
			t.Fatal("strategy should not be nil")
		}

		delay := strategy.NextDelay(0)
		if delay < 400*time.Millisecond || delay > 600*time.Millisecond {
			t.Errorf("expected delay around 500ms, got %v", delay)
		}
	})

	t.Run("creates constant backoff", func(t *testing.T) {
		bc := BackoffConfig{
			Type:  "constant",
			Delay: 3 * time.Second,
		}

		strategy, err := bc.CreateStrategy()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		delay := strategy.NextDelay(0)
		if delay != 3*time.Second {
			t.Errorf("expected delay 3s, got %v", delay)
		}
	})

	t.Run("creates linear backoff", func(t *testing.T) {
		bc := BackoffConfig{
			Type:      "linear",
			Increment: 2 * time.Second,
			Max:       20 * time.Second,
		}

		strategy, err := bc.CreateStrategy()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		delay := strategy.NextDelay(0)
		if delay != 2*time.Second {
			t.Errorf("expected first delay 2s, got %v", delay)
		}

		delay = strategy.NextDelay(1)
		if delay != 4*time.Second {
			t.Errorf("expected second delay 4s, got %v", delay)
		}
	})

	t.Run("returns error for unknown type", func(t *testing.T) {
		bc := BackoffConfig{
			Type: "unknown",
		}

		_, err := bc.CreateStrategy()
		if err == nil {
			t.Error("expected error for unknown backoff type")
		}
	})
}

func TestRetryConfig_ToConfig(t *testing.T) {
	t.Run("converts to Config successfully", func(t *testing.T) {
		rc := &RetryConfig{
			MaxAttempts: 5,
			Backoff: BackoffConfig{
				Type:  "constant",
				Delay: 1 * time.Second,
			},
		}

		cfg, err := rc.ToConfig()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if cfg.MaxAttempts != 5 {
			t.Errorf("expected max attempts 5, got %d", cfg.MaxAttempts)
		}

		if cfg.Strategy == nil {
			t.Fatal("strategy should not be nil")
		}

		if err := cfg.Validate(); err != nil {
			t.Errorf("config should be valid: %v", err)
		}
	})

	t.Run("returns error for invalid backoff type", func(t *testing.T) {
		rc := &RetryConfig{
			MaxAttempts: 3,
			Backoff: BackoffConfig{
				Type: "invalid",
			},
		}

		_, err := rc.ToConfig()
		if err == nil {
			t.Error("expected error for invalid backoff type")
		}
	})
}

func TestFindEnvFile(t *testing.T) {
	envFile := findEnvFile()
	_ = envFile
}
