.PHONY: help dev dev-backend logs db-setup db-seed db-reset go-run go-build go-test go-fmt node-install node-run node-test docker-up docker-down docker-build docker-clean docker-restart test-pdf health clean

# Colors for output
GREEN=\033[0;32m
YELLOW=\033[1;33m
NC=\033[0m # No Color

help: ## Show this help message
	@echo '${GREEN}Available targets:${NC}'
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "  ${YELLOW}%-20s${NC} %s\n", $$1, $$2}'

# Development targets
dev: node-install go-build ## Start all services locally (without Docker)
	@echo "${GREEN}Starting all services locally...${NC}"
	@make -j3 dev-postgres dev-backend dev-go

dev-backend: ## Start only Node.js backend
	@echo "${GREEN}Starting Node.js backend...${NC}"
	cd backend && npm run dev

dev-go: ## Start Go service locally
	@echo "${GREEN}Starting Go service...${NC}"
	cd go-service && go run cmd/api/main.go

logs: ## Show Docker container logs
	docker-compose logs -f

# Database targets
db-setup: ## Initialize database with Docker
	@echo "${GREEN}Setting up database...${NC}"
	docker-compose up -d postgres
	@echo "Waiting for database to be ready..."
	@sleep 5

db-seed: db-setup ## Seed database with test data
	@echo "${GREEN}Database will be seeded on first startup via seed_db directory${NC}"

db-reset: ## Reset database (drop and recreate)
	@echo "${YELLOW}Resetting database...${NC}"
	docker-compose down -v
	docker-compose up -d postgres

# Go service targets
go-run: ## Run Go service directly
	@echo "${GREEN}Running Go service...${NC}"
	cd go-service && go run cmd/api/main.go

go-build: ## Build Go binary
	@echo "${GREEN}Building Go service...${NC}"
	cd go-service && go build -o bin/server cmd/api/main.go

go-test: ## Run Go tests
	@echo "${GREEN}Running Go tests...${NC}"
	cd go-service && go test -v ./...

go-fmt: ## Format Go code
	@echo "${GREEN}Formatting Go code...${NC}"
	cd go-service && go fmt ./...

# Node.js backend targets
node-install: ## Install Node.js dependencies
	@echo "${GREEN}Installing Node.js dependencies...${NC}"
	cd backend && npm install

node-run: ## Run Node.js backend
	@echo "${GREEN}Running Node.js backend...${NC}"
	cd backend && npm start

node-test: ## Run Node.js tests
	@echo "${GREEN}Running Node.js tests...${NC}"
	cd backend && npm test

# Docker targets
docker-up: ## Start all containers
	@echo "${GREEN}Starting all Docker containers...${NC}"
	docker-compose up -d
	@echo "${GREEN}Containers started! Services available at:${NC}"
	@echo "  - Backend:     http://localhost:5007"
	@echo "  - Go Service:  http://localhost:8080"
	@echo "  - PostgreSQL:  localhost:5432"

docker-down: ## Stop all containers
	@echo "${YELLOW}Stopping all Docker containers...${NC}"
	docker-compose down

docker-build: ## Build Docker images
	@echo "${GREEN}Building Docker images...${NC}"
	docker-compose build --no-cache

docker-clean: ## Remove containers, volumes, and images
	@echo "${YELLOW}Cleaning up Docker resources...${NC}"
	docker-compose down -v --rmi all

docker-restart: docker-down docker-build docker-up ## Rebuild and restart all services

# Testing targets
test-pdf: ## Test PDF generation endpoint (requires running services)
	@echo "${GREEN}Testing PDF generation...${NC}"
	@echo "Downloading PDF report for student ID 1..."
	curl -O -J http://localhost:8080/api/v1/students/1/report
	@echo "\n${GREEN}PDF downloaded successfully!${NC}"

health: ## Check health of all services
	@echo "${GREEN}Checking service health...${NC}"
	@echo "Backend health:"
	@curl -s http://localhost:5007/health || echo "${YELLOW}Backend not responding${NC}"
	@echo "\nGo service health:"
	@curl -s http://localhost:8080/health | json_pp || echo "${YELLOW}Go service not responding${NC}"

# Utility targets
clean: ## Clean up build artifacts and temporary files
	@echo "${YELLOW}Cleaning up...${NC}"
	rm -rf go-service/bin
	rm -rf backend/node_modules
	rm -f *.pdf

# Setup target for first-time setup
setup: ## First-time setup (install dependencies, build)
	@echo "${GREEN}Running first-time setup...${NC}"
	make node-install
	make go-build
	@echo "${GREEN}Setup complete! Run 'make docker-up' to start services${NC}"
