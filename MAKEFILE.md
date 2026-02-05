# 🚀 Microservices Makefile

Makefile para gerenciar todos os microservices do monorepo.

## 📋 Comandos Principais

### Quick Start

```bash
# Setup + Start all services
make up

# Stop + Clean
make down

# Restart all services  
make restart

# Check status
make status
```

### Build

```bash
# Build all services
make build

# Build individual service
make build-course
make build-classroom
make build-lesson
make build-enrollment
```

### Run Services

**Background (todos os services):**
```bash
make run-all
```

**Foreground (development - um por vez):**
```bash
make run-course       # Port 8080
make run-classroom    # Port 8081
make run-lesson       # Port 8082
make run-enrollment   # Port 8083
```

### Health & Status

```bash
# Check health of all services
make health

# View service info
make info

# Check status (alias for health)
make status
```

### Logs

```bash
# View logs of all services
make logs

# Follow logs of specific service
make logs-course
make logs-classroom
make logs-lesson
make logs-enrollment
```

### Stop Services

```bash
# Stop all services
make stop
```

### Test

```bash
# Run all tests
make test

# Test specific package
make test-web
make test-logger

# Test all packages
make test-pkg
```

### Clean

```bash
# Clean build artifacts and logs
make clean

# Clean Go cache
make clean-cache
```

### Development

```bash
# Setup development environment
make dev

# Update dependencies
make deps

# Format code
make fmt

# Run linters
make lint
```

### Help

```bash
# Show all available commands
make help
```

## 🌐 Service Ports

| Service    | Port | URL                        |
|------------|------|----------------------------|
| course     | 8080 | http://localhost:8080      |
| classroom  | 8081 | http://localhost:8081      |
| lesson     | 8082 | http://localhost:8082      |
| enrollment | 8083 | http://localhost:8083      |

## 📍 Endpoints

Todos os services têm os seguintes endpoints:

- `GET /` - Service info
- `GET /health` - Liveness probe
- `GET /health/ready` - Readiness probe

## 🔧 Exemplos de Uso

### Iniciar ambiente de desenvolvimento

```bash
# 1. Setup inicial
make dev

# 2. Buildar tudo
make build

# 3. Iniciar todos services
make run-all

# 4. Verificar se estão rodando
make health

# 5. Ver logs
make logs
```

### Desenvolvimento de um service específico

```bash
# Terminal 1: Run service em foreground
make run-course

# Terminal 2: Fazer requests
curl http://localhost:8080/
curl http://localhost:8080/health
```

### Parar tudo e limpar

```bash
make down
```

## 📁 Estrutura de Arquivos

Todos os arquivos temporários são organizados em `tmp/`:

```
tmp/
├── log/                # Logs dos services
│   ├── course.log
│   ├── classroom.log
│   ├── lesson.log
│   └── enrollment.log
│
└── pid/                # PID files dos services em background
    ├── course.pid
    ├── classroom.pid
    ├── lesson.pid
    └── enrollment.pid
```

## 🛠️ Troubleshooting

### Services não iniciam

```bash
# Verificar se portas estão em uso
lsof -i :8080-8083

# Parar tudo
make stop

# Limpar e tentar novamente
make clean
make up
```

### Erro de dependências

```bash
# Atualizar todas dependências
make deps

# Limpar cache e rebuildar
make clean-cache
make build
```

### Ver erros nos logs

```bash
# Ver últimas linhas de todos logs
make logs

# Follow log específico
make logs-course
```

## 🎯 Workflow Comum

### Desenvolvimento

```bash
make up         # Inicia tudo
make logs       # Verifica se está OK
# Desenvolver...
make restart    # Após mudanças
make down       # Quando terminar
```

### CI/CD

```bash
make deps       # Baixar dependências
make lint       # Linters
make test       # Tests
make build      # Build all
```

## 💡 Tips

1. **Use `make help`** para ver todos comandos disponíveis
2. **`make up`** é o jeito mais rápido de iniciar tudo
3. **`make down`** limpa tudo ao parar
4. **`make health`** para verificar se services estão OK
5. **`make logs`** para debug rápido

## 🔗 Mais Documentação

- **pkg/web:** Ver `pkg/web/USAGE.md`
- **Middlewares:** Ver `pkg/web/middleware/USAGE.md`
- **Services:** Ver `service/*/` para cada microservice
