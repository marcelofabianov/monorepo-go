# 🌐 pkg/web - HTTP Server & Utilities

Self-contained HTTP server package for microservices with Chi Router.

## 📦 O que está incluído?

### Core
- **config.go** - Configuration with Viper (WEB_* env vars)
- **server.go** - HTTP/HTTPS server with graceful shutdown
- **health.go** - Health check endpoints (liveness/readiness)
- **response.go** - JSON response helpers

### Middlewares
- **middleware/** - 14 security-first middlewares
  - Ver: [middleware/USAGE.md](./middleware/USAGE.md)

## 🚀 Quick Start

### 1. Configuração

```bash
# .env
WEB_HTTP_PORT=8080
WEB_HTTP_READ_TIMEOUT=10s
WEB_HTTP_WRITE_TIMEOUT=10s
WEB_HTTP_MAX_BODY_SIZE=10485760

# TLS (opcional)
WEB_TLS_ENABLED=false
WEB_TLS_CERT_FILE=./certs/server.crt
WEB_TLS_KEY_FILE=./certs/server.key

# CORS
WEB_CORS_ENABLED=true
WEB_CORS_ALLOWED_ORIGINS=https://app.example.com,https://admin.example.com
WEB_CORS_ALLOWED_METHODS=GET,POST,PUT,DELETE,OPTIONS
WEB_CORS_ALLOWED_HEADERS=Accept,Authorization,Content-Type
WEB_CORS_MAX_AGE=300

# Rate Limiting (exemplo)
WEB_RATE_LIMIT_ENABLED=true
WEB_RATE_LIMIT_GLOBAL_LIMIT=100
WEB_RATE_LIMIT_GLOBAL_WINDOW=1m
```

### 2. Criar Servidor Simples

```go
package main

import (
    "log/slog"
    "net/http"
    "os"

    "github.com/go-chi/chi/v5"
    "github.com/marcelofabianov/web"
)

func main() {
    // Logger
    logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
    
    // Load config
    cfg, err := web.LoadConfig()
    if err != nil {
        logger.Error("failed to load config", "error", err)
        os.Exit(1)
    }
    
    // Router
    r := chi.NewRouter()
    
    // Routes
    r.Get("/", func(w http.ResponseWriter, r *http.Request) {
        web.Success(w, r, map[string]string{
            "message": "Hello, World!",
        })
    })
    
    // Health checks
    r.Get("/health", web.LivenessHandler)
    r.Get("/health/ready", web.ReadinessHandler())
    
    // Start server
    srv := web.NewServer(cfg, logger, r)
    if err := srv.Start(); err != nil {
        logger.Error("server error", "error", err)
        os.Exit(1)
    }
}
```

### 3. Servidor com Middlewares Security-First

```go
package main

import (
    "log/slog"
    "net/http"
    "os"
    "time"

    "github.com/go-chi/chi/v5"
    chimiddleware "github.com/go-chi/chi/v5/middleware"
    "github.com/redis/go-redis/v9"
    
    "github.com/marcelofabianov/web"
    "github.com/marcelofabianov/web/middleware"
)

func main() {
    // Logger
    logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
        Level: slog.LevelInfo,
    }))
    
    // Config
    cfg, _ := web.LoadConfig()
    
    // Redis (para rate limiting)
    redisClient := redis.NewClient(&redis.Options{
        Addr: "localhost:6379",
    })
    
    // Security Logger
    secLogger := middleware.NewSecurityLogger(logger)
    
    // Router
    r := chi.NewRouter()
    
    // ============================================
    // LAYER 1: Observability
    // ============================================
    r.Use(middleware.RequestID())
    r.Use(middleware.RealIP())
    r.Use(middleware.Recovery(logger))
    r.Use(middleware.Logger(logger))
    
    // ============================================
    // LAYER 2: Security
    // ============================================
    r.Use(middleware.SecurityHeaders(middleware.SecurityHeadersConfig{
        XContentTypeOptions:      "nosniff",
        XFrameOptions:            "DENY",
        ContentSecurityPolicy:    "default-src 'self'",
        ReferrerPolicy:           "no-referrer",
        StrictTransportSecurity:  "max-age=31536000; includeSubDomains",
    }))
    
    r.Use(middleware.HTTPSOnly(middleware.HTTPSOnlyConfig{
        Enabled: cfg.TLS.Enabled,
    }))
    
    r.Use(middleware.CORS(middleware.CORSConfig{
        AllowedOrigins:   cfg.CORS.AllowedOrigins,
        AllowedMethods:   cfg.CORS.AllowedMethods,
        AllowedHeaders:   cfg.CORS.AllowedHeaders,
        AllowCredentials: cfg.CORS.AllowCredentials,
        MaxAge:           cfg.CORS.MaxAge,
    }))
    
    // ============================================
    // LAYER 3: Protection
    // ============================================
    rateLimiter := middleware.NewRateLimiter(redisClient, true, []string{}, secLogger)
    r.Use(rateLimiter.GlobalLimit(100, time.Minute, 10))
    r.Use(middleware.RequestSize(cfg.HTTP.MaxBodySize))
    r.Use(middleware.Timeout(30 * time.Second))
    r.Use(chimiddleware.Compress(5))
    
    // ============================================
    // Routes
    // ============================================
    r.Get("/", func(w http.ResponseWriter, r *http.Request) {
        web.Success(w, r, map[string]string{"status": "ok"})
    })
    
    r.Get("/health", web.LivenessHandler)
    r.Get("/health/ready", web.ReadinessHandler())
    
    // API Routes
    r.Route("/api/v1", func(v1 chi.Router) {
        v1.Use(middleware.AcceptJSON())
        v1.Use(chimiddleware.AllowContentType("application/json"))
        
        v1.Get("/users", handleGetUsers)
        v1.Post("/users", handleCreateUser)
    })
    
    // Start
    srv := web.NewServer(cfg, logger, r)
    if err := srv.Start(); err != nil {
        logger.Error("server error", "error", err)
        os.Exit(1)
    }
}
```

## 📝 Configuration

### Config Struct

```go
type Config struct {
    HTTP struct {
        Port         int
        ReadTimeout  time.Duration
        WriteTimeout time.Duration
        IdleTimeout  time.Duration
        MaxBodySize  int64
    }
    
    TLS struct {
        Enabled  bool
        CertFile string
        KeyFile  string
    }
    
    CORS struct {
        Enabled          bool
        AllowedOrigins   []string
        AllowedMethods   []string
        AllowedHeaders   []string
        ExposedHeaders   []string
        AllowCredentials bool
        MaxAge           int
    }
    
    RateLimit struct {
        Enabled bool
        Global  struct {
            Limit  int
            Window time.Duration
            Burst  int
        }
    }
}
```

### Load Config

```go
cfg, err := web.LoadConfig()
if err != nil {
    log.Fatal(err)
}
```

## 🏥 Health Checks

### Liveness (Simple)

```go
r.Get("/health", web.LivenessHandler)
// Returns: 200 OK
```

### Readiness (with Checkers)

```go
// No checkers (always ready)
r.Get("/health/ready", web.ReadinessHandler())

// With custom checkers
checkers := []web.HealthChecker{
    databaseChecker,
    redisChecker,
}
r.Get("/health/ready", web.ReadinessHandler(checkers...))
```

**Custom Checker Example:**

```go
func databaseChecker() error {
    ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
    defer cancel()
    
    if err := db.PingContext(ctx); err != nil {
        return fmt.Errorf("database unhealthy: %w", err)
    }
    return nil
}
```

## 📤 Response Helpers

### Success Responses

```go
// 200 OK
web.Success(w, r, map[string]string{
    "message": "User created successfully",
    "id": "123",
})

// 201 Created
web.Created(w, r, user)

// 202 Accepted
web.Accepted(w, r, map[string]string{
    "job_id": "abc-123",
})

// 204 No Content
web.NoContent(w, r)
```

### Error Responses

```go
// Custom error
web.Error(w, r, &web.HTTPError{
    Code:    "VALIDATION_ERROR",
    Message: "Invalid email format",
    Status:  http.StatusBadRequest,
})

// Common errors
web.BadRequest(w, r, "Invalid input")
web.Unauthorized(w, r, "Invalid token")
web.Forbidden(w, r, "Access denied")
web.NotFound(w, r, "User not found")
web.InternalServerError(w, r, err)
```

## 🔧 Server Methods

### Start Server

```go
srv := web.NewServer(cfg, logger, router)

// Blocking start
if err := srv.Start(); err != nil {
    log.Fatal(err)
}
```

### Graceful Shutdown

```go
// Server handles SIGINT/SIGTERM automatically
// Waits 30 seconds for connections to close

// Manual shutdown
ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
defer cancel()

if err := srv.Shutdown(ctx); err != nil {
    log.Fatal(err)
}
```

## 🏗️ Microservice Structure

```
service/user-service/
├── cmd/
│   └── api/
│       └── main.go              # Server setup
├── internal/
│   ├── handler/
│   │   ├── user.go             # HTTP handlers
│   │   └── auth.go
│   ├── service/
│   │   └── user.go             # Business logic
│   ├── repository/
│   │   └── user.go             # Data access
│   └── middleware/
│       ├── auth.go             # JWT middleware (específico)
│       └── metrics.go          # Prometheus (específico)
├── .env                        # WEB_*, DATABASE_*, etc
├── .env.example
└── go.mod
```

**main.go:**

```go
package main

import (
    "github.com/marcelofabianov/web"
    "github.com/marcelofabianov/web/middleware"
    "github.com/marcelofabianov/database"
    
    "user-service/internal/handler"
    "user-service/internal/service"
    "user-service/internal/repository"
    authmw "user-service/internal/middleware"
)

func main() {
    // Load configs
    webCfg, _ := web.LoadConfig()
    dbCfg, _ := database.LoadConfig()
    
    // Logger
    logger := slog.Default()
    
    // Database
    db, _ := database.NewPostgres(dbCfg, logger)
    defer db.Close()
    
    // Dependencies
    userRepo := repository.NewUser(db)
    userSvc := service.NewUser(userRepo)
    userHandler := handler.NewUser(userSvc)
    
    // Router
    r := chi.NewRouter()
    
    // Generic middlewares (pkg/web)
    r.Use(middleware.RequestID())
    r.Use(middleware.Recovery(logger))
    r.Use(middleware.Logger(logger))
    
    // Health
    r.Get("/health", web.LivenessHandler)
    
    // API
    r.Route("/api/v1", func(v1 chi.Router) {
        // Service-specific middleware
        v1.Use(authmw.JWT(jwtSecret))
        
        v1.Get("/users", userHandler.List)
        v1.Post("/users", userHandler.Create)
    })
    
    // Start
    srv := web.NewServer(webCfg, logger, r)
    srv.Start()
}
```

## 📚 Environment Variables

Todas as variáveis com prefixo `WEB_*`:

```bash
# HTTP Server
WEB_HTTP_PORT=8080
WEB_HTTP_READ_TIMEOUT=10s
WEB_HTTP_WRITE_TIMEOUT=10s
WEB_HTTP_IDLE_TIMEOUT=120s
WEB_HTTP_MAX_BODY_SIZE=10485760

# TLS/HTTPS
WEB_TLS_ENABLED=false
WEB_TLS_CERT_FILE=./certs/server.crt
WEB_TLS_KEY_FILE=./certs/server.key

# CORS
WEB_CORS_ENABLED=true
WEB_CORS_ALLOWED_ORIGINS=https://app.example.com
WEB_CORS_ALLOWED_METHODS=GET,POST,PUT,DELETE,OPTIONS
WEB_CORS_ALLOWED_HEADERS=Accept,Authorization,Content-Type
WEB_CORS_EXPOSED_HEADERS=X-Request-ID
WEB_CORS_ALLOW_CREDENTIALS=true
WEB_CORS_MAX_AGE=300

# Rate Limiting
WEB_RATE_LIMIT_ENABLED=true
WEB_RATE_LIMIT_GLOBAL_LIMIT=100
WEB_RATE_LIMIT_GLOBAL_WINDOW=1m
WEB_RATE_LIMIT_GLOBAL_BURST=10
```

## 🔐 Security Best Practices

1. ✅ Always use HTTPS in production
2. ✅ Configure security headers
3. ✅ Enable rate limiting
4. ✅ Use CSRF protection for state-changing operations
5. ✅ Validate all inputs
6. ✅ Log security events
7. ✅ Use request IDs for tracing
8. ✅ Set appropriate timeouts
9. ✅ Limit request body size
10. ✅ Use graceful shutdown

## 📖 Mais Documentação

- **Middlewares:** [middleware/USAGE.md](./middleware/USAGE.md)
- **Examples:** Ver `service/` no monorepo
- **Config:** Ver `.env.example`

## 🎯 Features

✅ HTTP/HTTPS Server with TLS 1.2/1.3
✅ Graceful shutdown (30s timeout)
✅ Environment-based configuration
✅ Health checks (liveness/readiness)
✅ JSON response helpers
✅ 14 security-first middlewares
✅ Chi Router compatible
✅ Zero coupling (self-contained)
✅ Production ready

## 🚀 Quick Commands

```bash
# Build
go build ./...

# Test
go test ./...

# Run
WEB_HTTP_PORT=8080 go run cmd/api/main.go

# Generate TLS certs (dev)
openssl req -x509 -newkey rsa:4096 -nodes \
  -keyout server.key -out server.crt -days 365
```
