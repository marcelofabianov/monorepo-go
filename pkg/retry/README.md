# Retry Package

A robust, self-contained retry mechanism with multiple backoff strategies for Go applications.

## Features

- ✅ **Self-contained**: Zero dependencies on central config module
- ✅ **Multiple backoff strategies**: Exponential, Constant, Linear
- ✅ **Environment-based configuration**: 12-factor app compliant
- ✅ **Context-aware**: Respects cancellation and timeouts
- ✅ **Flexible**: Programmatic or environment-driven config
- ✅ **Observable**: Optional callbacks and structured logging
- ✅ **Thread-safe**: Safe for concurrent use
- ✅ **Jitter support**: Prevents thundering herd problem

## Installation

```bash
go get github.com/marcelofabianov/retry
```

## Quick Start

### Using Environment Variables

Create a `.env` file (see `.env.example`):

```env
RETRY_MAX_ATTEMPTS=5
RETRY_BACKOFF_TYPE=exponential
RETRY_BACKOFF_MIN=500ms
RETRY_BACKOFF_MAX=30s
RETRY_BACKOFF_FACTOR=2.0
RETRY_BACKOFF_JITTER=true
```

Use in your code:

```go
package main

import (
    "context"
    "github.com/marcelofabianov/retry"
)

func main() {
    // Load config from environment
    retryCfg := retry.LoadConfig()
    
    // Convert to retry.Config
    cfg, err := retryCfg.ToConfig()
    if err != nil {
        panic(err)
    }
    
    // Use retry
    err = retry.Do(context.Background(), cfg, func(ctx context.Context) error {
        // Your operation here
        return callExternalAPI()
    })
}
```

### Using Programmatic Configuration

```go
cfg := &retry.Config{
    MaxAttempts: 5,
    Strategy:    retry.NewDefaultExponentialBackoff(),
}

err := retry.Do(ctx, cfg, func(ctx context.Context) error {
    return doSomething()
})
```

## Configuration

### Environment Variables

All variables use the `RETRY_` prefix:

| Variable | Type | Default | Description |
|----------|------|---------|-------------|
| `RETRY_MAX_ATTEMPTS` | int | 3 | Maximum retry attempts |
| `RETRY_BACKOFF_TYPE` | string | exponential | Backoff type: exponential, constant, linear |
| `RETRY_BACKOFF_MIN` | duration | 1s | Minimum delay (exponential) |
| `RETRY_BACKOFF_MAX` | duration | 30s | Maximum delay |
| `RETRY_BACKOFF_FACTOR` | float | 2.0 | Growth factor (exponential) |
| `RETRY_BACKOFF_JITTER` | bool | true | Enable jitter (exponential) |
| `RETRY_BACKOFF_DELAY` | duration | 1s | Fixed delay (constant) |
| `RETRY_BACKOFF_INCREMENT` | duration | 1s | Increment per attempt (linear) |

### Backoff Strategies

#### Exponential Backoff (Default)
```go
strategy := retry.NewExponentialBackoff(retry.ExponentialBackoffConfig{
    Min:    500 * time.Millisecond,
    Max:    30 * time.Second,
    Factor: 2.0,
    Jitter: true,
})
```

#### Constant Backoff
```go
strategy := retry.NewConstantBackoff(1 * time.Second)
```

#### Linear Backoff
```go
strategy := retry.NewLinearBackoff(1*time.Second, 10*time.Second)
```

## Advanced Usage

See [USAGE.md](USAGE.md) for:
- Custom strategies
- Callbacks and logging
- Error handling patterns
- Testing strategies
- Production examples

## Architecture

This package follows the **self-contained pattern** for microservices monorepos:

- ✅ No imports of central config module
- ✅ Independent configuration via environment variables
- ✅ Can be extracted to separate repository
- ✅ Zero coupling with other packages

## Testing

```bash
go test ./...
```

## License

MIT

## Contributing

Pull requests are welcome. For major changes, please open an issue first.
