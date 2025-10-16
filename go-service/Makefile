.PHONY: help build run test test-verbose test-coverage test-unit test-integration clean lint fmt vet deps docker-build docker-run install-tools benchmark

# Variables
BINARY_NAME=student-report-service
GO=go
GOTEST=$(GO) test
GOCOVER=$(GO) tool cover
GOFMT=gofmt
GOVET=$(GO) vet
GOLINT=golangci-lint
DOCKER_IMAGE=student-report-service
VERSION?=$(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
BUILD_TIME=$(shell date -u '+%Y-%m-%d_%H:%M:%S')
LDFLAGS=-ldflags "-X main.Version=$(VERSION) -X main.BuildTime=$(BUILD_TIME)"

# Colors for terminal output
GREEN=\033[0;32m
YELLOW=\033[0;33m
RED=\033[0;31m
NC=\033[0m # No Color

help: ## Display this help screen
	@echo "$(GREEN)Available targets:$(NC)"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "  $(YELLOW)%-20s$(NC) %s\n", $$1, $$2}'

deps: ## Download dependencies
	@echo "$(GREEN)Downloading dependencies...$(NC)"
	$(GO) mod download
	$(GO) mod verify
	@echo "$(GREEN)✓ Dependencies downloaded$(NC)"

tidy: ## Tidy up dependencies
	@echo "$(GREEN)Tidying up dependencies...$(NC)"
	$(GO) mod tidy
	@echo "$(GREEN)✓ Dependencies tidied$(NC)"

build: ## Build the application
	@echo "$(GREEN)Building $(BINARY_NAME)...$(NC)"
	$(GO) build $(LDFLAGS) -o bin/$(BINARY_NAME) ./cmd/api
	@echo "$(GREEN)✓ Build complete: bin/$(BINARY_NAME)$(NC)"

build-race: ## Build with race detector
	@echo "$(GREEN)Building $(BINARY_NAME) with race detector...$(NC)"
	$(GO) build -race $(LDFLAGS) -o bin/$(BINARY_NAME)-race ./cmd/api
	@echo "$(GREEN)✓ Build complete: bin/$(BINARY_NAME)-race$(NC)"

run: ## Run the application
	@echo "$(GREEN)Running $(BINARY_NAME)...$(NC)"
	$(GO) run ./cmd/api

run-race: ## Run with race detector
	@echo "$(GREEN)Running $(BINARY_NAME) with race detector...$(NC)"
	$(GO) run -race ./cmd/api

test: ## Run all tests
	@echo "$(GREEN)Running tests...$(NC)"
	$(GOTEST) -v ./... -count=1
	@echo "$(GREEN)✓ All tests passed$(NC)"

test-verbose: ## Run tests with verbose output
	@echo "$(GREEN)Running tests with verbose output...$(NC)"
	$(GOTEST) -v -count=1 ./...

test-short: ## Run short tests only
	@echo "$(GREEN)Running short tests...$(NC)"
	$(GOTEST) -short ./... -count=1

test-unit: ## Run unit tests only
	@echo "$(GREEN)Running unit tests...$(NC)"
	$(GOTEST) -v ./internal/service/... ./internal/cache/... ./internal/handler/... -count=1
	@echo "$(GREEN)✓ Unit tests passed$(NC)"

test-coverage: ## Run tests with coverage
	@echo "$(GREEN)Running tests with coverage...$(NC)"
	$(GOTEST) -coverprofile=coverage.out ./internal/service/... ./internal/cache/... ./internal/handler/...
	@echo ""
	@echo "$(YELLOW)Coverage Summary:$(NC)"
	@$(GO) tool cover -func=coverage.out | grep total
	@echo ""
	@echo "$(GREEN)✓ Coverage report generated: coverage.out$(NC)"
	@echo "  Run 'make coverage-html' to view in browser"

coverage-html: test-coverage ## Generate HTML coverage report
	@echo "$(GREEN)Generating HTML coverage report...$(NC)"
	$(GOCOVER) -html=coverage.out -o coverage.html
	@echo "$(GREEN)✓ Coverage HTML generated: coverage.html$(NC)"
	@open coverage.html 2>/dev/null || xdg-open coverage.html 2>/dev/null || echo "Open coverage.html manually in your browser"

coverage-func: test-coverage ## Show function-level coverage
	@echo "$(YELLOW)Function-level coverage:$(NC)"
	@$(GOCOVER) -func=coverage.out

benchmark: ## Run benchmarks
	@echo "$(GREEN)Running benchmarks...$(NC)"
	$(GOTEST) -bench=. -benchmem ./internal/... -run=^$$

benchmark-cpu: ## Run benchmarks with CPU profiling
	@echo "$(GREEN)Running benchmarks with CPU profiling...$(NC)"
	$(GOTEST) -bench=. -benchmem -cpuprofile=cpu.prof ./internal/... -run=^$$
	@echo "$(GREEN)✓ CPU profile generated: cpu.prof$(NC)"
	@echo "  View with: go tool pprof cpu.prof"

benchmark-mem: ## Run benchmarks with memory profiling
	@echo "$(GREEN)Running benchmarks with memory profiling...$(NC)"
	$(GOTEST) -bench=. -benchmem -memprofile=mem.prof ./internal/... -run=^$$
	@echo "$(GREEN)✓ Memory profile generated: mem.prof$(NC)"
	@echo "  View with: go tool pprof mem.prof"

lint: ## Run linter
	@echo "$(GREEN)Running linter...$(NC)"
	@if command -v $(GOLINT) > /dev/null; then \
		$(GOLINT) run ./...; \
		echo "$(GREEN)✓ Linting complete$(NC)"; \
	else \
		echo "$(RED)golangci-lint not installed. Run 'make install-tools'$(NC)"; \
		exit 1; \
	fi

fmt: ## Format code
	@echo "$(GREEN)Formatting code...$(NC)"
	@$(GOFMT) -w -s .
	@echo "$(GREEN)✓ Code formatted$(NC)"

fmt-check: ## Check if code is formatted
	@echo "$(GREEN)Checking code formatting...$(NC)"
	@UNFORMATTED=$$($(GOFMT) -l .); \
	if [ -n "$$UNFORMATTED" ]; then \
		echo "$(RED)The following files are not formatted:$(NC)"; \
		echo "$$UNFORMATTED"; \
		exit 1; \
	else \
		echo "$(GREEN)✓ All files are properly formatted$(NC)"; \
	fi

vet: ## Run go vet
	@echo "$(GREEN)Running go vet...$(NC)"
	@$(GOVET) ./...
	@echo "$(GREEN)✓ Vet complete$(NC)"

check: fmt-check vet lint test ## Run all checks (format, vet, lint, test)
	@echo "$(GREEN)✓ All checks passed$(NC)"

clean: ## Clean build artifacts
	@echo "$(GREEN)Cleaning build artifacts...$(NC)"
	@rm -rf bin/
	@rm -f coverage.out coverage.html
	@rm -f cpu.prof mem.prof
	@rm -rf tmp/
	@$(GO) clean -cache -testcache
	@echo "$(GREEN)✓ Clean complete$(NC)"

install-tools: ## Install development tools
	@echo "$(GREEN)Installing development tools...$(NC)"
	@go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	@go install github.com/swaggo/swag/cmd/swag@latest
	@echo "$(GREEN)✓ Tools installed$(NC)"

swagger: ## Generate Swagger documentation
	@echo "$(GREEN)Generating Swagger documentation...$(NC)"
	@if command -v swag > /dev/null; then \
		swag init -g cmd/api/main.go -o ./docs; \
		echo "$(GREEN)✓ Swagger documentation generated$(NC)"; \
	else \
		echo "$(YELLOW)swag not installed. Run 'make install-tools'$(NC)"; \
	fi

docker-build: ## Build Docker image
	@echo "$(GREEN)Building Docker image...$(NC)"
	docker build -t $(DOCKER_IMAGE):$(VERSION) -t $(DOCKER_IMAGE):latest .
	@echo "$(GREEN)✓ Docker image built: $(DOCKER_IMAGE):$(VERSION)$(NC)"

docker-run: ## Run Docker container
	@echo "$(GREEN)Running Docker container...$(NC)"
	docker run --rm -p 8080:8080 --env-file .env $(DOCKER_IMAGE):latest

docker-clean: ## Remove Docker images
	@echo "$(GREEN)Removing Docker images...$(NC)"
	docker rmi $(DOCKER_IMAGE):$(VERSION) $(DOCKER_IMAGE):latest 2>/dev/null || true
	@echo "$(GREEN)✓ Docker images removed$(NC)"

ci: deps fmt-check vet lint test-coverage ## Run CI pipeline
	@echo "$(GREEN)✓ CI pipeline complete$(NC)"

dev: ## Run in development mode with hot reload (requires air)
	@echo "$(GREEN)Starting development server...$(NC)"
	@if command -v air > /dev/null; then \
		air; \
	else \
		echo "$(YELLOW)'air' not installed. Install with: go install github.com/cosmtrek/air@latest$(NC)"; \
		$(GO) run ./cmd/api; \
	fi

security: ## Run security checks
	@echo "$(GREEN)Running security checks...$(NC)"
	@if command -v gosec > /dev/null; then \
		gosec -quiet ./...; \
		echo "$(GREEN)✓ Security scan complete$(NC)"; \
	else \
		echo "$(YELLOW)'gosec' not installed. Install with: go install github.com/securego/gosec/v2/cmd/gosec@latest$(NC)"; \
	fi

mod-upgrade: ## Upgrade all dependencies to latest versions
	@echo "$(GREEN)Upgrading dependencies...$(NC)"
	$(GO) get -u ./...
	$(GO) mod tidy
	@echo "$(GREEN)✓ Dependencies upgraded$(NC)"

mod-graph: ## Show dependency graph
	@echo "$(YELLOW)Dependency graph:$(NC)"
	@$(GO) mod graph

info: ## Show project information
	@echo "$(YELLOW)Project Information:$(NC)"
	@echo "  Binary name: $(BINARY_NAME)"
	@echo "  Version:     $(VERSION)"
	@echo "  Build time:  $(BUILD_TIME)"
	@echo "  Go version:  $$($(GO) version)"
	@echo ""
	@echo "$(YELLOW)Project Statistics:$(NC)"
	@echo "  Total files:     $$(find . -name '*.go' | wc -l)"
	@echo "  Lines of code:   $$(find . -name '*.go' -exec cat {} \; | wc -l)"
	@echo "  Test files:      $$(find . -name '*_test.go' | wc -l)"

.DEFAULT_GOAL := help
