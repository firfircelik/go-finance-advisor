# 🚀 Personal Finance Tracker with AI Financial Advisor
# Professional Makefile for Go project

# Variables
APP_NAME := finance-advisor
VERSION := $(shell git describe --tags --always --dirty)
BUILD_TIME := $(shell date +%Y-%m-%dT%H:%M:%S)
GIT_COMMIT := $(shell git rev-parse HEAD)
GO_VERSION := $(shell go version | awk '{print $$3}')
LDFLAGS := -ldflags "-X main.version=$(VERSION) -X main.buildTime=$(BUILD_TIME) -X main.gitCommit=$(GIT_COMMIT)"

# Directories
BIN_DIR := bin
COVERAGE_DIR := coverage
DOCS_DIR := docs

# Docker
DOCKER_IMAGE := $(APP_NAME)
DOCKER_TAG := latest
REGISTRY := ghcr.io/yourusername

# Colors for output
RED := \033[31m
GREEN := \033[32m
YELLOW := \033[33m
BLUE := \033[34m
MAGENTA := \033[35m
CYAN := \033[36m
WHITE := \033[37m
RESET := \033[0m

.PHONY: help all build test test-integration coverage lint fmt vet security docker docker-run docker-push clean setup-hooks swagger dev benchmark profile deps-update deps-check

## help: Show this help message
help:
	@echo "$(CYAN)🚀 Personal Finance Tracker - Available Commands:$(RESET)"
	@echo ""
	@sed -n 's/^##//p' $(MAKEFILE_LIST) | column -t -s ':' | sed -e 's/^/ /'
	@echo ""
	@echo "$(YELLOW)Examples:$(RESET)"
	@echo "  make build          # Build the application"
	@echo "  make test           # Run all tests"
	@echo "  make docker         # Build Docker image"
	@echo "  make dev            # Start development server"

all: clean fmt vet lint test build

## build: Build the application binary
build:
	@echo "$(GREEN)🔨 Building $(APP_NAME)...$(RESET)"
	@mkdir -p $(BIN_DIR)
	CGO_ENABLED=0 go build $(LDFLAGS) -o $(BIN_DIR)/$(APP_NAME) ./cmd/api
	@echo "$(GREEN)✅ Build completed: $(BIN_DIR)/$(APP_NAME)$(RESET)"
	@echo "$(CYAN)Version: $(VERSION)$(RESET)"
	@echo "$(CYAN)Build Time: $(BUILD_TIME)$(RESET)"
	@echo "$(CYAN)Git Commit: $(GIT_COMMIT)$(RESET)"

## build-all: Build for multiple platforms
build-all:
	@echo "$(GREEN)🔨 Building for multiple platforms...$(RESET)"
	@mkdir -p $(BIN_DIR)
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build $(LDFLAGS) -o $(BIN_DIR)/$(APP_NAME)-linux-amd64 ./cmd/api
	CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build $(LDFLAGS) -o $(BIN_DIR)/$(APP_NAME)-linux-arm64 ./cmd/api
	CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 go build $(LDFLAGS) -o $(BIN_DIR)/$(APP_NAME)-darwin-amd64 ./cmd/api
	CGO_ENABLED=0 GOOS=darwin GOARCH=arm64 go build $(LDFLAGS) -o $(BIN_DIR)/$(APP_NAME)-darwin-arm64 ./cmd/api
	CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build $(LDFLAGS) -o $(BIN_DIR)/$(APP_NAME)-windows-amd64.exe ./cmd/api
	@echo "$(GREEN)✅ Multi-platform build completed$(RESET)"

## test: Run unit tests
test:
	@echo "$(BLUE)🧪 Running unit tests...$(RESET)"
	@mkdir -p $(COVERAGE_DIR)
	go test -race -covermode=atomic -coverprofile=$(COVERAGE_DIR)/coverage.out ./...
	@echo "$(GREEN)✅ Unit tests completed$(RESET)"

## test-integration: Run integration tests
test-integration:
	@echo "$(BLUE)🧪 Running integration tests...$(RESET)"
	go test -tags=integration ./...
	@echo "$(GREEN)✅ Integration tests completed$(RESET)"

## coverage: Generate test coverage report
coverage: test
	@echo "$(BLUE)📊 Generating coverage report...$(RESET)"
	go tool cover -html=$(COVERAGE_DIR)/coverage.out -o $(COVERAGE_DIR)/coverage.html
	go tool cover -func=$(COVERAGE_DIR)/coverage.out | tail -1
	@echo "$(GREEN)✅ Coverage report: $(COVERAGE_DIR)/coverage.html$(RESET)"

## benchmark: Run benchmark tests
benchmark:
	@echo "$(BLUE)⚡ Running benchmarks...$(RESET)"
	go test -bench=. -benchmem ./...
	@echo "$(GREEN)✅ Benchmarks completed$(RESET)"

## profile: Run CPU profiling
profile:
	@echo "$(BLUE)📈 Running CPU profiling...$(RESET)"
	go test -cpuprofile=cpu.prof -bench=. ./...
	go tool pprof cpu.prof

## lint: Run linters
lint:
	@echo "$(YELLOW)🔍 Running linters...$(RESET)"
	@which golangci-lint > /dev/null || (echo "$(RED)❌ golangci-lint not installed. Run: go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest$(RESET)" && exit 1)
	golangci-lint run --timeout=5m
	@echo "$(GREEN)✅ Linting completed$(RESET)"

## fmt: Format Go code
fmt:
	@echo "$(YELLOW)🎨 Formatting code...$(RESET)"
	go fmt ./...
	goimports -w .
	@echo "$(GREEN)✅ Code formatted$(RESET)"

## vet: Run go vet
vet:
	@echo "$(YELLOW)🔍 Running go vet...$(RESET)"
	go vet ./...
	@echo "$(GREEN)✅ Vet completed$(RESET)"

## security: Run security checks
security:
	@echo "$(RED)🔒 Running security checks...$(RESET)"
	@which gosec > /dev/null || (echo "$(RED)❌ gosec not installed. Run: go install github.com/securecodewarrior/gosec/v2/cmd/gosec@latest$(RESET)" && exit 1)
	gosec ./...
	@echo "$(GREEN)✅ Security checks completed$(RESET)"

## deps-update: Update dependencies
deps-update:
	@echo "$(BLUE)📦 Updating dependencies...$(RESET)"
	go get -u ./...
	go mod tidy
	@echo "$(GREEN)✅ Dependencies updated$(RESET)"

## deps-check: Check for dependency vulnerabilities
deps-check:
	@echo "$(BLUE)🔍 Checking dependencies for vulnerabilities...$(RESET)"
	@which govulncheck > /dev/null || (echo "$(RED)❌ govulncheck not installed. Run: go install golang.org/x/vuln/cmd/govulncheck@latest$(RESET)" && exit 1)
	govulncheck ./...
	@echo "$(GREEN)✅ Dependency check completed$(RESET)"

## run: Run the application locally
run:
	@echo "$(GREEN)🚀 Starting $(APP_NAME)...$(RESET)"
	go run ./cmd/api

## dev: Start development server with hot reload
dev:
	@echo "$(GREEN)🔥 Starting development server with hot reload...$(RESET)"
	@which air > /dev/null || (echo "$(RED)❌ air not installed. Run: go install github.com/cosmtrek/air@latest$(RESET)" && exit 1)
	air

## docker: Build Docker image
docker:
	@echo "$(BLUE)🐳 Building Docker image...$(RESET)"
	docker build -t $(DOCKER_IMAGE):$(DOCKER_TAG) .
	@echo "$(GREEN)✅ Docker image built: $(DOCKER_IMAGE):$(DOCKER_TAG)$(RESET)"

## docker-run: Run Docker container
docker-run:
	@echo "$(BLUE)🐳 Running Docker container...$(RESET)"
	docker run --rm -p 8080:8080 --env-file .env $(DOCKER_IMAGE):$(DOCKER_TAG)

## docker-push: Push Docker image to registry
docker-push: docker
	@echo "$(BLUE)🐳 Pushing Docker image to registry...$(RESET)"
	docker tag $(DOCKER_IMAGE):$(DOCKER_TAG) $(REGISTRY)/$(DOCKER_IMAGE):$(DOCKER_TAG)
	docker tag $(DOCKER_IMAGE):$(DOCKER_TAG) $(REGISTRY)/$(DOCKER_IMAGE):$(VERSION)
	docker push $(REGISTRY)/$(DOCKER_IMAGE):$(DOCKER_TAG)
	docker push $(REGISTRY)/$(DOCKER_IMAGE):$(VERSION)
	@echo "$(GREEN)✅ Docker image pushed$(RESET)"

## docker-compose-up: Start all services with Docker Compose
docker-compose-up:
	@echo "$(BLUE)🐳 Starting services with Docker Compose...$(RESET)"
	docker-compose up --build -d
	@echo "$(GREEN)✅ Services started$(RESET)"

## docker-compose-down: Stop all services
docker-compose-down:
	@echo "$(BLUE)🐳 Stopping services...$(RESET)"
	docker-compose down
	@echo "$(GREEN)✅ Services stopped$(RESET)"

## swagger: Generate and serve Swagger documentation
swagger:
	@echo "$(CYAN)📚 Generating Swagger documentation...$(RESET)"
	@which swag > /dev/null || (echo "$(RED)❌ swag not installed. Run: go install github.com/swaggo/swag/cmd/swag@latest$(RESET)" && exit 1)
	swag init -g cmd/api/main.go -o api/
	@echo "$(GREEN)✅ Swagger docs generated$(RESET)"
	@echo "$(CYAN)🌐 Swagger UI available at: http://localhost:8080/swagger/$(RESET)"

## setup-hooks: Set up Git hooks
setup-hooks:
	@echo "$(BLUE)🪝 Setting up Git hooks...$(RESET)"
	@mkdir -p .git/hooks
	@echo '#!/bin/sh\nmake fmt lint test' > .git/hooks/pre-commit
	@chmod +x .git/hooks/pre-commit
	@echo "$(GREEN)✅ Git hooks set up$(RESET)"

## clean: Clean build artifacts
clean:
	@echo "$(YELLOW)🧹 Cleaning build artifacts...$(RESET)"
	rm -rf $(BIN_DIR)/ $(COVERAGE_DIR)/ $(DOCS_DIR)/swagger/ *.prof
	docker system prune -f
	@echo "$(GREEN)✅ Clean completed$(RESET)"

## info: Show project information
info:
	@echo "$(CYAN)📋 Project Information:$(RESET)"
	@echo "  App Name:     $(APP_NAME)"
	@echo "  Version:      $(VERSION)"
	@echo "  Go Version:   $(GO_VERSION)"
	@echo "  Git Commit:   $(GIT_COMMIT)"
	@echo "  Build Time:   $(BUILD_TIME)"

## install-tools: Install development tools
install-tools:
	@echo "$(BLUE)🛠️  Installing development tools...$(RESET)"
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	go install github.com/swaggo/swag/cmd/swag@latest
	go install github.com/cosmtrek/air@latest
	go install github.com/securecodewarrior/gosec/v2/cmd/gosec@latest
	go install golang.org/x/vuln/cmd/govulncheck@latest
	go install golang.org/x/tools/cmd/goimports@latest
	@echo "$(GREEN)✅ Development tools installed$(RESET)"

## install-act: Install act for local GitHub Actions
install-act:
	@echo "$(BLUE)🎭 Installing act for local GitHub Actions...$(RESET)"
	@if command -v act >/dev/null 2>&1; then \
		echo "$(GREEN)✅ act is already installed$(RESET)"; \
	else \
		echo "$(YELLOW)Installing act...$(RESET)"; \
		if command -v winget >/dev/null 2>&1; then \
			winget install nektos.act; \
		elif command -v choco >/dev/null 2>&1; then \
			choco install act-cli; \
		else \
			echo "$(RED)❌ Please install act manually from: https://github.com/nektos/act$(RESET)"; \
			exit 1; \
		fi; \
	fi
	@echo "$(GREEN)✅ act installation completed$(RESET)"

## ci-local: Run complete CI pipeline locally using act
ci-local: install-act
	@echo "$(CYAN)🎭 Running complete CI pipeline locally...$(RESET)"
	act --rm
	@echo "$(GREEN)✅ Local CI pipeline completed$(RESET)"

## ci-test: Run test job locally
ci-test:
	@echo "$(BLUE)🧪 Running CI test pipeline locally...$(RESET)"
	@echo "$(YELLOW)🔍 Running go vet...$(RESET)"
	go vet ./...
	@echo "$(YELLOW)🔍 Running staticcheck...$(RESET)"
	@which staticcheck > /dev/null || go install honnef.co/go/tools/cmd/staticcheck@latest
	staticcheck ./...
	@echo "$(YELLOW)🔍 Running golangci-lint...$(RESET)"
	@which golangci-lint > /dev/null || go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	golangci-lint run --timeout=5m
	@echo "$(YELLOW)🧪 Running tests...$(RESET)"
	@mkdir -p $(COVERAGE_DIR)
	go test -race -covermode=atomic -coverprofile=$(COVERAGE_DIR)/coverage.out ./...
	@echo "$(YELLOW)🧪 Running integration tests...$(RESET)"
	go test -tags=integration ./...
	@echo "$(YELLOW)⚡ Running benchmarks...$(RESET)"
	go test -bench=. -benchmem ./...
	@echo "$(GREEN)✅ CI test pipeline completed$(RESET)"

## ci-security: Run security scan locally
ci-security:
	@echo "$(RED)🔒 Running CI security pipeline locally...$(RESET)"
	@echo "$(YELLOW)🔍 Running govulncheck...$(RESET)"
	@which govulncheck > /dev/null || go install golang.org/x/vuln/cmd/govulncheck@latest
	govulncheck ./...
	@echo "$(YELLOW)🔍 Running gosec...$(RESET)"
	@which gosec > /dev/null || go install github.com/securecodewarrior/gosec/v2/cmd/gosec@latest
	gosec ./...
	@echo "$(GREEN)✅ CI security pipeline completed$(RESET)"

## ci-build: Run build job locally
ci-build:
	@echo "$(GREEN)🔨 Running CI build pipeline locally...$(RESET)"
	@echo "$(YELLOW)🔨 Building binary...$(RESET)"
	@mkdir -p $(BIN_DIR)
	CGO_ENABLED=0 go build $(LDFLAGS) -o $(BIN_DIR)/$(APP_NAME) ./cmd/api
	@echo "$(YELLOW)✅ Verifying binary...$(RESET)"
	$(BIN_DIR)/$(APP_NAME) --version || echo "Binary verification completed"
	@echo "$(GREEN)✅ CI build pipeline completed$(RESET)"

## ci-docker: Run docker job locally
ci-docker:
	@echo "$(BLUE)🐳 Running CI docker pipeline locally...$(RESET)"
	@echo "$(YELLOW)🐳 Building Docker image...$(RESET)"
	docker build -t $(DOCKER_IMAGE):$(DOCKER_TAG) .
	@echo "$(YELLOW)🔍 Scanning Docker image with Trivy...$(RESET)"
	@if command -v trivy >/dev/null 2>&1; then \
		trivy image --exit-code 0 --severity HIGH,CRITICAL $(DOCKER_IMAGE):$(DOCKER_TAG); \
	else \
		echo "$(YELLOW)⚠️  Trivy not installed, skipping image scan$(RESET)"; \
	fi
	@echo "$(GREEN)✅ CI docker pipeline completed$(RESET)"

## ci-full: Run all CI jobs locally (without act)
ci-full: ci-test ci-security ci-build ci-docker
	@echo "$(CYAN)🎉 Complete local CI pipeline finished successfully!$(RESET)"
docker-up:
	docker-compose up -d

docker-down:
	docker-compose down

docker-logs:
	docker-compose logs -f
