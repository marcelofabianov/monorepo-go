# Self-Contained Packages Pattern

## 🎯 Problema Resolvido

**Antes:** Config centralizado criava acoplamento entre services

```
service/course → config → config/adapter → pkg/cache
                                        → pkg/logger
                                        → pkg/database
```

**Problema:** Mudança em 1 pkg afeta TODOS os services

---

## ✅ Solução: Self-Contained Packages

Cada `pkg/` é **completamente independente** com sua própria configuração.

### Arquitetura

```
pkg/logger/
├── config.go          # LoadConfig() com Viper
├── logger.go          # Implementação
├── .env.example       # Configurações exemplo
└── go.mod             # Dependências próprias

pkg/cache/
├── config.go          # LoadConfig() com Viper
├── cache.go           # Implementação
├── .env.example       # Configurações exemplo
└── go.mod             # Dependências próprias

service/course/
├── main.go            # Usa apenas o que precisa
└── .env               # Configurações do service
```

---

## 📊 Comparação

### ❌ ANTES: Config Centralizado

```go
// service/course/main.go
import "github.com/marcelofabianov/config"
import "github.com/marcelofabianov/config/adapter"

cfg, _ := config.Load()  // ❌ Traz TUDO (DB, Redis, JWT...)
log := adapter.NewLoggerInstance(cfg)  // ❌ Via adapter
```

**Problemas:**
- Importa todas as configs (Database, Redis, JWT, Migrations...)
- Dependência transitiva de TODOS os pkgs
- Mudança em config afeta TODOS os services
- Deploy independente impossível

### ✅ DEPOIS: Self-Contained

```go
// service/course/main.go
import "github.com/marcelofabianov/logger"

cfg, _ := logger.LoadConfig()  // ✅ Apenas logger config
log := logger.New(cfg)          // ✅ Direto
```

**Benefícios:**
- Importa APENAS o pkg necessário
- Zero dependências transitivas
- Mudança em logger NÃO afeta cache
- Deploy independente ✅

---

## 🏗️ Padrão de Implementação

### 1. Estrutura do Pacote

```
pkg/example/
├── config.go          # LoadConfig() + Config struct
├── example.go         # Implementação principal
├── .env.example       # Template de configuração
├── config_test.go     # Testes de config
├── example_test.go    # Testes de funcionalidade
├── README.md          # Documentação
└── go.mod             # go mod init github.com/user/example
```

### 2. config.go Template

```go
package example

import (
    "github.com/spf13/viper"
    "os"
    "path/filepath"
    "strings"
)

type Config struct {
    // Configurações específicas do pacote
    Host string
    Port int
}

func LoadConfig() (*Config, error) {
    v := viper.New()
    
    // Buscar .env
    envFile := findEnvFile()
    if envFile != "" {
        v.SetConfigFile(envFile)
        v.SetConfigType("env")
        _ = v.ReadInConfig()
    }
    
    // Environment variables com prefixo
    v.AutomaticEnv()
    v.SetEnvPrefix("EXAMPLE")  // EXAMPLE_HOST, EXAMPLE_PORT
    v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
    
    // Defaults
    v.SetDefault("host", "localhost")
    v.SetDefault("port", 8080)
    
    // Build config
    return &Config{
        Host: v.GetString("host"),
        Port: v.GetInt("port"),
    }, nil
}

func findEnvFile() string {
    dir, _ := os.Getwd()
    for i := 0; i < 5; i++ {
        envPath := filepath.Join(dir, ".env")
        if _, err := os.Stat(envPath); err == nil {
            return envPath
        }
        parent := filepath.Dir(dir)
        if parent == dir {
            break
        }
        dir = parent
    }
    return ".env"
}
```

### 3. .env.example Template

```env
# Example Package Configuration
# Copy to .env and configure

# Host and port
EXAMPLE_HOST=localhost
EXAMPLE_PORT=8080
```

### 4. README.md Template

```markdown
# Example Package

Self-contained package for...

## Quick Start

\```go
cfg, _ := example.LoadConfig()
ex := example.New(cfg)
\```

## Configuration

| Variable | Default | Description |
|----------|---------|-------------|
| `EXAMPLE_HOST` | `localhost` | Server host |
| `EXAMPLE_PORT` | `8080` | Server port |

## Features

- ✅ Self-contained (own config)
- ✅ Zero external dependencies
- ✅ Microservices-ready
```

---

## 🎯 Uso em Services

### Service com Logger

```go
// service/course/main.go
package main

import "github.com/marcelofabianov/logger"

func main() {
    cfg, _ := logger.LoadConfig()  // Lê LOGGER_* vars
    log := logger.New(cfg)
    
    log.Info("Service starting")
}
```

### Service com Cache

```go
// service/course/main.go
package main

import "github.com/marcelofabianov/cache"

func main() {
    cfg, _ := cache.LoadConfig()  // Lê CACHE_* vars
    c, _ := cache.New(cfg)
    
    c.Connect(ctx)
}
```

### Service com Ambos

```go
// service/course/main.go
package main

import (
    "github.com/marcelofabianov/logger"
    "github.com/marcelofabianov/cache"
)

func main() {
    // Cada pkg carrega sua própria config
    logCfg, _ := logger.LoadConfig()
    log := logger.New(logCfg)
    
    cacheCfg, _ := cache.LoadConfig()
    c, _ := cache.New(cacheCfg)
    
    log.Info("Service with cache starting")
}
```

---

## 📋 Convenções

### 1. Prefixo de Environment Variables

Cada pacote usa seu próprio prefixo:

```bash
# Logger
LOGGER_LEVEL=info
LOGGER_ENVIRONMENT=production

# Cache  
CACHE_HOST=redis
CACHE_PORT=6379

# Database
DATABASE_HOST=postgres
DATABASE_PORT=5432
```

### 2. Arquivo .env

- ✅ Cada pkg tem `.env.example` com suas configs
- ✅ Service pode ter `.env` combinando múltiplos pkgs
- ✅ Environment variables têm precedência sobre .env

### 3. Defaults Sensíveis

Cada pkg deve funcionar **out-of-the-box**:

```go
v.SetDefault("level", "info")
v.SetDefault("host", "localhost")
v.SetDefault("port", 8080)
```

### 4. Busca de .env

Buscar em diretório atual e até 5 níveis acima:
- Permite rodar de qualquer subdiretório
- Encontra .env na raiz do workspace
- Fallback para defaults se não encontrar

---

## 🚀 Benefícios para Microservices

### 1. Independência Total

```
service/course    → pkg/logger ✅
                  → pkg/cache  ✅

service/classroom → pkg/logger ✅
                  → pkg/database ✅
```

Cada service escolhe seus pkgs, sem trazer o resto.

### 2. Deploy Independente

- Mudança em `pkg/cache` → rebuild só de services que usam cache
- Mudança em `pkg/logger` → rebuild só de services que usam logger
- Services não afetam uns aos outros ✅

### 3. Testabilidade

```go
// pkg/logger/config_test.go
func TestLoadConfig(t *testing.T) {
    cfg, err := LoadConfig()
    assert.NoError(t, err)
    // Testa em isolamento
}
```

### 4. Reusabilidade

Cada pkg pode ser:
- Extraído para repositório separado
- Versionado independentemente
- Usado em outros projetos
- Publicado como biblioteca

---

## 📊 Checklist de Migração

### Para cada pkg/:

- [ ] Criar `LoadConfig()` com Viper
- [ ] Definir prefixo de env vars único
- [ ] Criar `.env.example`
- [ ] Adicionar defaults sensíveis
- [ ] Implementar `findEnvFile()`
- [ ] Atualizar testes para usar `LoadConfig()`
- [ ] Atualizar README com novo padrão
- [ ] Remover dependência de config global

### Ordem sugerida:

1. ✅ pkg/logger (concluído)
2. pkg/cache
3. pkg/database
4. pkg/web
5. pkg/validation

---

## 🎓 Lições Aprendidas

### ✅ O que funciona

1. **Prefixos únicos**: Evita colisão de variáveis
2. **Defaults sensíveis**: Funciona sem configuração
3. **Busca de .env**: Flexibilidade de execução
4. **Viper**: Poder e simplicidade

### ⚠️ O que evitar

1. ❌ Compartilhar config entre pkgs
2. ❌ Importar pkg de config centralizado
3. ❌ Criar adaptadores centralizados
4. ❌ Coupling via configuração

### 💡 Dicas

1. Documentar **todas** as env vars no .env.example
2. Validar configs críticas
3. Usar `viper.GetString()` com defaults
4. Testar com e sem .env

---

## 📚 Referências

- [12-Factor App - Config](https://12factor.net/config)
- [Viper Documentation](https://github.com/spf13/viper)
- [Go Project Layout](https://github.com/golang-standards/project-layout)

---

**Data:** 2026-02-05  
**Status:** ✅ Implementado em pkg/logger  
**Próximo:** Aplicar em pkg/cache
