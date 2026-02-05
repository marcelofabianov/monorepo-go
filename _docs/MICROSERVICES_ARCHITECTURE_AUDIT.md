# Análise Arquitetural: Microservices em Monorepo

## 🎯 Objetivo
Garantir que o monorepo permita implantação independente de microservices sem violações de dependência que causem acoplamento.

---

## 📊 Estrutura Atual

```
work/
├── config/                    # ⚠️ RISCO: Config centralizado
│   ├── config.go
│   ├── port.go
│   ├── helper.go
│   └── adapter/
│       ├── cache.go
│       └── logger.go
│
├── pkg/                       # ✅ OK: Bibliotecas compartilhadas
│   ├── cache/
│   ├── database/
│   ├── logger/
│   ├── retry/
│   ├── validation/
│   └── web/
│
└── service/                   # 🎯 Microservices
    ├── classroom/
    ├── course/
    ├── enrollment/
    └── lesson/
```

---

## 🚨 VIOLAÇÕES IDENTIFICADAS

### 1. ❌ CRÍTICO: Config Centralizado

**Problema:**
```
config/
  ├── config.go          # Carrega TODAS as configurações
  ├── adapter/
  │   ├── cache.go       # Conhece pkg/cache
  │   └── logger.go      # Conhece pkg/logger
```

**Por que é um problema?**
- Se `service/course` usa `config`, ele traz TODAS as configs do monorepo
- Mudança em config afeta TODOS os services
- Impossível implantar services independentemente
- Deploy de um service pode quebrar outro

**Fluxo de Dependência Problemático:**
```
service/course → config → pkg/cache
                       → pkg/logger
                       → pkg/database
                       → pkg/validation
                       → ...

service/classroom → config → (mesmas dependências)
```

**Acoplamento:**
- Todos os services dependem do MESMO config
- Config conhece TODOS os pkgs
- Mudança em 1 pkg força rebuild de TODOS os services

---

### 2. ⚠️ MÉDIO: config/adapter Viola Isolamento

**Problema:**
```go
// config/adapter/cache.go
package adapter

import (
    "github.com/marcelofabianov/cache"    // ⚠️
    "github.com/marcelofabianov/config"   // ⚠️
)

func NewCacheConfig(c *config.Config) *cache.Config { ... }
```

**Por que é um problema?**
- adapter conhece implementações concretas (cache, logger)
- Viola princípio de inversão de dependência
- Adicionar novo pkg requer modificar config/adapter

---

### 3. ⚠️ MÉDIO: Config Monolítico

**Problema:**
```go
// config/config.go
type Config struct {
    General    GeneralConfig
    Logger     LoggerConfig
    HTTP       HTTPConfig
    Server     ServerConfig
    Database   DatabaseConfig      // ⚠️ course precisa
    Redis      RedisConfig          // ⚠️ todos precisam?
    Migrations MigrationsConfig     // ⚠️ só alguns precisam
    JWT        JWTConfig            // ⚠️ só auth precisa
}
```

**Por que é um problema?**
- service/lesson precisa de JWT mas carrega Database, Redis, Migrations...
- Configurações vazadas entre services
- .env global com configs de TODOS os services

---

## ✅ SOLUÇÕES ARQUITETURAIS

### Solução 1: Config Distribuído (Recomendada)

**Estrutura:**
```
service/course/
  ├── cmd/
  ├── internal/
  └── config/              # ✅ Config específico do service
      ├── config.go        # Apenas configs que course precisa
      └── .env             # Apenas vars que course precisa

service/classroom/
  ├── cmd/
  ├── internal/
  └── config/              # ✅ Config específico do service
      ├── config.go
      └── .env

pkg/                       # ✅ Libs compartilhadas (sem config)
  ├── cache/
  ├── logger/
  └── ...
```

**Vantagens:**
- ✅ Services completamente independentes
- ✅ Deploy independente
- ✅ Mudança em config de course NÃO afeta classroom
- ✅ Cada service carrega apenas o que precisa

**Exemplo:**
```go
// service/course/config/config.go
package config

type Config struct {
    Server   ServerConfig
    Database DatabaseConfig
    Logger   LoggerConfig
    Cache    CacheConfig
    // Apenas o que course precisa!
}
```

---

### Solução 2: Config como Interface (Alternativa)

**Manter config centralizado mas usar interfaces:**

```go
// pkg/cache/config.go (atual - OK!)
type ConfigProvider interface {
    GetHost() string
    GetPort() int
    // ...
}

// service/course/config/config.go
type CourseConfig struct {
    Server   ServerConfig
    Database DatabaseConfig
}

// Implementa apenas as interfaces que precisa
func (c *CourseConfig) GetHost() string { return c.Database.Host }
```

**Vantagens:**
- ✅ Cada service tem seu próprio config
- ✅ Implementa apenas interfaces necessárias
- ✅ Compartilha contratos, não implementações

**Desvantagens:**
- ⚠️ Ainda há coupling via interfaces
- ⚠️ Mudança em interface afeta múltiplos services

---

### Solução 3: Configuração via Environment Variables (Simples)

**Cada service lê suas próprias variáveis:**

```go
// service/course/config/config.go
package config

import "os"

func Load() (*Config, error) {
    return &Config{
        Server: ServerConfig{
            Port: getEnvInt("COURSE_SERVER_PORT", 8001),
            Host: getEnv("COURSE_SERVER_HOST", "0.0.0.0"),
        },
        Database: DatabaseConfig{
            Host: getEnv("COURSE_DB_HOST", "localhost"),
            Port: getEnvInt("COURSE_DB_PORT", 5432),
            // ...
        },
    }, nil
}
```

**Vantagens:**
- ✅ Máxima simplicidade
- ✅ Zero dependências entre services
- ✅ Padrão cloud-native (12-factor app)

---

## 🎯 ARQUITETURA RECOMENDADA PARA MICROSERVICES

### Princípios

1. **Independência de Deploy**
   - Cada service pode ser implantado sem afetar outros
   - Cada service tem seu próprio config

2. **Compartilhamento Mínimo**
   - pkg/* são bibliotecas utilitárias (cache, logger, retry)
   - NÃO compartilhar lógica de negócio
   - NÃO compartilhar models entre services

3. **Comunicação via API**
   - Services se comunicam via HTTP/gRPC
   - NÃO imports diretos entre services

4. **Dados Isolados**
   - Cada service tem seu próprio banco de dados
   - NÃO compartilhar schemas

### Estrutura Proposta

```
work/
├── pkg/                           # Libs técnicas (OK!)
│   ├── cache/                     # Redis client
│   ├── logger/                    # Structured logging
│   ├── retry/                     # Retry logic
│   ├── httpclient/                # HTTP client wrapper
│   └── errors/                    # Error handling
│
├── service/course/
│   ├── cmd/
│   │   └── api/
│   │       └── main.go
│   ├── internal/
│   │   ├── domain/                # Entidades
│   │   ├── usecase/               # Regras de negócio
│   │   ├── repository/            # Acesso a dados
│   │   └── api/                   # HTTP handlers
│   ├── config/                    # ✅ Config próprio
│   │   ├── config.go
│   │   └── .env.example
│   ├── go.mod                     # Dependencies próprias
│   └── Dockerfile                 # Deploy independente
│
├── service/classroom/
│   ├── cmd/
│   ├── internal/
│   ├── config/                    # ✅ Config próprio
│   ├── go.mod
│   └── Dockerfile
│
└── shared/                        # ⚠️ Use com cuidado
    └── proto/                     # Apenas contratos gRPC
```

---

## 🔍 CHECKLIST DE VALIDAÇÃO

### Para cada service, verificar:

- [ ] Tem seu próprio `config/` ?
- [ ] Tem seu próprio `.env` ?
- [ ] Pode ser compilado independentemente?
- [ ] Pode ser implantado sem outros services?
- [ ] NÃO importa outros services?
- [ ] NÃO compartilha models de domínio?
- [ ] Usa pkg/* apenas como libs técnicas?

### Para pkg/*, verificar:

- [ ] É uma biblioteca técnica (não negócio)?
- [ ] NÃO importa services?
- [ ] NÃO importa config global?
- [ ] Pode ser versionado independentemente?
- [ ] Pode ser extraído para lib externa?

---

## 🚀 PLANO DE MIGRAÇÃO

### Fase 1: Remover Config Centralizado

1. Criar `service/course/config/`
2. Mover configs relevantes para lá
3. Remover dependência de `config` global

### Fase 2: Criar Configs por Service

1. `service/classroom/config/`
2. `service/enrollment/config/`
3. `service/lesson/config/`

### Fase 3: Deprecar Config Global

1. Mover `config/adapter/` para cada service
2. Deprecar `config/` global
3. Cada service gerencia seus adaptadores

### Fase 4: Validar Independência

```bash
# Cada service deve compilar sozinho
cd service/course && go build ./...
cd service/classroom && go build ./...

# Cada service deve ter deps mínimas
go mod graph | grep "marcelofabianov"
# Deve mostrar apenas pkg/*, não outros services
```

---

## 📋 REGRAS DE ARQUITETURA

### ✅ PERMITIDO

```
service/course → pkg/cache      ✅
service/course → pkg/logger     ✅
service/course → pkg/httpclient ✅
```

### ❌ PROIBIDO

```
service/course → service/classroom     ❌ NUNCA!
service/course → config (global)       ❌ Cria acoplamento
pkg/cache      → config                ❌ Já corrigido!
pkg/cache      → service/*             ❌ Inversão de deps
```

### ⚠️ CUIDADO

```
service/course → shared/models         ⚠️ Acoplamento de dados
service/course → shared/proto          ✅ OK se apenas contratos
```

---

## 🎓 PADRÕES RECOMENDADOS

### 1. Repository Pattern
Cada service tem seus próprios repositories (NÃO compartilhar).

### 2. Use Cases
Lógica de negócio isolada em cada service.

### 3. DTOs para Comunicação
```go
// service/course/internal/api/dto/course.go
type CourseDTO struct {
    ID   string
    Name string
}

// NÃO compartilhar entre services!
```

### 4. Anti-Corruption Layer
```go
// service/enrollment/internal/client/course.go
type CourseClient struct {
    httpClient *httpclient.Client
}

func (c *CourseClient) GetCourse(id string) (*Course, error) {
    // Chama API do service/course
    // Converte DTO externo para model interno
}
```

---

## 🔥 ALERTAS CRÍTICOS

### ⚠️ Se você vê isso, há problema:

1. **Import entre services:**
   ```go
   import "github.com/marcelofabianov/service/course/domain"  // ❌
   ```

2. **Config global usado por service:**
   ```go
   cfg, _ := config.Load()  // ❌ Se config é global
   ```

3. **Shared models de domínio:**
   ```go
   // shared/domain/course.go  // ❌ Vazamento de domínio
   ```

4. **Dependência transitiva entre services:**
   ```go
   // go.mod do service/enrollment
   require (
       github.com/marcelofabianov/service/course v1.0.0  // ❌
   )
   ```

---

## 📊 MÉTRICAS DE SAÚDE ARQUITETURAL

### Boa Arquitetura de Microservices:

- **Acoplamento**: Baixo (< 5 dependências por service)
- **Coesão**: Alta (cada service faz 1 coisa bem)
- **Independência**: 100% (pode deployar sozinho)
- **Compartilhamento**: Mínimo (apenas libs técnicas)

### Sinais de Alerta:

- 🔴 Service depende de > 10 outros módulos
- 🔴 Mudança em pkg/* quebra múltiplos services
- 🔴 Service importa outro service
- 🔴 Config global com > 20 propriedades

---

## 💡 PRÓXIMOS PASSOS

1. **Auditar Dependências Atuais**
   ```bash
   ./scripts/audit-deps.sh
   ```

2. **Criar Configs por Service**
   - Começar com service/course
   - Replicar padrão para outros

3. **Documentar Contratos de API**
   - OpenAPI/Swagger
   - gRPC proto files

4. **Implementar Health Checks**
   - Cada service tem /health
   - Independente de outros services

5. **CI/CD por Service**
   - Pipeline por service
   - Deploy independente

---

## 📚 REFERÊNCIAS

- [Microservices Patterns](https://microservices.io/patterns/index.html)
- [12-Factor App](https://12factor.net/)
- [Go Modules in Monorepos](https://go.dev/doc/modules/managing-dependencies)
- [Domain-Driven Design](https://martinfowler.com/tags/domain%20driven%20design.html)

---

**Data:** 2026-02-05  
**Status:** 🚨 **AÇÃO NECESSÁRIA** - Config centralizado viola independência de microservices
