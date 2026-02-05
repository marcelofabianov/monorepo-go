# Cache DIP Refactor - Dependency Inversion Principle

## Problema Identificado

O `pkg/cache` tinha um **vazamento de abstração crítico** que violava o Dependency Inversion Principle (DIP):

```go
// ANTES - Violação do DIP
type Cache struct {
    client *redis.Client
    config *Config // <--- Dependência de implementação concreta
    logger *slog.Logger
}
```

### Por que isso era um problema?

1. **Dependência Circular Potencial**: Se o módulo `config` precisasse instanciar cache via `GetCacheAdapter()`, haveria ciclo de dependência (config → cache → config)

2. **Violação do DIP**: Uma biblioteca de infraestrutura (pkg) **não deveria conhecer** as structs de configuração global do sistema

3. **Baixa Testabilidade**: Difícil criar mocks sem trazer toda a estrutura do módulo config

4. **Acoplamento Alto**: pkg/cache estava fortemente acoplado ao módulo config

## Solução Implementada

Implementamos o **Dependency Inversion Principle** através da interface `ConfigProvider`:

### 1. Interface Local no pkg/cache

```go
// ConfigProvider defines what the cache needs to operate.
// This interface decouples pkg/cache from the global config module.
type ConfigProvider interface {
    GetHost() string
    GetPort() int
    GetPassword() string
    GetDB() int
    GetMaxIdleConns() int
    GetMaxActiveConns() int
    GetQueryTimeout() time.Duration
    GetExecTimeout() time.Duration
    GetBackoffMin() time.Duration
    GetBackoffMax() time.Duration
    GetBackoffFactor() int
    GetBackoffJitter() bool
    GetBackoffRetries() int
}
```

### 2. Cache Depende da Interface

```go
// DEPOIS - Conforme DIP
type Cache struct {
    client *redis.Client
    config ConfigProvider // <--- Dependência de abstração
    logger *slog.Logger
}

func New(cfg ConfigProvider) (*Cache, error) {
    if cfg == nil {
        return nil, ErrInvalidConfig
    }
    return &Cache{
        config: cfg,
        logger: slog.Default(),
    }, nil
}
```

### 3. Config Implementa Interface Implicitamente

A struct `cache.Config` implementa `ConfigProvider` através de métodos:

```go
func (c *Config) GetHost() string { return c.Redis.Credentials.Host }
func (c *Config) GetPort() int { return c.Redis.Credentials.Port }
// ... todos os outros métodos
```

## Mudanças Realizadas

### Arquivos Modificados

1. **pkg/cache/config.go**
   - ✅ Adicionada interface `ConfigProvider` com 13 métodos
   - ✅ Implementados todos os métodos em `*Config`
   - ✅ Adicionada verificação em tempo de compilação: `var _ ConfigProvider = (*Config)(nil)`
   - ✅ Mantidas funções de adaptação: `NewConfigFromPort()` e `NewConfigFromCentral()`

2. **pkg/cache/cache.go**
   - ✅ Alterado campo `config *Config` para `config ConfigProvider`
   - ✅ Alterado `func New(*Config)` para `func New(ConfigProvider)`
   - ✅ Substituídos todos os acessos `c.config.Redis.X` por `c.config.GetX()`
   - ✅ Adicionado método `getRetryConfig()` interno para converter config em retry.Config

3. **pkg/cache/README.md**
   - ✅ Atualizado para documentar interface `ConfigProvider`
   - ✅ Adicionados 3 padrões de uso: Interface (recomendado), Concrete, Manual
   - ✅ Explicadas vantagens do DIP
   - ✅ Exemplo de mock para testes

## Benefícios Obtidos

### 1. Desacoplamento Total
- pkg/cache não importa o módulo config
- pkg/cache não conhece a estrutura global de configuração
- Dependência apenas da interface local

### 2. Testabilidade
```go
// Antes: Precisava criar toda a estrutura config
cfg := &cache.Config{
    Redis: cache.RedisConfig{
        Credentials: cache.RedisCredentialsConfig{...},
        Pool: cache.RedisPoolConfig{...},
        Connect: cache.RedisConnectConfig{...},
    },
}

// Depois: Mock simples da interface
type MockConfig struct{}
func (m *MockConfig) GetHost() string { return "localhost" }
// ... implementar apenas os métodos necessários
mock := &MockConfig{}
c, _ := cache.New(mock)
```

### 3. Inversão de Dependência (DIP)

**Antes**: 
```
pkg/cache (high-level) → config (low-level)
```

**Depois**:
```
pkg/cache (high-level) → ConfigProvider (abstraction)
                                ↑
                            config implements
```

### 4. Flexibilidade
Qualquer struct que implemente `ConfigProvider` pode ser usada:
- Mock para testes
- Config em memória
- Config de arquivo JSON
- Config do módulo centralizado

### 5. Evita Dependência Circular
O módulo config pode agora importar e usar cache sem risco de ciclo:
```go
// config/adapters.go
func GetCacheInstance(c *Config) (*cache.Cache, error) {
    port := c.GetRedisPort()
    cacheCfg := cache.NewConfigFromPort(port)
    return cache.New(cacheCfg)
}
```

## Padrão Recomendado

**Use a interface (Padrão 1)**:
```go
cfg, _ := config.Load()
redisPort := cfg.GetRedisPort()      // Retorna RedisConfigPort
cacheCfg := cache.NewConfigFromPort(redisPort)
c, _ := cache.New(cacheCfg)           // cacheCfg implementa ConfigProvider
```

**Por quê?**
- Máximo desacoplamento
- Fácil testar
- Segue princípios SOLID
- Evita dependências circulares

## Testes

Todos os testes continuam passando:
```bash
$ go test ./config/... ./pkg/cache/... ./pkg/logger/...
ok      github.com/marcelofabianov/config    0.002s
ok      github.com/marcelofabianov/cache     0.002s
ok      github.com/marcelofabianov/logger    0.003s
```

## Retrocompatibilidade

✅ **100% compatível** - Nenhuma quebra na API pública:
- `cache.New()` aceita qualquer `ConfigProvider`
- `cache.Config` implementa `ConfigProvider` automaticamente
- `NewConfigFromPort()` e `NewConfigFromCentral()` continuam funcionando
- Código existente continua compilando sem alterações

## Princípios SOLID Aplicados

| Princípio | Como foi aplicado |
|-----------|-------------------|
| **S**RP | Cache foca em operações Redis, config em carregar configurações |
| **O**CP | Extensível via interface ConfigProvider sem modificar cache |
| **L**SP | Qualquer implementação de ConfigProvider pode substituir outra |
| **I**SP | Interface minimalista com apenas o necessário (13 métodos) |
| **D**IP | Cache depende de abstração (ConfigProvider), não de implementação |

## Lições Aprendidas

1. **Sempre prefira interfaces em limites de módulos**: Reduz acoplamento e facilita testes
2. **Defina a interface no consumidor**: pkg/cache define ConfigProvider, não o config
3. **Go's structural typing é poderoso**: Config implementa interface sem declaração explícita
4. **Verificação em tempo de compilação**: `var _ Interface = (*Type)(nil)` garante implementação
5. **DIP não é só sobre abstrações**: É sobre direção de dependências (high → low level)

## Próximos Passos

Este padrão deve ser replicado em outros pacotes:
- [ ] pkg/database → DatabaseConfigProvider
- [ ] pkg/web → HTTPConfigProvider  
- [ ] pkg/auth → AuthConfigProvider

Data: 2026-02-05
Autor: Refatoração baseada em feedback de code review
