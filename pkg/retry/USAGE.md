# Retry Package - Usage Guide

Complete guide for using the retry package in your applications.

## Table of Contents

- [Basic Usage](#basic-usage)
- [Configuration Methods](#configuration-methods)
- [Backoff Strategies](#backoff-strategies)
- [Advanced Features](#advanced-features)
- [Production Examples](#production-examples)
- [Testing](#testing)

## Basic Usage

### Simple Retry with Defaults

```go
package main

import (
    "context"
    "fmt"
    "github.com/marcelofabianov/retry"
)

func main() {
    cfg := &retry.Config{
        MaxAttempts: 3,
        Strategy:    retry.NewDefaultExponentialBackoff(),
    }

    err := retry.Do(context.Background(), cfg, func(ctx context.Context) error {
        // Your operation
        return callAPI()
    })

    if err != nil {
        fmt.Printf("Failed after retries: %v\n", err)
    }
}
```

### Using Environment Configuration

```go
package main

import (
    "context"
    "github.com/marcelofabianov/retry"
    "log/slog"
)

func main() {
    // Load from .env or environment variables
    retryCfg := retry.LoadConfig()
    
    // Convert to retry.Config
    cfg, err := retryCfg.ToConfig()
    if err != nil {
        panic(err)
    }
    
    // Add logger if needed
    cfg.Logger = slog.Default()
    
    // Use retry
    err = retry.Do(context.Background(), cfg, func(ctx context.Context) error {
        return doWork()
    })
}
```

## Configuration Methods

### Method 1: Environment Variables (Recommended)

Create `.env` file:

```env
RETRY_MAX_ATTEMPTS=5
RETRY_BACKOFF_TYPE=exponential
RETRY_BACKOFF_MIN=500ms
RETRY_BACKOFF_MAX=30s
RETRY_BACKOFF_FACTOR=2.0
RETRY_BACKOFF_JITTER=true
```

Load in code:

```go
cfg := retry.LoadConfig()
retryCfg, _ := cfg.ToConfig()
```

### Method 2: Programmatic Configuration

```go
cfg := &retry.Config{
    MaxAttempts: 5,
    Strategy: retry.NewExponentialBackoff(retry.ExponentialBackoffConfig{
        Min:    500 * time.Millisecond,
        Max:    30 * time.Second,
        Factor: 2.0,
        Jitter: true,
    }),
    Logger: slog.Default(),
}
```

### Method 3: Hybrid (Environment + Programmatic)

```go
retryCfg := retry.LoadConfig()
cfg, _ := retryCfg.ToConfig()

// Override with custom settings
cfg.Logger = customLogger
cfg.OnRetry = func(attempt int, err error) {
    metrics.IncrementRetryCounter(attempt)
}
```

## Backoff Strategies

### Exponential Backoff

Best for: External API calls, database connections

```go
// Default: 1s, 2s, 4s, 8s, 16s, 30s (capped)
strategy := retry.NewDefaultExponentialBackoff()

// Custom exponential
strategy := retry.NewExponentialBackoff(retry.ExponentialBackoffConfig{
    Min:    100 * time.Millisecond,  // Start small
    Max:    60 * time.Second,         // Cap at 1 minute
    Factor: 3.0,                      // Faster growth
    Jitter: true,                     // Randomize to prevent thundering herd
})
```

**With Jitter (Recommended for distributed systems):**
- Prevents multiple clients retrying simultaneously
- Adds randomization: delay * [0.5, 1.5)

### Constant Backoff

Best for: Rate limiting, fixed delays

```go
// Always wait 2 seconds between retries
strategy := retry.NewConstantBackoff(2 * time.Second)
```

**Use cases:**
- Rate-limited APIs with fixed windows
- Simple retry logic without complexity
- Testing scenarios

### Linear Backoff

Best for: Gradual backoff without exponential growth

```go
// 1s, 2s, 3s, 4s, 5s, 10s (capped)
strategy := retry.NewLinearBackoff(1*time.Second, 10*time.Second)
```

**Use cases:**
- Resource initialization
- Moderate retry pressure
- Predictable delay patterns

## Advanced Features

### With Logging

```go
cfg := &retry.Config{
    MaxAttempts: 5,
    Strategy:    retry.NewDefaultExponentialBackoff(),
    Logger:      slog.Default(),
}

// Logs will include:
// - Debug: Retry attempts, delays
// - Warn: All attempts failed
```

### With Callbacks (Metrics)

```go
cfg := &retry.Config{
    MaxAttempts: 5,
    Strategy:    retry.NewDefaultExponentialBackoff(),
    OnRetry: func(attempt int, err error) {
        // Send metrics
        metrics.Increment("retries.total", map[string]string{
            "attempt": fmt.Sprintf("%d", attempt),
            "error":   err.Error(),
        })
        
        // Alert on high retry count
        if attempt >= 3 {
            alerting.Notify("High retry count detected")
        }
    },
}
```

### With Context Timeout

```go
ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
defer cancel()

err := retry.Do(ctx, cfg, func(ctx context.Context) error {
    // This will respect the 30s timeout
    return callExternalService(ctx)
})

// If context times out, retry stops immediately
```

### Custom Retry Logic

```go
cfg := &retry.Config{
    MaxAttempts: 3,
    Strategy:    retry.NewDefaultExponentialBackoff(),
}

err := retry.Do(ctx, cfg, func(ctx context.Context) error {
    resp, err := http.Get("https://api.example.com/data")
    if err != nil {
        return err // Retry on network errors
    }
    defer resp.Body.Close()
    
    // Only retry on 5xx errors
    if resp.StatusCode >= 500 {
        return fmt.Errorf("server error: %d", resp.StatusCode)
    }
    
    // Don't retry on 4xx
    if resp.StatusCode >= 400 {
        return nil // Success (no retry)
    }
    
    return processResponse(resp)
})
```

## Production Examples

### HTTP Client with Retry

```go
type HTTPClient struct {
    client *http.Client
    retry  *retry.Config
}

func NewHTTPClient() *HTTPClient {
    retryCfg := retry.LoadConfig()
    cfg, _ := retryCfg.ToConfig()
    cfg.Logger = slog.Default()
    
    return &HTTPClient{
        client: &http.Client{Timeout: 10 * time.Second},
        retry:  cfg,
    }
}

func (c *HTTPClient) Get(ctx context.Context, url string) (*http.Response, error) {
    var resp *http.Response
    
    err := retry.Do(ctx, c.retry, func(ctx context.Context) error {
        req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
        if err != nil {
            return err
        }
        
        resp, err = c.client.Do(req)
        if err != nil {
            return err
        }
        
        // Retry on 5xx
        if resp.StatusCode >= 500 {
            resp.Body.Close()
            return fmt.Errorf("server error: %d", resp.StatusCode)
        }
        
        return nil
    })
    
    return resp, err
}
```

### Database Connection with Retry

```go
func ConnectDB(ctx context.Context) (*sql.DB, error) {
    retryCfg := retry.LoadConfig()
    cfg, _ := retryCfg.ToConfig()
    cfg.Logger = slog.Default()
    
    var db *sql.DB
    
    err := retry.Do(ctx, cfg, func(ctx context.Context) error {
        var err error
        db, err = sql.Open("postgres", connString)
        if err != nil {
            return err
        }
        
        // Verify connection
        if err := db.PingContext(ctx); err != nil {
            db.Close()
            return err
        }
        
        return nil
    })
    
    return db, err
}
```

### Message Queue Consumer

```go
func ConsumeMessages(ctx context.Context, queue Queue) error {
    retryCfg := retry.LoadConfig()
    cfg, _ := retryCfg.ToConfig()
    cfg.Logger = slog.Default()
    
    for {
        select {
        case <-ctx.Done():
            return ctx.Err()
        default:
        }
        
        msg, err := queue.Receive(ctx)
        if err != nil {
            return err
        }
        
        // Process with retry
        err = retry.Do(ctx, cfg, func(ctx context.Context) error {
            return processMessage(ctx, msg)
        })
        
        if err != nil {
            // Dead letter queue after all retries failed
            queue.SendToDLQ(msg, err)
        } else {
            // Acknowledge successful processing
            queue.Ack(msg)
        }
    }
}
```

### Circuit Breaker Pattern

```go
type CircuitBreaker struct {
    failures int
    maxFailures int
    resetTimeout time.Duration
    lastFailure time.Time
    retry *retry.Config
}

func (cb *CircuitBreaker) Call(ctx context.Context, fn func() error) error {
    // Check if circuit is open
    if cb.isOpen() {
        return fmt.Errorf("circuit breaker open")
    }
    
    // Try with retry
    err := retry.Do(ctx, cb.retry, func(ctx context.Context) error {
        return fn()
    })
    
    if err != nil {
        cb.recordFailure()
        return err
    }
    
    cb.reset()
    return nil
}
```

## Testing

### Testing with Retry

```go
func TestServiceWithRetry(t *testing.T) {
    // Use constant backoff for predictable tests
    cfg := &retry.Config{
        MaxAttempts: 3,
        Strategy:    retry.NewConstantBackoff(10 * time.Millisecond),
    }
    
    attempts := 0
    err := retry.Do(context.Background(), cfg, func(ctx context.Context) error {
        attempts++
        if attempts < 3 {
            return fmt.Errorf("temporary error")
        }
        return nil
    })
    
    if err != nil {
        t.Errorf("should succeed on 3rd attempt: %v", err)
    }
    
    if attempts != 3 {
        t.Errorf("expected 3 attempts, got %d", attempts)
    }
}
```

### Testing Environment Config

```go
func TestLoadConfig(t *testing.T) {
    os.Setenv("RETRY_MAX_ATTEMPTS", "5")
    os.Setenv("RETRY_BACKOFF_TYPE", "constant")
    defer func() {
        os.Unsetenv("RETRY_MAX_ATTEMPTS")
        os.Unsetenv("RETRY_BACKOFF_TYPE")
    }()
    
    cfg := retry.LoadConfig()
    
    if cfg.MaxAttempts != 5 {
        t.Errorf("expected 5 attempts, got %d", cfg.MaxAttempts)
    }
}
```

## Best Practices

1. **Use exponential backoff with jitter** for distributed systems
2. **Set reasonable max attempts** (3-5 for APIs, 10+ for critical operations)
3. **Always use context** for cancellation and timeouts
4. **Log retry attempts** for debugging and monitoring
5. **Add metrics** via OnRetry callback
6. **Test with constant backoff** for predictable timing
7. **Don't retry on 4xx errors** (client errors are permanent)
8. **Use circuit breakers** for cascading failure prevention

## Common Patterns

### Idempotent Operations Only
```go
// ✅ Good: Safe to retry
err := retry.Do(ctx, cfg, func(ctx context.Context) error {
    return fetchData(id) // GET operation
})

// ❌ Bad: Not idempotent
err := retry.Do(ctx, cfg, func(ctx context.Context) error {
    return createUser(user) // POST might create duplicates
})
```

### Conditional Retry
```go
err := retry.Do(ctx, cfg, func(ctx context.Context) error {
    err := operation()
    
    // Don't retry on permanent errors
    if errors.Is(err, ErrNotFound) {
        return nil // Treat as success
    }
    
    // Retry on transient errors
    if errors.Is(err, ErrTemporary) {
        return err
    }
    
    return err
})
```

## Environment Configuration Examples

### Development
```env
RETRY_MAX_ATTEMPTS=2
RETRY_BACKOFF_TYPE=constant
RETRY_BACKOFF_DELAY=100ms
```

### Production
```env
RETRY_MAX_ATTEMPTS=5
RETRY_BACKOFF_TYPE=exponential
RETRY_BACKOFF_MIN=1s
RETRY_BACKOFF_MAX=60s
RETRY_BACKOFF_FACTOR=2.0
RETRY_BACKOFF_JITTER=true
```

### Rate-Limited APIs
```env
RETRY_MAX_ATTEMPTS=10
RETRY_BACKOFF_TYPE=linear
RETRY_BACKOFF_INCREMENT=5s
RETRY_BACKOFF_MAX=60s
```
