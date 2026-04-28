SHELL := /usr/bin/env bash
COMPOSE ?= docker compose

.DEFAULT_GOAL := help

.PHONY: build up down logs lint test proto migrate clean config help

build: ## Build Docker images for implemented services
	$(COMPOSE) build

up: ## Start project services in the background
	$(COMPOSE) up -d --build

down: ## Stop project services
	$(COMPOSE) down

logs: ## Tail service logs
	$(COMPOSE) logs -f --tail=200

test: ## Run available test suites
	bash shared/scripts/test.sh

lint: ## Run available linters
	bash shared/scripts/lint.sh

proto: ## Generate gRPC stubs where toolchains are available
	bash shared/scripts/gen_proto.sh

migrate: ## Run database migrations
	bash shared/scripts/seed_db.sh

clean: ## Stop services and remove compose volumes
	$(COMPOSE) down -v

config: ## Validate rendered Docker Compose configuration
	$(COMPOSE) config

help: ## Show this help
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}'
