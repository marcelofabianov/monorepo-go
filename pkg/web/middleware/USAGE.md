# 🔐 pkg/web/middleware - Security-First Middlewares

Stack completo de middlewares para microservices seguros com Chi Router.

## 📦 Middlewares Disponíveis (14 essenciais)

### 🛡️ Security (9 middlewares)

1. **csrf.go** - CSRF Protection (OWASP Top 10)
2. **security_logger.go** - Security event logging  
3. **security_headers.go** - Security headers (CSP, HSTS, etc)
4. **rate_limit.go** - DDoS/Rate limiting
5. **https_only.go** - HTTPS enforcement
6. **cors.go** - CORS policy
7. **logger.go** - Request/response logging
8. **recovery.go** - Panic recovery
9. **request_size.go** - Body size limit protection

### ⚙️ Utilities (5 middlewares)

10. **accept.go** - Content-Type validation
11. **request_id.go** - Request ID tracking
12. **real_ip.go** - Real IP detection
13. **timeout.go** - Request timeout
14. **config.go** - Config structs

## 🚀 Uso com Chi Router

### Exemplo Completo - Microservice com Security-First

```go
package main

import (
    "log/slog"
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
    logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
    
    // Redis para rate limiting (opcional)
    redisClient := redis.NewClient(&redis.Options{
        Addr: "localhost:6379",
    })
    
    // Security Logger
    secLogger := middleware.NewSecurityLogger(logger)
    
    // Router
    r := chi.NewRouter()
    
    // ============================================
    // CAMADA 1: Middlewares Básicos
    // ============================================
    
    // Request ID (primeiro sempre)
    r.Use(middleware.RequestID())
    
    // Real IP detection
    r.Use(middleware.RealIP())
    
    // Recovery (panic recovery)
    r.Use(middleware.Recovery(logger))
    
    // Request logging
    r.Use(middleware.Logger(logger))
    
    // ============================================
    // CAMADA 2: Security Headers & HTTPS
    // ============================================
    
    // Security Headers
    r.Use(middleware.SecurityHeaders(middleware.SecurityHeadersConfig{
        XContentTypeOptions:      "nosniff",
        XFrameOptions:            "DENY",
        ContentSecurityPolicy:    "default-src 'self'",
        ReferrerPolicy:           "no-referrer",
        StrictTransportSecurity:  "max-age=31536000; includeSubDomains",
        CacheControl:             "no-store",
        PermissionsPolicy:        "camera=(), microphone=(), geolocation=()",
        XDNSPrefetchControl:      "off",
        XDownloadOptions:         "noopen",
    }))
    
    // HTTPS Only
    r.Use(middleware.HTTPSOnly(middleware.HTTPSOnlyConfig{
        Enabled: true,
    }))
    
    // ============================================
    // CAMADA 3: CORS & Rate Limiting
    // ============================================
    
    // CORS
    r.Use(middleware.CORS(middleware.CORSConfig{
        AllowedOrigins:   []string{"https://app.example.com"},
        AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
        AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
        ExposedHeaders:   []string{"X-Request-ID"},
        AllowCredentials: true,
        MaxAge:           300,
    }))
    
    // Rate Limiting (Global)
    rateLimiter := middleware.NewRateLimiter(
        redisClient,
        true,
        []string{"10.0.0.0/8"}, // Trusted proxies
        secLogger,
    )
    r.Use(rateLimiter.GlobalLimit(100, time.Minute, 10))
    
    // ============================================
    // CAMADA 4: Request Protection
    // ============================================
    
    // Request size limit (10MB)
    r.Use(middleware.RequestSize(10 * 1024 * 1024))
    
    // Request timeout (30s)
    r.Use(middleware.Timeout(30 * time.Second))
    
    // Chi native: Compression (opcional)
    r.Use(chimiddleware.Compress(5))
    
    // ============================================
    // CAMADA 5: API Routes
    // ============================================
    
    r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
        w.WriteHeader(http.StatusOK)
        w.Write([]byte(`{"status":"healthy"}`))
    })
    
    // API v1 com CSRF Protection
    r.Route("/api/v1", func(v1 chi.Router) {
        // Accept JSON only
        v1.Use(middleware.AcceptJSON())
        v1.Use(chimiddleware.AllowContentType("application/json"))
        
        // CSRF Protection
        csrf := middleware.NewCSRFProtection(
            "your-secret-key-32-bytes-minimum",
            "csrf_token",
            "X-CSRF-Token",
            24*time.Hour,
            []string{"/api/v1/health"},
            true,
            secLogger,
        )
        v1.Use(csrf.Protect())
        v1.Get("/csrf-token", csrf.GetTokenHandler())
        
        // Rate limit per user
        v1.Use(rateLimiter.PerUserLimit(50, time.Minute, 5))
        
        // Your routes
        v1.Get("/users", handleGetUsers)
        v1.Post("/users", handleCreateUser)
    })
    
    // Start server
    webCfg, _ := web.LoadConfig()
    srv := web.NewServer(webCfg, logger, r)
    srv.Start()
}
```

## 🎯 Configuração por Ambiente

### Development
```go
// Desabilitar HTTPS only
r.Use(middleware.HTTPSOnly(middleware.HTTPSOnlyConfig{
    Enabled: false, // Dev sem TLS
}))

// CORS permissivo
r.Use(middleware.CORS(middleware.CORSConfig{
    AllowedOrigins: []string{"*"},
    // ...
}))

// Rate limit mais alto
rateLimiter.GlobalLimit(1000, time.Minute, 100)
```

### Production
```go
// HTTPS obrigatório
r.Use(middleware.HTTPSOnly(middleware.HTTPSOnlyConfig{
    Enabled: true,
}))

// CORS restrito
r.Use(middleware.CORS(middleware.CORSConfig{
    AllowedOrigins: []string{"https://app.example.com"},
    // ...
}))

// Rate limit estrito
rateLimiter.GlobalLimit(100, time.Minute, 10)
```

## 📊 Estratégias de Rate Limiting

```go
// Global (por IP)
rateLimiter.GlobalLimit(100, time.Minute, 10)

// Por usuário autenticado
rateLimiter.PerUserLimit(50, time.Minute, 5)

// Por rota específica
rateLimiter.PerRouteLimit("/api/v1/login", 5, time.Minute, 1)

// Estratégia customizada
rateLimiter.Limit(middleware.RateLimitRule{
    Limit:    10,
    Window:   time.Minute,
    Burst:    2,
    Strategy: middleware.ByUser(rateLimiter),
})
```

## 🔐 CSRF Protection

```go
// Criar proteção CSRF
csrf := middleware.NewCSRFProtection(
    "secret-key-min-32-bytes",  // Secret key
    "csrf_token",               // Cookie name
    "X-CSRF-Token",             // Header name
    24*time.Hour,               // TTL
    []string{"/public/api"},    // Exempt paths
    true,                       // Enabled
    secLogger,                  // Security logger
)

// Aplicar proteção
r.Use(csrf.Protect())

// Endpoint para obter token
r.Get("/csrf-token", csrf.GetTokenHandler())
```

## 📝 Security Logging

```go
// Criar security logger
secLogger := middleware.NewSecurityLogger(logger)

// Usado automaticamente por:
// - CSRF violations
// - Rate limit exceeded
// - IP spoofing detection

// Uso manual no seu código:
secLogger.LogAuthEvent(
    middleware.EventLoginSuccess,
    "user@example.com",
    r,
    true,
    "",
)
```

## ⚡ Performance Tips

1. **Request ID** - Sempre primeiro
2. **Real IP** - Logo após Request ID
3. **Recovery** - Antes de logger
4. **Rate Limiting** - Antes de lógica de negócio
5. **Compression** - Por último (Chi nativo)

## 🏗️ Arquitetura Recomendada

```
service/user-service/
├── cmd/
│   └── api/
│       └── main.go          # Setup de middlewares aqui
├── internal/
│   ├── handler/            # HTTP handlers
│   ├── middleware/         # Middlewares ESPECÍFICOS do service
│   │   ├── auth.go        # JWT, OAuth
│   │   └── metrics.go     # Prometheus
│   └── service/           # Business logic
└── go.mod
```

## 🔒 Checklist Security-First

- [ ] Request ID tracking
- [ ] Real IP detection
- [ ] Panic recovery
- [ ] Request logging
- [ ] Security headers
- [ ] HTTPS enforcement
- [ ] CORS policy
- [ ] Rate limiting (global)
- [ ] CSRF protection
- [ ] Request size limit
- [ ] Request timeout
- [ ] Accept JSON validation
- [ ] Content-Type validation

## 📚 Dependências

```bash
go get github.com/go-chi/chi/v5
go get github.com/redis/go-redis/v9        # Para rate limiting
go get github.com/go-redis/redis_rate/v10  # Para rate limiting
go get github.com/go-chi/cors              # Para CORS
go get github.com/google/uuid              # Para request ID
```

## 🎯 Resultado

Com essa stack, cada microservice terá:

✅ **OWASP Top 10 Protection**
✅ **DDoS/Brute Force Protection**  
✅ **Security Observability**
✅ **Distributed Tracing**
✅ **100% Self-Contained**
