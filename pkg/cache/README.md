# Cache Package

A robust, self-contained Redis cache implementation with automatic retry, connection pooling, and comprehensive error handling for Go applications.

## Features

- ✅ **Self-contained**: Zero dependencies on central config module
- ✅ **Environment-based configuration**: 12-factor app compliant
- ✅ **Automatic retry**: Built-in exponential backoff
- ✅ **Connection pooling**: Efficient resource management
- ✅ **Context-aware**: Respects cancellation and timeouts
- ✅ **Type-safe operations**: Set, Get, Delete, Exists, TTL
- ✅ **Health checks**: Monitor connection status
- ✅ **Structured logging**: slog integration
- ✅ **Comprehensive error handling**: Using fault package

## Installation

```bash
go get github.com/marcelofabianov/cache
```

## Quick Start

### Using Environment Variables

Create a `.env` file (see `.env.example`):

```env
CACHE_REDIS_HOST=localhost
CACHE_REDIS_PORT=6379
CACHE_REDIS_PASSWORD=
CACHE_REDIS_DB=0
CACHE_REDIS_POOL_MAX_IDLE_CONNS=10
CACHE_REDIS_POOL_MAX_ACTIVE_CONNS=20
```

Use in your code:

```go
package main

import (
    "context"
    "log/slog"
    "github.com/marcelofabianov/cache"
)

func main() {
    cfg, err := cache.LoadConfig()
    if err != nil {
        panic(err)
    }
    
    c, err := cache.New(cfg, slog.Default())
    if err != nil {
        panic(err)
    }
    defer c.Close()
    
    ctx := context.Background()
    
    if err := c.Set(ctx, "key", "value", 0); err != nil {
        panic(err)
    }
    
    value, err := c.Get(ctx, "key")
    if err != nil {
        panic(err)
    }
    
    println(value)
}
```

## Configuration

### Environment Variables

All variables use the `CACHE_` prefix:

| Variable | Type | Default | Description |
|----------|------|---------|-------------|
| `CACHE_REDIS_HOST` | string | localhost | Redis server host |
| `CACHE_REDIS_PORT` | int | 6379 | Redis server port |
| `CACHE_REDIS_PASSWORD` | string | "" | Redis password (optional) |
| `CACHE_REDIS_DB` | int | 0 | Redis database number |
| `CACHE_REDIS_CONNECT_QUERY_TIMEOUT` | duration | 2s | Query timeout |
| `CACHE_REDIS_CONNECT_EXEC_TIMEOUT` | duration | 2s | Exec timeout |
| `CACHE_REDIS_CONNECT_BACKOFF_MIN` | duration | 200ms | Min backoff delay |
| `CACHE_REDIS_CONNECT_BACKOFF_MAX` | duration | 15s | Max backoff delay |
| `CACHE_REDIS_CONNECT_BACKOFF_FACTOR` | int | 2 | Backoff growth factor |
| `CACHE_REDIS_CONNECT_BACKOFF_JITTER` | bool | true | Enable jitter |
| `CACHE_REDIS_CONNECT_BACKOFF_RETRIES` | int | 7 | Max retry attempts |
| `CACHE_REDIS_POOL_MAX_IDLE_CONNS` | int | 10 | Max idle connections |
| `CACHE_REDIS_POOL_MAX_ACTIVE_CONNS` | int | 20 | Max active connections |

## Operations

### Set

```go
err := c.Set(ctx, "user:123", userData, 1*time.Hour)
```

### Get

```go
value, err := c.Get(ctx, "user:123")
```

### Delete

```go
err := c.Delete(ctx, "user:123")
```

### Exists

```go
exists, err := c.Exists(ctx, "user:123")
```

### TTL

```go
ttl, err := c.TTL(ctx, "user:123")
```

### Health Check

```go
err := c.HealthCheck(ctx)
```

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
