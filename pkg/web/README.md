# Web Package

A self-contained HTTP server foundation for building RESTful APIs in Go microservices. Designed to be used independently by multiple services in a monorepo.

## Features

- ✅ **Self-contained**: Zero dependencies on central config module
- ✅ **Environment-based configuration**: 12-factor app compliant
- ✅ **HTTP/HTTPS support**: TLS 1.2/1.3 with secure cipher suites
- ✅ **Graceful shutdown**: Proper context handling
- ✅ **CORS configuration**: Flexible cross-origin settings
- ✅ **Rate limiting**: Request throttling support
- ✅ **Health checks**: Liveness and readiness endpoints
- ✅ **Structured responses**: JSON response helpers
- ✅ **Structured logging**: slog integration

## Installation

```bash
go get github.com/marcelofabianov/web
```

## Quick Start

### Using Environment Variables

Create a `.env` file (see `.env.example`):

```env
WEB_HTTP_HOST=0.0.0.0
WEB_HTTP_PORT=8080
WEB_HTTP_CORS_ENABLED=true
```

Use in your microservice:

```go
package main

import (
    "context"
    "log/slog"
    "net/http"
    
    "github.com/marcelofabianov/web"
    "github.com/go-chi/chi/v5"
)

func main() {
    cfg, _ := web.LoadConfig()
    
    router := chi.NewRouter()
    router.Get("/health", web.LivenessHandler())
    router.Get("/api/users", handleGetUsers)
    
    server := web.NewServer(cfg, slog.Default(), router)
    
    if err := server.Start(); err != nil {
        panic(err)
    }
}
```

## Configuration

### Environment Variables

All variables use the `WEB_` prefix:

| Variable | Type | Default | Description |
|----------|------|---------|-------------|
| `WEB_HTTP_HOST` | string | 0.0.0.0 | Server bind address |
| `WEB_HTTP_PORT` | int | 8080 | Server port |
| `WEB_HTTP_READ_TIMEOUT` | duration | 15s | Read timeout |
| `WEB_HTTP_WRITE_TIMEOUT` | duration | 15s | Write timeout |
| `WEB_HTTP_IDLE_TIMEOUT` | duration | 60s | Idle timeout |
| `WEB_HTTP_TLS_ENABLED` | bool | false | Enable HTTPS |
| `WEB_HTTP_TLS_CERT_FILE` | string | "" | TLS certificate file |
| `WEB_HTTP_TLS_KEY_FILE` | string | "" | TLS key file |
| `WEB_HTTP_CORS_ENABLED` | bool | true | Enable CORS |
| `WEB_HTTP_CORS_ALLOWED_ORIGINS` | []string | * | Allowed origins |
| `WEB_HTTP_CORS_ALLOWED_METHODS` | []string | GET,POST,PUT... | Allowed methods |
| `WEB_HTTP_CORS_ALLOW_CREDENTIALS` | bool | true | Allow credentials |
| `WEB_HTTP_RATE_LIMIT_ENABLED` | bool | false | Enable rate limiting |
| `WEB_HTTP_RATE_LIMIT_REQUESTS_PER_SECOND` | int | 100 | Max requests/second |
| `WEB_HTTP_RATE_LIMIT_BURST` | int | 50 | Burst capacity |

## Server Operations

### Start Server

```go
server := web.NewServer(cfg, logger, router)
go server.Start()
```

### Graceful Shutdown

```go
ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
defer cancel()

if err := server.Shutdown(ctx); err != nil {
    log.Fatal(err)
}
```

### Get Server Address

```go
addr := server.Addr() // Returns "0.0.0.0:8080"
```

## Health Checks

### Liveness Probe

```go
router.Get("/health/live", web.LivenessHandler())
```

Returns HTTP 200 OK always. Used by orchestrators to check if service is running.

### Readiness Probe

```go
router.Get("/health/ready", web.ReadinessHandler(
    web.WithDatabaseCheck(db),
    web.WithCacheCheck(cache),
))
```

Returns HTTP 200 OK if all checks pass, 503 Service Unavailable otherwise.

## Response Helpers

```go
// Success response (HTTP 200)
web.Success(w, data)

// Created response (HTTP 201)
web.Created(w, data)

// No content (HTTP 204)
web.NoContent(w)

// Accepted (HTTP 202)
web.Accepted(w, data)

// Error response (HTTP 400/500)
web.Error(w, err)
```

## TLS/HTTPS Configuration

Enable HTTPS:

```env
WEB_HTTP_TLS_ENABLED=true
WEB_HTTP_TLS_CERT_FILE=/path/to/cert.pem
WEB_HTTP_TLS_KEY_FILE=/path/to/key.pem
```

Supported TLS versions: 1.2, 1.3

Secure cipher suites included by default.

## CORS Configuration

Production example:

```env
WEB_HTTP_CORS_ENABLED=true
WEB_HTTP_CORS_ALLOWED_ORIGINS=https://app.example.com,https://admin.example.com
WEB_HTTP_CORS_ALLOWED_METHODS=GET,POST,PUT,DELETE
WEB_HTTP_CORS_ALLOWED_HEADERS=Authorization,Content-Type
WEB_HTTP_CORS_ALLOW_CREDENTIALS=true
WEB_HTTP_CORS_MAX_AGE=3600
```

## Rate Limiting

Enable rate limiting:

```env
WEB_HTTP_RATE_LIMIT_ENABLED=true
WEB_HTTP_RATE_LIMIT_REQUESTS_PER_SECOND=100
WEB_HTTP_RATE_LIMIT_BURST=50
```

## Multi-Service Usage

Each microservice can have its own configuration:

```
service/
├── user-service/
│   ├── .env (WEB_HTTP_PORT=8081)
│   └── main.go
├── order-service/
│   ├── .env (WEB_HTTP_PORT=8082)
│   └── main.go
└── payment-service/
    ├── .env (WEB_HTTP_PORT=8083)
    └── main.go
```

Each service is completely independent with its own port, timeouts, and configuration.

## Architecture

This package follows the **self-contained pattern** for microservices monorepos:

- ✅ No imports of central config module
- ✅ Independent configuration via environment variables
- ✅ Can be extracted to separate repository
- ✅ Zero coupling with other packages
- ✅ Each microservice configures independently

## Testing

```bash
go test ./...
```

## License

MIT
