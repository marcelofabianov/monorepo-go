# Database Package

A robust, self-contained PostgreSQL database wrapper with connection pooling, health checks, and comprehensive error handling for Go applications.

## Features

- ✅ **Self-contained**: Zero dependencies on central config module
- ✅ **Environment-based configuration**: 12-factor app compliant
- ✅ **Connection pooling**: Efficient resource management
- ✅ **Context-aware**: Respects cancellation and timeouts
- ✅ **Health checks**: Monitor connection status and pool statistics
- ✅ **Structured logging**: slog integration
- ✅ **Comprehensive error handling**: Using fault package
- ✅ **Transaction support**: BeginTx with options
- ✅ **pgx driver**: Modern PostgreSQL driver

## Installation

```bash
go get github.com/marcelofabianov/database
```

## Quick Start

### Using Environment Variables

Create a `.env` file (see `.env.example`):

```env
DATABASE_HOST=localhost
DATABASE_PORT=5432
DATABASE_USER=postgres
DATABASE_PASSWORD=secret
DATABASE_NAME=mydb
DATABASE_SSLMODE=disable
```

Use in your code:

```go
package main

import (
    "context"
    "log/slog"
    "github.com/marcelofabianov/database"
)

func main() {
    cfg, err := database.LoadConfig()
    if err != nil {
        panic(err)
    }
    
    db, err := database.New(cfg, slog.Default())
    if err != nil {
        panic(err)
    }
    defer db.Close()
    
    ctx := context.Background()
    if err := db.Connect(ctx); err != nil {
        panic(err)
    }
    
    rows, err := db.QueryContext(ctx, "SELECT * FROM users")
    if err != nil {
        panic(err)
    }
    defer rows.Close()
}
```

## Configuration

### Environment Variables

All variables use the `DATABASE_` prefix:

| Variable | Type | Default | Description |
|----------|------|---------|-------------|
| `DATABASE_HOST` | string | localhost | PostgreSQL server host |
| `DATABASE_PORT` | int | 5432 | PostgreSQL server port |
| `DATABASE_USER` | string | postgres | Database user |
| `DATABASE_PASSWORD` | string | "" | Database password |
| `DATABASE_NAME` | string | postgres | Database name |
| `DATABASE_SSLMODE` | string | disable | SSL mode (disable, require, verify-ca, verify-full) |
| `DATABASE_CONNECT_QUERY_TIMEOUT` | duration | 5s | Query timeout |
| `DATABASE_CONNECT_EXEC_TIMEOUT` | duration | 10s | Exec timeout |
| `DATABASE_CONNECT_BACKOFF_MIN` | duration | 500ms | Min backoff delay |
| `DATABASE_CONNECT_BACKOFF_MAX` | duration | 30s | Max backoff delay |
| `DATABASE_CONNECT_BACKOFF_FACTOR` | int | 2 | Backoff growth factor |
| `DATABASE_CONNECT_BACKOFF_JITTER` | bool | true | Enable jitter |
| `DATABASE_CONNECT_BACKOFF_RETRIES` | int | 5 | Max retry attempts |
| `DATABASE_POOL_MAX_OPEN_CONNS` | int | 25 | Max open connections |
| `DATABASE_POOL_MAX_IDLE_CONNS` | int | 5 | Max idle connections |
| `DATABASE_POOL_CONN_MAX_LIFETIME` | duration | 5m | Connection max lifetime |
| `DATABASE_POOL_CONN_MAX_IDLE_TIME` | duration | 5m | Connection max idle time |
| `DATABASE_POOL_HEALTH_CHECK_PERIOD` | duration | 30s | Health check interval |

## Operations

### Connect

```go
ctx := context.Background()
err := db.Connect(ctx)
```

### Query

```go
rows, err := db.QueryContext(ctx, "SELECT * FROM users WHERE id = $1", userID)
```

### Query Row

```go
row := db.QueryRowContext(ctx, "SELECT name FROM users WHERE id = $1", userID)
var name string
err := row.Scan(&name)
```

### Execute

```go
result, err := db.ExecContext(ctx, "UPDATE users SET name = $1 WHERE id = $2", name, userID)
```

### Transaction

```go
tx, err := db.BeginTx(ctx, nil)
if err != nil {
    return err
}
defer tx.Rollback()

_, err = tx.ExecContext(ctx, "INSERT INTO users (name) VALUES ($1)", name)
if err != nil {
    return err
}

return tx.Commit()
```

### Health Check

```go
err := db.HealthCheck(ctx)
```

### Background Health Check

```go
db.StartHealthCheckRoutine(ctx)
```

### Pool Statistics

```go
stats := db.Stats()
fmt.Printf("Open: %d, InUse: %d, Idle: %d\n", 
    stats.OpenConnections, stats.InUse, stats.Idle)
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
