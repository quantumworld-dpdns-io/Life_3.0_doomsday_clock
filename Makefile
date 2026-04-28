.PHONY: proto build up down logs lint test migrate clean help

proto: ## Generate gRPC stubs for all services
	bash shared/scripts/gen_proto.sh

build: ## Build all Docker images
	docker compose build

up: ## Start all services in background
	docker compose up -d

down: ## Stop all services
	docker compose down

logs: ## Tail all service logs
	docker compose logs -f

lint: ## Lint all services
	@echo "--- Go ---" && cd services/api-gateway && golangci-lint run ./...
	@echo "--- Python intelligence ---" && cd services/intelligence && ruff check .
	@echo "--- Python quantum-sim ---" && cd services/quantum-sim && ruff check .
	@echo "--- Rust ---" && cd services/risk-engine && cargo clippy -- -D warnings
	@echo "--- Frontend ---" && cd apps/web && npm run lint

test: ## Run all unit tests
	@echo "--- Go ---" && cd services/api-gateway && go test ./...
	@echo "--- Python intelligence ---" && cd services/intelligence && pytest
	@echo "--- Python quantum-sim ---" && cd services/quantum-sim && pytest
	@echo "--- Rust ---" && cd services/risk-engine && cargo test
	@echo "--- Frontend ---" && cd apps/web && npm test -- --run

migrate: ## Run database migrations
	bash shared/scripts/seed_db.sh

clean: ## Remove build artifacts
	docker compose down -v
	cd services/risk-engine && cargo clean
	cd apps/web && rm -rf dist node_modules

help: ## Show this help
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}'
