# Logger Package

**Self-contained structured logger** for Go with built-in configuration management.

## ⚡ Quick Start

```go
package main

import "github.com/marcelofabianov/logger"

func main() {
    // Load config from .env + environment variables
    cfg, _ := logger.LoadConfig()
    
    // Create logger
    log := logger.New(cfg)
    
    // Use it!
    log.Info("Hello, World!", "version", "1.0.0")
}
```

## ✨ Features

- ✅ **Self-contained**: Own config with Viper + .env (no external dependencies)
- ✅ **Zero setup**: Sensible defaults, works out-of-the-box
- ✅ **Environment-aware**: Auto JSON for prod, Text for dev
- ✅ **Structured**: Key-value pairs via slog
- ✅ **Context support**: Trace IDs and distributed tracing
- ✅ **Performance**: Go 1.21+ slog (zero allocations)

## 📦 Installation

```bash
go get github.com/marcelofabianov/logger
```

## 📖 Documentation

See [USAGE.md](./USAGE.md) for complete documentation.

## 🎯 Design Philosophy

### Self-Contained Package
Each `pkg/` in this monorepo is **completely independent**:
- ✅ Own configuration (Viper + .env)
- ✅ No dependency on central config module
- ✅ Can be extracted to separate repo
- ✅ Services choose which packages to use

### Microservices-Ready
- Each service loads only what it needs
- No coupling between packages
- Independent deployment
- Zero shared state

## 🔧 Configuration

### Option 1: Environment Variables (Recommended)

```bash
export LOGGER_LEVEL=debug
export LOGGER_ENVIRONMENT=development
export LOGGER_SERVICE_NAME=my-service
```

### Option 2: .env File

```bash
# .env
LOGGER_LEVEL=info
LOGGER_ENVIRONMENT=production
LOGGER_SERVICE_NAME=api-service
```

### Option 3: Manual

```go
cfg := &logger.Config{
    Level:       logger.LevelInfo,
    Format:      logger.FormatJSON,
    ServiceName: "my-service",
    Environment: "production",
}
log := logger.New(cfg)
```

## 📊 Configuration Reference

| Variable | Default | Values | Description |
|----------|---------|--------|-------------|
| `LOGGER_LEVEL` | `info` | `debug`, `info`, `warn`, `error` | Minimum log level |
| `LOGGER_ENVIRONMENT` | `development` | `development`, `staging`, `production` | Determines format and source tracking |
| `LOGGER_SERVICE_NAME` | `app` | Any string | Service identifier in logs |

## 🎨 Usage Examples

### Basic Logging

```go
log.Debug("Debug message")
log.Info("Info message")
log.Warn("Warning message")
log.Error("Error message")
```

### Structured Logging

```go
log.Info("User created",
    "user_id", 123,
    "email", "user@example.com",
    "role", "admin")
```

### With Context (Tracing)

```go
ctx := context.WithValue(context.Background(), "trace_id", "abc-123")
log.InfoContext(ctx, "Request processed", "duration_ms", 42)
```

### Child Loggers

```go
// Add persistent fields
requestLog := log.With("request_id", "xyz-789")
requestLog.Info("Request started")
requestLog.Info("Request completed")

// Group fields
dbLog := log.WithGroup("database")
dbLog.Info("Query executed", "duration_ms", 10)
```

## 🧪 Testing

```bash
go test -v
go test -cover
go test -race
```

## 📁 Files

```
pkg/logger/
├── config.go          # Configuration with Viper
├── logger.go          # Logger implementation  
├── .env.example       # Example environment file
├── config_test.go     # Config tests
├── logger_test.go     # Logger tests
├── USAGE.md           # Complete documentation
└── README.md          # This file
```

## 🏗️ Architecture

```
Your Service
     ↓
logger.LoadConfig()  ← Reads .env + env vars
     ↓
logger.New(cfg)
     ↓
log.Info(...)  → stdout (JSON or Text)
```

**Key Point:** No dependency on external config module!

## 🚀 For Microservices

Each microservice can:
1. Import just `pkg/logger`
2. Configure via environment variables
3. Deploy independently
4. No coupling with other services

```go
// service/course/main.go
import "github.com/marcelofabianov/logger"

func main() {
    cfg, _ := logger.LoadConfig()  // Reads LOGGER_* vars
    log := logger.New(cfg)
    
    log.Info("Course service starting")
}
```

## 📝 License

MIT

---

**Go Version:** 1.21+  
**Status:** ✅ Production Ready  
**Type:** Self-Contained Package
