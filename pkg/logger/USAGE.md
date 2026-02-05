# Logger Package - Self-Contained Configuration

Structured logger for Go based on `slog` with built-in Viper configuration support.

## ✅ Features

- ✅ **Self-contained**: Own configuration with Viper + .env
- ✅ **Zero external dependencies**: No config module needed
- ✅ **Environment-based**: Auto-detects format (JSON/Text) based on environment
- ✅ **Source tracking**: Automatic source location in development
- ✅ **Structured logging**: Key-value pairs via slog
- ✅ **Context support**: Trace IDs and request context
- ✅ **Performance**: Uses Go 1.21+ slog (zero allocations)

## 📦 Installation

```bash
go get github.com/marcelofabianov/logger
```

## 🚀 Quick Start

### 1. Create .env file

```bash
cp .env.example .env
```

### 2. Configure (optional)

```env
# .env
LOGGER_LEVEL=info
LOGGER_ENVIRONMENT=development
LOGGER_SERVICE_NAME=my-service
```

### 3. Use in your code

```go
package main

import (
    "github.com/marcelofabianov/logger"
)

func main() {
    // Load configuration from .env
    cfg, err := logger.LoadConfig()
    if err != nil {
        panic(err)
    }

    // Create logger
    log := logger.New(cfg)
    
    // Set as default (optional)
    log.SetDefault()

    // Use it!
    log.Info("Application started", 
        "version", "1.0.0",
        "port", 8080)
    
    log.Debug("Debug information", "details", "...")
    log.Warn("Warning message")
    log.Error("Error occurred", "error", err)
}
```

## 📋 Configuration Options

### Environment Variables

| Variable | Description | Default | Values |
|----------|-------------|---------|--------|
| `LOGGER_LEVEL` | Log level | `info` | `debug`, `info`, `warn`, `error` |
| `LOGGER_ENVIRONMENT` | Environment | `development` | `development`, `staging`, `production` |
| `LOGGER_SERVICE_NAME` | Service name | `app` | Any string |

### Behavior by Environment

| Environment | Format | Source Location | Use Case |
|-------------|--------|-----------------|----------|
| `development` | Text | ✅ Enabled | Local development |
| `staging` | JSON | ❌ Disabled | Pre-production |
| `production` | JSON | ❌ Disabled | Production |

## 🎯 Usage Patterns

### Pattern 1: Default Configuration (Recommended)

```go
cfg, _ := logger.LoadConfig()
log := logger.New(cfg)
```

**Reads from:**
1. `.env` file (if exists)
2. Environment variables (precedence)
3. Defaults (fallback)

### Pattern 2: Manual Configuration

```go
cfg := &logger.Config{
    Level:       logger.LevelDebug,
    Format:      logger.FormatJSON,
    Output:      os.Stdout,
    ServiceName: "my-service",
    Environment: "production",
    AddSource:   false,
    TimeFormat:  time.RFC3339,
}

log := logger.New(cfg)
```

### Pattern 3: With Context (Tracing)

```go
ctx := context.WithValue(context.Background(), "trace_id", "abc-123")

log.InfoContext(ctx, "Processing request",
    "user_id", 456,
    "action", "create")
```

### Pattern 4: Child Loggers

```go
// Add persistent fields
requestLogger := log.With(
    "request_id", "xyz-789",
    "user_id", 123,
)

requestLogger.Info("Request started")
requestLogger.Info("Request completed")
// All logs include request_id and user_id

// Group related fields
dbLogger := log.WithGroup("database")
dbLogger.Info("Query executed",
    "query", "SELECT * FROM users",
    "duration_ms", 42,
)
// Output: {database: {query: "...", duration_ms: 42}}
```

## 🔍 Log Levels

```go
log.Debug("Detailed debugging info")   // Only in debug level
log.Info("Informational message")      // General info
log.Warn("Warning condition")          // Warnings
log.Error("Error occurred")            // Errors
```

**Level Hierarchy:**
```
DEBUG < INFO < WARN < ERROR
```

If level is `INFO`:
- ✅ Info, Warn, Error are logged
- ❌ Debug is filtered out

## 📊 Output Formats

### Development (Text)

```
time=2026-02-05T12:00:00-03:00 level=INFO source=main.go:42 msg="User created" user_id=123 service=api environment=development
```

### Production (JSON)

```json
{
  "time": "2026-02-05T15:00:00Z",
  "level": "INFO",
  "msg": "User created",
  "user_id": 123,
  "service": "api",
  "environment": "production"
}
```

## 🧪 Testing

```bash
# Run tests
go test -v

# With coverage
go test -cover

# Race detection
go test -race
```

## 📁 Project Structure

```
pkg/logger/
├── config.go         # Configuration with Viper
├── logger.go         # Logger implementation
├── .env.example      # Example configuration
├── config_test.go    # Configuration tests
├── logger_test.go    # Logger tests
└── README.md         # This file
```

## 🎓 Design Principles

### 1. Self-Contained
- No external config module required
- All configuration managed internally
- Can be used standalone in any project

### 2. Convention over Configuration
- Sensible defaults for all values
- Auto-detection based on environment
- Minimal required configuration

### 3. Cloud-Native
- 12-factor app compliant
- Environment variables
- JSON output for production (log aggregation)
- Structured logging (searchable)

### 4. Developer Experience
- Human-readable text format in dev
- Source location for debugging
- Simple API (just 4 log methods)

## 🔧 Advanced Usage

### Custom Output

```go
var buf bytes.Buffer

cfg, _ := logger.LoadConfig()
cfg.Output = &buf  // Write to buffer instead of stdout

log := logger.New(cfg)
log.Info("test")

fmt.Println(buf.String())  // Get logs as string
```

### Custom Time Format

```go
cfg, _ := logger.LoadConfig()
cfg.TimeFormat = time.RFC822  // "02 Jan 06 15:04 MST"

log := logger.New(cfg)
```

### Integration with slog

```go
cfg, _ := logger.LoadConfig()
log := logger.New(cfg)

// Get underlying slog.Logger
slogLogger := log.Slog()

// Use slog API directly
slogLogger.LogAttrs(ctx, slog.LevelInfo, 
    "message",
    slog.String("key", "value"),
)
```

## 🚫 Anti-Patterns

### ❌ Don't create multiple logger instances

```go
// BAD
func handleRequest() {
    log := logger.New(cfg)  // Creates new instance
    log.Info("...")
}

// GOOD
var log *logger.Logger

func init() {
    cfg, _ := logger.LoadConfig()
    log = logger.New(cfg)
}

func handleRequest() {
    log.Info("...")  // Reuse singleton
}
```

### ❌ Don't log sensitive data

```go
// BAD
log.Info("User login", "password", password)

// GOOD
log.Info("User login", "user_id", userID)
```

### ❌ Don't use string formatting

```go
// BAD (loses structured data)
log.Info(fmt.Sprintf("User %d created", userID))

// GOOD (searchable, filterable)
log.Info("User created", "user_id", userID)
```

## 🔄 Migration from Other Loggers

### From log package

```go
// Before
log.Println("User created:", userID)

// After
logger.Info("User created", "user_id", userID)
```

### From logrus

```go
// Before
logrus.WithFields(logrus.Fields{
    "user_id": userID,
}).Info("User created")

// After
logger.Info("User created", "user_id", userID)
```

## 🤝 Contributing

This is a self-contained package designed for microservices.
Each package manages its own configuration independently.

## 📄 License

MIT License

---

**Version:** 1.0.0  
**Go Version:** 1.21+  
**Dependencies:** slog (stdlib), viper
