.PHONY: help build test clean run-all run-course run-classroom run-lesson run-enrollment stop health

# Colors for output (using tput for better compatibility)
RED := $(shell tput setaf 1 2>/dev/null || echo '')
GREEN := $(shell tput setaf 2 2>/dev/null || echo '')
YELLOW := $(shell tput setaf 3 2>/dev/null || echo '')
BLUE := $(shell tput setaf 4 2>/dev/null || echo '')
BOLD := $(shell tput bold 2>/dev/null || echo '')
NC := $(shell tput sgr0 2>/dev/null || echo '')

# Service ports
PORT_COURSE=8080
PORT_CLASSROOM=8081
PORT_LESSON=8082
PORT_ENROLLMENT=8083

# Directories
PID_DIR=tmp/pid
LOGS_DIR=tmp/log

# Default target
help: ## Show this help message
	@echo "$(BLUE)════════════════════════════════════════════════════════════════$(NC)"
	@echo "$(GREEN)  📦 Microservices Management$(NC)"
	@echo "$(BLUE)════════════════════════════════════════════════════════════════$(NC)"
	@echo ""
	@echo "$(YELLOW)Available targets:$(NC)"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "  $(GREEN)%-20s$(NC) %s\n", $$1, $$2}'
	@echo ""
	@echo "$(YELLOW)Service Ports:$(NC)"
	@echo "  $(GREEN)course$(NC)      → http://localhost:$(PORT_COURSE)"
	@echo "  $(GREEN)classroom$(NC)   → http://localhost:$(PORT_CLASSROOM)"
	@echo "  $(GREEN)lesson$(NC)      → http://localhost:$(PORT_LESSON)"
	@echo "  $(GREEN)enrollment$(NC)  → http://localhost:$(PORT_ENROLLMENT)"
	@echo ""

# Build targets
build: build-course build-classroom build-lesson build-enrollment ## Build all services
	@echo "$(GREEN)✅ All services built successfully!$(NC)"

build-course: ## Build course service
	@echo "$(BLUE)🔨 Building course service...$(NC)"
	@cd service/course && go build -o bin/course cmd/api/main.go
	@echo "$(GREEN)✅ course built$(NC)"

build-classroom: ## Build classroom service
	@echo "$(BLUE)🔨 Building classroom service...$(NC)"
	@cd service/classroom && go build -o bin/classroom cmd/api/main.go
	@echo "$(GREEN)✅ classroom built$(NC)"

build-lesson: ## Build lesson service
	@echo "$(BLUE)🔨 Building lesson service...$(NC)"
	@cd service/lesson && go build -o bin/lesson cmd/api/main.go
	@echo "$(GREEN)✅ lesson built$(NC)"

build-enrollment: ## Build enrollment service
	@echo "$(BLUE)🔨 Building enrollment service...$(NC)"
	@cd service/enrollment && go build -o bin/enrollment cmd/api/main.go
	@echo "$(GREEN)✅ enrollment built$(NC)"

# Test targets
test: ## Run all tests
	@echo "$(BLUE)🧪 Running all tests...$(NC)"
	@go test ./pkg/... -v -cover
	@echo "$(GREEN)✅ All tests passed!$(NC)"

test-pkg: ## Test all packages
	@echo "$(BLUE)🧪 Testing packages...$(NC)"
	@go test ./pkg/... -v

test-web: ## Test web package
	@echo "$(BLUE)🧪 Testing pkg/web...$(NC)"
	@cd pkg/web && go test ./... -v

test-logger: ## Test logger package
	@echo "$(BLUE)🧪 Testing pkg/logger...$(NC)"
	@cd pkg/logger && go test ./... -v

# Run targets - Background
run-all: ## Start all services in background
	@echo "$(BLUE)🚀 Starting all services...$(NC)"
	@make run-course-bg
	@make run-classroom-bg
	@make run-lesson-bg
	@make run-enrollment-bg
	@sleep 2
	@make health
	@echo ""
	@echo "$(GREEN)✅ All services started!$(NC)"
	@echo ""
	@echo "$(YELLOW)Services running at:$(NC)"
	@echo "  course:     http://localhost:$(PORT_COURSE)"
	@echo "  classroom:  http://localhost:$(PORT_CLASSROOM)"
	@echo "  lesson:     http://localhost:$(PORT_LESSON)"
	@echo "  enrollment: http://localhost:$(PORT_ENROLLMENT)"
	@echo ""
	@echo "$(YELLOW)Commands:$(NC)"
	@echo "  make health → Check health of all services"
	@echo "  make logs   → View logs of all services"
	@echo "  make stop   → Stop all services"

run-course-bg: ## Start course service in background
	@echo "$(BLUE)▶ Starting course service on port $(PORT_COURSE)...$(NC)"
	@mkdir -p $(PID_DIR) $(LOGS_DIR)
	@WEB_HTTP_PORT=$(PORT_COURSE) nohup go run service/course/cmd/api/main.go > $(LOGS_DIR)/course.log 2>&1 & echo $$! > $(PID_DIR)/course.pid
	@sleep 1
	@echo "$(GREEN)✅ course started (PID: $$(cat $(PID_DIR)/course.pid))$(NC)"

run-classroom-bg: ## Start classroom service in background
	@echo "$(BLUE)▶ Starting classroom service on port $(PORT_CLASSROOM)...$(NC)"
	@mkdir -p $(PID_DIR) $(LOGS_DIR)
	@WEB_HTTP_PORT=$(PORT_CLASSROOM) nohup go run service/classroom/cmd/api/main.go > $(LOGS_DIR)/classroom.log 2>&1 & echo $$! > $(PID_DIR)/classroom.pid
	@sleep 1
	@echo "$(GREEN)✅ classroom started (PID: $$(cat $(PID_DIR)/classroom.pid))$(NC)"

run-lesson-bg: ## Start lesson service in background
	@echo "$(BLUE)▶ Starting lesson service on port $(PORT_LESSON)...$(NC)"
	@mkdir -p $(PID_DIR) $(LOGS_DIR)
	@WEB_HTTP_PORT=$(PORT_LESSON) nohup go run service/lesson/cmd/api/main.go > $(LOGS_DIR)/lesson.log 2>&1 & echo $$! > $(PID_DIR)/lesson.pid
	@sleep 1
	@echo "$(GREEN)✅ lesson started (PID: $$(cat $(PID_DIR)/lesson.pid))$(NC)"

run-enrollment-bg: ## Start enrollment service in background
	@echo "$(BLUE)▶ Starting enrollment service on port $(PORT_ENROLLMENT)...$(NC)"
	@mkdir -p $(PID_DIR) $(LOGS_DIR)
	@WEB_HTTP_PORT=$(PORT_ENROLLMENT) nohup go run service/enrollment/cmd/api/main.go > $(LOGS_DIR)/enrollment.log 2>&1 & echo $$! > $(PID_DIR)/enrollment.pid
	@sleep 1
	@echo "$(GREEN)✅ enrollment started (PID: $$(cat $(PID_DIR)/enrollment.pid))$(NC)"

# Run targets - Foreground (development)
run-course: ## Run course service in foreground
	@echo "$(BLUE)▶ Running course service on port $(PORT_COURSE)...$(NC)"
	@WEB_HTTP_PORT=$(PORT_COURSE) go run service/course/cmd/api/main.go

run-classroom: ## Run classroom service in foreground
	@echo "$(BLUE)▶ Running classroom service on port $(PORT_CLASSROOM)...$(NC)"
	@WEB_HTTP_PORT=$(PORT_CLASSROOM) go run service/classroom/cmd/api/main.go

run-lesson: ## Run lesson service in foreground
	@echo "$(BLUE)▶ Running lesson service on port $(PORT_LESSON)...$(NC)"
	@WEB_HTTP_PORT=$(PORT_LESSON) go run service/lesson/cmd/api/main.go

run-enrollment: ## Run enrollment service in foreground
	@echo "$(BLUE)▶ Running enrollment service on port $(PORT_ENROLLMENT)...$(NC)"
	@WEB_HTTP_PORT=$(PORT_ENROLLMENT) go run service/enrollment/cmd/api/main.go

# Health check
health: ## Check health of all services
	@echo "$(BLUE)🏥 Checking services health...$(NC)"
	@echo ""
	@-curl -s http://localhost:$(PORT_COURSE)/health > /dev/null 2>&1 && echo "$(GREEN)✅ course ($(PORT_COURSE))$(NC)" || echo "$(RED)❌ course ($(PORT_COURSE))$(NC)"
	@-curl -s http://localhost:$(PORT_CLASSROOM)/health > /dev/null 2>&1 && echo "$(GREEN)✅ classroom ($(PORT_CLASSROOM))$(NC)" || echo "$(RED)❌ classroom ($(PORT_CLASSROOM))$(NC)"
	@-curl -s http://localhost:$(PORT_LESSON)/health > /dev/null 2>&1 && echo "$(GREEN)✅ lesson ($(PORT_LESSON))$(NC)" || echo "$(RED)❌ lesson ($(PORT_LESSON))$(NC)"
	@-curl -s http://localhost:$(PORT_ENROLLMENT)/health > /dev/null 2>&1 && echo "$(GREEN)✅ enrollment ($(PORT_ENROLLMENT))$(NC)" || echo "$(RED)❌ enrollment ($(PORT_ENROLLMENT))$(NC)"

# Logs
logs: ## Show logs of all services
	@echo "$(BLUE)📋 Service Logs:$(NC)"
	@echo ""
	@echo "$(YELLOW)=== COURSE ===$(NC)"
	@tail -n 20 $(LOGS_DIR)/course.log 2>/dev/null || echo "No logs yet"
	@echo ""
	@echo "$(YELLOW)=== CLASSROOM ===$(NC)"
	@tail -n 20 $(LOGS_DIR)/classroom.log 2>/dev/null || echo "No logs yet"
	@echo ""
	@echo "$(YELLOW)=== LESSON ===$(NC)"
	@tail -n 20 $(LOGS_DIR)/lesson.log 2>/dev/null || echo "No logs yet"
	@echo ""
	@echo "$(YELLOW)=== ENROLLMENT ===$(NC)"
	@tail -n 20 $(LOGS_DIR)/enrollment.log 2>/dev/null || echo "No logs yet"

logs-course: ## Show course service logs
	@tail -f $(LOGS_DIR)/course.log

logs-classroom: ## Show classroom service logs
	@tail -f $(LOGS_DIR)/classroom.log

logs-lesson: ## Show lesson service logs
	@tail -f $(LOGS_DIR)/lesson.log

logs-enrollment: ## Show enrollment service logs
	@tail -f $(LOGS_DIR)/enrollment.log

# Stop services
stop: ## Stop all services
	@echo "$(BLUE)🛑 Stopping all services...$(NC)"
	@-[ -f $(PID_DIR)/course.pid ] && kill $$(cat $(PID_DIR)/course.pid) 2>/dev/null && rm $(PID_DIR)/course.pid && echo "$(GREEN)✅ course stopped$(NC)" || echo "$(YELLOW)⚠ course not running$(NC)"
	@-[ -f $(PID_DIR)/classroom.pid ] && kill $$(cat $(PID_DIR)/classroom.pid) 2>/dev/null && rm $(PID_DIR)/classroom.pid && echo "$(GREEN)✅ classroom stopped$(NC)" || echo "$(YELLOW)⚠ classroom not running$(NC)"
	@-[ -f $(PID_DIR)/lesson.pid ] && kill $$(cat $(PID_DIR)/lesson.pid) 2>/dev/null && rm $(PID_DIR)/lesson.pid && echo "$(GREEN)✅ lesson stopped$(NC)" || echo "$(YELLOW)⚠ lesson not running$(NC)"
	@-[ -f $(PID_DIR)/enrollment.pid ] && kill $$(cat $(PID_DIR)/enrollment.pid) 2>/dev/null && rm $(PID_DIR)/enrollment.pid && echo "$(GREEN)✅ enrollment stopped$(NC)" || echo "$(YELLOW)⚠ enrollment not running$(NC)"
	@echo "$(GREEN)✅ All services stopped!$(NC)"

# Clean targets
clean: stop ## Clean build artifacts and logs
	@echo "$(BLUE)🧹 Cleaning...$(NC)"
	@rm -rf service/*/bin
	@rm -rf $(LOGS_DIR)/*.log
	@rm -rf $(PID_DIR)/*.pid
	@find . -name "*.test" -delete
	@find . -name "*.out" -delete
	@echo "$(GREEN)✅ Clean complete!$(NC)"

clean-cache: ## Clean Go module cache
	@echo "$(BLUE)🧹 Cleaning Go cache...$(NC)"
	@go clean -cache -modcache -testcache
	@echo "$(GREEN)✅ Cache cleaned!$(NC)"

# Development helpers
dev: ## Setup development environment
	@echo "$(BLUE)🔧 Setting up development environment...$(NC)"
	@mkdir -p $(LOGS_DIR)
	@mkdir -p $(PID_DIR)
	@mkdir -p service/course/bin
	@mkdir -p service/classroom/bin
	@mkdir -p service/lesson/bin
	@mkdir -p service/enrollment/bin
	@echo "$(GREEN)✅ Development environment ready!$(NC)"

deps: ## Download and tidy dependencies
	@echo "$(BLUE)📦 Downloading dependencies...$(NC)"
	@cd pkg/web && go mod tidy
	@cd pkg/logger && go mod tidy
	@cd pkg/cache && go mod tidy
	@cd pkg/database && go mod tidy
	@cd pkg/retry && go mod tidy
	@cd pkg/validation && go mod tidy
	@cd service/course && go mod tidy
	@cd service/classroom && go mod tidy
	@cd service/lesson && go mod tidy
	@cd service/enrollment && go mod tidy
	@echo "$(GREEN)✅ Dependencies updated!$(NC)"

fmt: ## Format all Go code
	@echo "$(BLUE)✨ Formatting code...$(NC)"
	@go fmt ./...
	@echo "$(GREEN)✅ Code formatted!$(NC)"

lint: ## Run linters
	@echo "$(BLUE)🔍 Running linters...$(NC)"
	@go vet ./...
	@echo "$(GREEN)✅ Linting complete!$(NC)"

# Quick commands
up: dev run-all ## Quick start: setup + run all services

down: stop clean ## Quick stop: stop + clean

restart: stop run-all ## Restart all services

status: health ## Check status of all services

# Info
info: ## Show service information
	@echo "$(BLUE)ℹ️  Service Information:$(NC)"
	@echo ""
	@echo "$(YELLOW)Services:$(NC)"
	@echo "  • course     - Course management"
	@echo "  • classroom  - Classroom management"
	@echo "  • lesson     - Lesson management"
	@echo "  • enrollment - Student enrollment"
	@echo ""
	@echo "$(YELLOW)Ports:$(NC)"
	@echo "  • course:     $(PORT_COURSE)"
	@echo "  • classroom:  $(PORT_CLASSROOM)"
	@echo "  • lesson:     $(PORT_LESSON)"
	@echo "  • enrollment: $(PORT_ENROLLMENT)"
	@echo ""
	@echo "$(YELLOW)Packages:$(NC)"
	@echo "  • pkg/web        - HTTP server + middlewares"
	@echo "  • pkg/logger     - Structured logging"
	@echo "  • pkg/cache      - Redis cache"
	@echo "  • pkg/database   - PostgreSQL"
	@echo "  • pkg/retry      - Retry strategies"
	@echo "  • pkg/validation - Input validation"
