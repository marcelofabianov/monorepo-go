# 🚀 Go Microservices Monorepo

[![Go Version](https://img.shields.io/badge/Go-1.25.1-blue.svg)](https://golang.org/)
[![Architecture](https://img.shields.io/badge/Architecture-Self--Contained%20Packages-green.svg)]()
[![Security](https://img.shields.io/badge/Security-First-red.svg)]()

> **Monorepo moderno em Go com arquitetura self-contained packages e security-first approach**

Sistema completo de microservices educacionais com packages reutilizáveis, middlewares de segurança e gerenciamento via Makefile.

---

## 📋 Índice

- [Visão Geral](#-visão-geral)
- [Arquitetura](#-arquitetura)
- [Quick Start](#-quick-start)
- [Estrutura do Projeto](#-estrutura-do-projeto)
- [Packages Disponíveis](#-packages-disponíveis)
- [Microservices](#-microservices)
- [Makefile](#-makefile)
- [Configuração](#-configuração)
- [Segurança](#-segurança)
- [Documentação](#-documentação)
- [Desenvolvimento](#-desenvolvimento)

---

## 🎯 Visão Geral

Monorepo Go que implementa 4 microservices educacionais (course, classroom, lesson, enrollment) com 6 packages reutilizáveis e independentes.

### ✨ Características Principais

- **🔐 Security-First:** 14 middlewares focados em segurança (CSRF, Rate Limiting, Security Headers, etc.)
- **📦 Self-Contained Packages:** Zero coupling - cada package pode ser extraído para repositório separado
- **🔧 Viper Configuration:** Cada package tem configuração independente com prefixo único
- **🎭 Chi Router:** HTTP router moderno e minimalista para todos os services
- **🛠️ Makefile Completo:** 30+ comandos para build, run, test, logs, health checks
- **🧪 Testável:** Dependency injection, interfaces, mocks fáceis
- **📚 Documentado:** USAGE.md em cada package + Makefile documentation

---

## 🏗️ Arquitetura

### Self-Contained Packages Pattern

```
┌─────────────────────────────────────────────────────────────┐
│                    Go Workspace (go.work)                    │
└─────────────────────────────────────────────────────────────┘
                              │
          ┌───────────────────┼───────────────────┐
          │                   │                   │
          ▼                   ▼                   ▼
    ┌──────────┐        ┌──────────┐       ┌──────────┐
    │ pkg/web  │        │pkg/cache │       │pkg/logger│
    │          │        │          │       │          │
    │ .env     │        │ .env     │       │ .env     │
    │ Viper    │        │ Viper    │       │ Viper    │
    │ WEB_*    │        │ CACHE_*  │       │ LOGGER_* │
    └──────────┘        └──────────┘       └──────────┘
          │                   │                   │
          └───────────────────┼───────────────────┘
                              │
                    ┌─────────┴─────────┐
                    │  Services Layer   │
                    └─────────┬─────────┘
                              │
          ┌───────────────────┼───────────────────┐
          │                   │                   │
          ▼                   ▼                   ▼
    ┌──────────┐        ┌──────────┐       ┌──────────┐
    │ course   │        │classroom │       │  lesson  │
    │  :8080   │        │  :8081   │       │  :8082   │
    └──────────┘        └──────────┘       └──────────┘
```

### Principles

- ✅ **Zero Coupling:** Packages não importam uns aos outros
- ✅ **Environment Isolation:** Cada package tem prefixo único (LOGGER_*, CACHE_*, WEB_*)
- ✅ **Dependency Injection:** Interfaces e DI em vez de imports diretos
- ✅ **Extract Ready:** Qualquer package pode virar repo standalone instantly
- ✅ **Viper Auto-Discovery:** `.env` files encontrados automaticamente (até 5 níveis)

---

## ⚡ Quick Start

### 1. Clone & Setup

```bash
# Clone do repositório
cd /path/to/workspace

# Criar .env files (cada package tem .env.example)
cp pkg/logger/.env.example pkg/logger/.env
cp pkg/cache/.env.example pkg/cache/.env
cp pkg/web/.env.example pkg/web/.env
# ... (repeat para outros packages/services)
```

### 2. Start All Services

```bash
# Build + Start todos os services em background
make up

# Verificar health
make status
```

### 3. Test Endpoints

```bash
# Course service (port 8080)
curl http://localhost:8080/
curl http://localhost:8080/health

# Classroom service (port 8081)
curl http://localhost:8081/health/ready

# Ver logs
make logs
```

### 4. Stop Everything

```bash
make down
```

---

## 📁 Estrutura do Projeto

```
work/
├── go.work                    # Go workspace configuration
├── Makefile                   # 30+ management commands
├── MAKEFILE.md               # Makefile documentation
├── README.md                 # Este arquivo
│
├── pkg/                      # 🔧 Self-Contained Packages
│   ├── logger/              # Structured logging (slog wrapper)
│   │   ├── config.go       # Viper config (LOGGER_* prefix)
│   │   ├── .env.example
│   │   ├── README.md
│   │   └── go.mod
│   │
│   ├── retry/               # Retry with backoff strategies
│   │   ├── config.go       # Viper config (RETRY_* prefix)
│   │   ├── backoff_strategy.go
│   │   ├── .env.example
│   │   └── go.mod
│   │
│   ├── cache/               # Redis cache with pool
│   │   ├── config.go       # Viper config (CACHE_* prefix)
│   │   ├── cache.go
│   │   ├── .env.example
│   │   └── go.mod
│   │
│   ├── database/            # Database connections
│   │   ├── config.go       # Viper config (DATABASE_* prefix)
│   │   ├── .env.example
│   │   └── go.mod
│   │
│   ├── validation/          # Input validation
│   │   ├── config.go       # Viper config (VALIDATION_* prefix)
│   │   ├── .env.example
│   │   └── go.mod
│   │
│   └── web/                 # 🌐 HTTP utilities + Chi integration
│       ├── config.go       # Viper config (WEB_* prefix)
│       ├── response.go     # JSON response helpers
│       ├── health.go       # Health check handlers
│       ├── USAGE.md        # Complete usage guide
│       ├── .env.example
│       ├── go.mod
│       │
│       ├── middleware/     # 🔐 14 Security-First Middlewares
│       │   ├── accept.go           # Content-Type validation
│       │   ├── cors.go             # CORS configuration
│       │   ├── csrf.go             # CSRF protection (HMAC)
│       │   ├── https_only.go      # Force HTTPS
│       │   ├── logger.go           # Request/response logging
│       │   ├── rate_limit.go      # Distributed rate limiting (Redis)
│       │   ├── real_ip.go         # Real IP detection
│       │   ├── recovery.go        # Panic recovery
│       │   ├── request_id.go      # Request ID tracking
│       │   ├── request_size.go    # Body size limits
│       │   ├── security_headers.go # Security headers (CSP, HSTS, etc)
│       │   ├── security_logger.go # Security event logging
│       │   ├── timeout.go         # Request timeout
│       │   ├── config.go          # Middleware configurations
│       │   └── USAGE.md           # Middleware integration guide
│       │
│       └── chi/            # Chi router integrations
│           └── ...
│
└── service/                # 🎓 Microservices Layer
    ├── course/            # Course management service
    │   ├── main.go       # HTTP server (port 8080)
    │   ├── .env.example
    │   └── go.mod
    │
    ├── classroom/         # Classroom management service
    │   ├── main.go       # HTTP server (port 8081)
    │   ├── .env.example
    │   └── go.mod
    │
    ├── lesson/            # Lesson management service
    │   ├── main.go       # HTTP server (port 8082)
    │   ├── .env.example
    │   └── go.mod
    │
    └── enrollment/        # Enrollment management service
        ├── main.go       # HTTP server (port 8083)
        ├── .env.example
        └── go.mod
```

---

## 📦 Packages Disponíveis

### `pkg/logger` - Structured Logging
- Wrapper sobre `slog` com configuração via Viper
- Prefixo: `LOGGER_*`
- Features: Múltiplos níveis, JSON/Text output, adicionar campos extras

### `pkg/retry` - Retry with Backoff
- Retry com estratégias configuráveis
- Prefixo: `RETRY_*`
- Strategies: Exponential, Linear, Constant Backoff
- Features: Jitter, max retries, configurável

### `pkg/cache` - Redis Cache
- Cliente Redis com pool de conexões
- Prefixo: `CACHE_*`
- Features: TTL, Pool management, retry automático

### `pkg/database` - Database Connections
- Gerenciamento de conexões SQL
- Prefixo: `DATABASE_*`
- Features: Connection pooling, health checks

### `pkg/validation` - Input Validation
- Validação de inputs HTTP
- Prefixo: `VALIDATION_*`
- Features: Schema validation, custom rules

### `pkg/web` - HTTP Utilities
- Response helpers, health checks, Chi integration
- Prefixo: `WEB_*`
- **14 Security-First Middlewares:**
  - ✅ CSRF Protection (HMAC-based)
  - ✅ Rate Limiting (Redis + Circuit Breaker)
  - ✅ Security Headers (CSP, HSTS, X-Frame, etc)
  - ✅ Request Size Limits
  - ✅ CORS Configuration
  - ✅ Real IP Detection
  - ✅ Request ID Tracking
  - ✅ Timeout Control
  - ✅ Logger (structured slog)
  - ✅ Recovery (panic handler)
  - ✅ Accept Header Validation
  - ✅ HTTPS Enforcement
  - ✅ Security Event Logging

📖 **Documentação:** [pkg/web/USAGE.md](pkg/web/USAGE.md) | [pkg/web/middleware/USAGE.md](pkg/web/middleware/USAGE.md)

---

## 🎓 Microservices

| Service | Port | Description | Endpoints |
|---------|------|-------------|-----------|
| **course** | 8080 | Course management | `/`, `/health`, `/health/ready` |
| **classroom** | 8081 | Classroom management | `/`, `/health`, `/health/ready` |
| **lesson** | 8082 | Lesson management | `/`, `/health`, `/health/ready` |
| **enrollment** | 8083 | Enrollment management | `/`, `/health`, `/health/ready` |

### Common Endpoints

Todos os services implementam:

- `GET /` - Service info (JSON)
- `GET /health` - Health check (simple)
- `GET /health/ready` - Readiness check (detailed)

### Example Service Structure

```go
package main

import (
    "log/slog"
    "github.com/go-chi/chi/v5"
    "github.com/go-chi/chi/v5/middleware"
    "github.com/marcelofabianov/web"
)

func main() {
    r := chi.NewRouter()
    
    // Basic middlewares
    r.Use(middleware.RequestID)
    r.Use(middleware.RealIP)
    r.Use(middleware.Recoverer)
    
    // Routes
    r.Get("/", handleRoot)
    r.Get("/health", web.HealthCheckHandler())
    r.Get("/health/ready", web.ReadinessCheckHandler(checks))
    
    // Start server...
}
```

---

## 🛠️ Makefile

**30+ comandos disponíveis** - [Ver documentação completa](MAKEFILE.md)

### Quick Commands

```bash
# Setup + Start tudo
make up

# Stop + Clean
make down

# Restart everything
make restart

# Check health
make status
```

### Build & Test

```bash
make build              # Build all services
make build-course       # Build specific service
make test               # Run all tests
make test-pkg-web       # Test specific package
make lint               # Run linters
```

### Run Services

```bash
# Background (todos)
make run-all

# Foreground (individual)
make run-course
make run-classroom
make run-lesson
make run-enrollment
```

### Monitoring

```bash
# Health checks
make health
make health-course

# View logs
make logs
make logs-course

# Process status
make ps
```

### Cleanup

```bash
make stop        # Stop all services
make clean       # Remove binaries
make clean-logs  # Remove logs
make down        # Stop + Clean everything
```

📖 **Documentação:** [MAKEFILE.md](MAKEFILE.md)

---

## ⚙️ Configuração

### Environment Variables Pattern

Cada package usa **prefixo único** para evitar conflitos:

```bash
# pkg/logger - Prefix: LOGGER_*
LOGGER_LEVEL=info
LOGGER_FORMAT=json

# pkg/cache - Prefix: CACHE_*
CACHE_REDIS_HOST=localhost
CACHE_REDIS_PORT=6379

# pkg/web - Prefix: WEB_*
WEB_HTTP_PORT=8080
WEB_HTTP_TIMEOUT=30s

# service/course - Prefix: COURSE_*
COURSE_SERVER_PORT=8080
COURSE_SERVER_HOST=0.0.0.0
```

### Configuration Loading

Cada package usa Viper com auto-discovery:

```go
// pkg/logger/config.go
func LoadConfig() (*Config, error) {
    v := viper.New()
    v.SetEnvPrefix("LOGGER")                    // Prefix único
    v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
    v.AutomaticEnv()
    
    // Busca .env automaticamente (até 5 níveis acima)
    if envFile := findEnvFile(); envFile != "" {
        v.SetConfigFile(envFile)
        _ = v.ReadInConfig()
    }
    
    setDefaults(v)
    
    return &Config{
        Level:  v.GetString("level"),
        Format: v.GetString("format"),
    }, nil
}
```

### Creating .env Files

```bash
# Para cada package/service
cp pkg/logger/.env.example pkg/logger/.env
cp pkg/cache/.env.example pkg/cache/.env
cp pkg/web/.env.example pkg/web/.env
cp service/course/.env.example service/course/.env
# ... etc
```

---

## 🔐 Segurança

### Security-First Approach

Todos os services implementam camadas de segurança:

```go
// Example: service/course/main.go
r.Use(middleware.RequestID)
r.Use(middleware.RealIP)
r.Use(middleware.Recoverer)
r.Use(webmw.SecurityHeaders(securityConfig))
r.Use(webmw.RateLimit(rateLimiter))
r.Use(webmw.CSRFProtection(csrfConfig))
```

### Middlewares de Segurança

| Middleware | Proteção | Configurável |
|------------|----------|--------------|
| **Security Headers** | CSP, HSTS, X-Frame, X-Content-Type | ✅ |
| **CSRF** | Cross-Site Request Forgery | ✅ |
| **Rate Limiting** | DDoS, brute force | ✅ |
| **Request Size** | Large payloads | ✅ |
| **HTTPS Only** | Man-in-the-middle | ✅ |
| **Real IP** | IP spoofing | ✅ |
| **Security Logger** | Audit trail | ✅ |

### Rate Limiting Strategies

```go
// IP-based
limiter := middleware.RateLimitByIP(redisClient, 100, time.Minute)

// User-based
limiter := middleware.RateLimitByUser(redisClient, 1000, time.Hour)

// Composite
limiter := middleware.RateLimitComposite(
    middleware.RateLimitByIP(...),
    middleware.RateLimitByUser(...),
)
```

---

## 📚 Documentação

Cada package/módulo tem documentação completa:

- [pkg/web/USAGE.md](pkg/web/USAGE.md) - HTTP utilities + Chi integration
- [pkg/web/middleware/USAGE.md](pkg/web/middleware/USAGE.md) - Middleware guide
- [MAKEFILE.md](MAKEFILE.md) - Makefile commands
- `pkg/*/README.md` - Documentação de cada package

---

## 🔧 Desenvolvimento

### Prerequisites

- Go 1.25.1 ou superior
- Redis (para cache e rate limiting)
- Make (para usar Makefile)

### Setup Development

```bash
# 1. Clone repo
cd /path/to/workspace

# 2. Setup .env files
for pkg in logger retry cache database validation web; do
    cp pkg/$pkg/.env.example pkg/$pkg/.env
done

for svc in course classroom lesson enrollment; do
    cp service/$svc/.env.example service/$svc/.env
done

# 3. Install dependencies
go mod download

# 4. Run tests
make test

# 5. Start services
make up
```

### Adding New Service

```bash
# 1. Create service directory
mkdir -p service/newservice

# 2. Create main.go (copy from existing service)
cp service/course/main.go service/newservice/

# 3. Create go.mod
cd service/newservice
go mod init github.com/marcelofabianov/newservice

# 4. Add replace directives
echo 'replace github.com/marcelofabianov/web => ../../pkg/web' >> go.mod

# 5. Add to go.work
cd ../..
echo './service/newservice' >> go.work

# 6. Add to Makefile (build, run, health targets)
```

### Adding New Package

```bash
# 1. Create package directory
mkdir -p pkg/newpkg

# 2. Create config.go with Viper
cat > pkg/newpkg/config.go << 'EOF'
package newpkg

import "github.com/spf13/viper"

func LoadConfig() (*Config, error) {
    v := viper.New()
    v.SetEnvPrefix("NEWPKG")  // Unique prefix
    v.AutomaticEnv()
    // ... implement
}
EOF

# 3. Create .env.example
cat > pkg/newpkg/.env.example << 'EOF'
NEWPKG_SETTING=value
EOF

# 4. Create go.mod
cd pkg/newpkg
go mod init github.com/marcelofabianov/newpkg

# 5. Add to go.work
cd ../..
echo './pkg/newpkg' >> go.work
```

### Running Tests

```bash
# All tests
make test

# Specific package
make test-pkg-web
make test-pkg-cache

# With coverage
go test -v -cover ./...

# Watch mode
make dev-watch
```

### Code Quality

```bash
# Format code
make fmt

# Lint
make lint

# Vet
make vet

# All checks
make check
```

---

## 🎯 Roadmap

### ✅ Completed
- [x] Self-contained packages architecture
- [x] Security-first middlewares
- [x] Basic HTTP endpoints for all services
- [x] Makefile management system
- [x] Complete documentation
- [x] Viper configuration per package

### 🚧 Future Enhancements
- [ ] Add business logic to services (API routes)
- [ ] Implement authentication/authorization middleware (JWT)
- [ ] Add database integration to services
- [ ] Create Docker Compose setup
- [ ] Add Kubernetes manifests
- [ ] Implement service-to-service communication (gRPC?)
- [ ] Add OpenAPI/Swagger documentation
- [ ] Implement distributed tracing (OpenTelemetry)
- [ ] Add metrics collection (Prometheus)
- [ ] Create CI/CD pipelines (GitHub Actions)

---

## 📄 Licença

MIT

---

## 🤝 Contribuindo

1. Fork o projeto
2. Crie uma branch para sua feature (`git checkout -b feature/AmazingFeature`)
3. Commit suas mudanças (`git commit -m 'Add some AmazingFeature'`)
4. Push para a branch (`git push origin feature/AmazingFeature`)
5. Abra um Pull Request

---

## 📞 Contato

**Marcelo Fabiano** - [@marcelofabianov](https://github.com/marcelofabianov)

Project Link: [https://github.com/marcelofabianov/go-microservices-monorepo](https://github.com/marcelofabianov/go-microservices-monorepo)

---

<p align="center">Made with ❤️ using Go</p>
