.PHONY: help start stop restart logs db-seed docker-build docker-clean go-test go-fmt node-test

# Colors for output
GREEN=\033[0;32m
YELLOW=\033[1;33m
BLUE=\033[0;34m
RED=\033[0;31m
NC=\033[0m # No Color

help: ## Show this help message
	@echo "${BLUE}=== Student Management System - Available Commands ===${NC}"
	@echo ""
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk -v yellow="${YELLOW}" -v nc="${NC}" 'BEGIN {FS = ":.*?## "}; {printf "  " yellow "%-20s" nc " %s\n", $$1, $$2}'
	@echo ""
	@echo "${BLUE}Quick Start:${NC}"
	@echo "  ${GREEN}make start${NC}        - Complete setup and start (one command!)"
	@echo "  ${GREEN}make logs${NC}         - View live logs"
	@echo "  ${GREEN}make stop${NC}         - Stop all services"

# ==========================================
# Main Commands (Most Used)
# ==========================================

start: ## Complete setup: build images, start services, seed database
	@echo "${BLUE}╔════════════════════════════════════════════════════════╗${NC}"
	@echo "${BLUE}║         Starting Student Management System             ║${NC}"
	@echo "${BLUE}╚════════════════════════════════════════════════════════╝${NC}"
	@echo ""
	@echo "${GREEN}[1/4] Building Docker images...${NC}"
	@docker-compose build
	@echo ""
	@echo "${GREEN}[2/4] Starting containers...${NC}"
	@docker-compose up -d
	@echo ""
	@echo "${GREEN}[3/4] Waiting for services to be ready...${NC}"
	@sleep 5
	@printf "  Waiting for database"
	@until docker exec school_mgmt_db pg_isready -U postgres > /dev/null 2>&1; do \
		printf "."; \
		sleep 1; \
	done
	@echo " ${GREEN}✅${NC}"
	@echo ""
	@echo "${GREEN}[4/4] Seeding database...${NC}"
	@docker exec -i school_mgmt_db psql -U postgres -d school_mgmt < seed_db/tables.sql > /dev/null 2>&1 && echo "  ${GREEN}✅ Tables created${NC}" || echo "  ${RED}❌ Failed to create tables${NC}"
	@docker exec -i school_mgmt_db psql -U postgres -d school_mgmt < seed_db/seed-db.sql > /dev/null 2>&1 && echo "  ${GREEN}✅ Data seeded${NC}" || echo "  ${RED}❌ Failed to seed data${NC}"
	@echo ""
	@echo "${BLUE}╔════════════════════════════════════════════════════════╗${NC}"
	@echo "${BLUE}║              ✅  System Ready!                         ║${NC}"
	@echo "${BLUE}╠════════════════════════════════════════════════════════╣${NC}"
	@echo "${BLUE}║  Services:                                             ║${NC}"
	@echo "${BLUE}║    • Backend:     http://localhost:5007                ║${NC}"
	@echo "${BLUE}║    • Go Service:  http://localhost:8080                ║${NC}"
	@echo "${BLUE}║    • PostgreSQL:  localhost:5432                       ║${NC}"
	@echo "${BLUE}║                                                        ║${NC}"
	@echo "${BLUE}║  Next steps:                                           ║${NC}"
	@echo "${BLUE}║    make logs        - View live logs                   ║${NC}"
	@echo "${BLUE}║    make restart     - Rebuild and restart              ║${NC}"
	@echo "${BLUE}║    make stop        - Stop all services                ║${NC}"
	@echo "${BLUE}╚════════════════════════════════════════════════════════╝${NC}"

stop: ## Stop all services
	@echo "${YELLOW}Stopping all services...${NC}"
	@docker-compose down
	@echo "${GREEN}All services stopped.${NC}"

restart: ## Rebuild and restart all services
	@echo "${YELLOW}Restarting services...${NC}"
	@make stop
	@make start


# ==========================================
# Database Management
# ==========================================

db-seed: ## Seed database with test data
	@echo "${GREEN}Seeding database...${NC}"
	@docker exec -i school_mgmt_db psql -U postgres -d school_mgmt < seed_db/tables.sql > /dev/null 2>&1 && echo "  ${GREEN}✅ Tables created${NC}" || echo "  ${RED}❌ Failed to create tables${NC}"
	@docker exec -i school_mgmt_db psql -U postgres -d school_mgmt < seed_db/seed-db.sql > /dev/null 2>&1 && echo "  ${GREEN}✅ Data seeded${NC}" || echo "  ${RED}❌ Failed to seed data${NC}"

# ==========================================
# Docker Management
# ==========================================

docker-build: ## Build Docker images
	@echo "${GREEN}Building Docker images...${NC}"
	@docker-compose build

docker-clean: ## Remove all containers, volumes, and images
	@echo "${RED}Cleaning up Docker resources...${NC}"
	@docker-compose down -v --rmi all
	@echo "${GREEN}Cleanup complete.${NC}"

logs: ## Show live logs from all containers
	@echo "${GREEN}Showing live logs (Ctrl+C to exit)...${NC}"
	@docker-compose logs -f

# ==========================================
# Development Tools
# ==========================================

go-test: ## Run Go tests
	@echo "${GREEN}Running Go tests...${NC}"
	@cd go-service && go test -v ./...

go-fmt: ## Format Go code
	@echo "${GREEN}Formatting Go code...${NC}"
	@cd go-service && go fmt ./...

node-test: ## Run Node.js tests
	@echo "${GREEN}Running Node.js tests...${NC}"
	@cd backend && npm test

# ==========================================
# Default target
# ==========================================

.DEFAULT_GOAL := help
