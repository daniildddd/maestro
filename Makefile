-include .env
export

.DEFAULT_GOAL := help
.PHONY: build run migrate-action migrate-create docker-up docker-down ps lint lint-fix mocks test test-integration validate-swagger

build:
	@go build -o bin/maestro ./cmd/maestro

run: build
	@./bin/maestro

migrate-create: ## PostgreSQL: Create a new schema version
	@if [ -z "$(seq)" ]; then \
		echo "Missing required parameter seq. Example: make migrate-create seq=init"; \
		exit 1; \
	fi; \
	docker compose run --rm maestro-migrate \
		create \
		-ext sql \
		-dir /migrations \
		-seq "$(seq)"

migrate-up: ## PostgreSQL: Apply migrations
	@make migrate-action action=up

migrate-down: ## PostgreSQL: Roll back migrations
	@make migrate-action action=down

migrate-action:
	@if [ -z "$(action)" ]; then \
		echo "Missing required parameter action. Example: make migrate-action action=up"; \
		exit 1; \
	fi; \
	docker compose run --rm maestro-migrate \
		-path /migrations \
		-database postgres://${POSTGRES_USER}:${POSTGRES_PASSWORD}@maestro-postgres:5432/${POSTGRES_DB}?sslmode=disable \
		"$(action)"

docker-up: ## Start containers
	@docker compose up -d --build

docker-down: ## Stop containers
	@docker compose down

ps: ## Show running Docker Compose services
	@docker compose ps

lint: ## Run the linter
	@golangci-lint run ./...

lint-fix: ## Auto-fix linter issues
	@golangci-lint run --fix ./...

mocks: ## Generate mocks with mockery
	@mockery

test: ## Run unit tests with race detection and coverage
	@go test -race -cover ./...

test-integration: ## Run integration tests against a live Postgres (docker compose up -d)
	@go test -tags integration -race ./...

validate-swagger: ## Validate OpenAPI spec
	@vacuum lint -r vacuum.yml docs/swagger.yaml

help: ## Show available commands
	@echo "Available commands:"
	@awk 'BEGIN {FS = ":.*## "}; /^[a-zA-Z%_-]+:.*## / {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}' $(MAKEFILE_LIST) | sort