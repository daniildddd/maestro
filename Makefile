-include .env
export

.DEFAULT_GOAL := help
.PHONY: build run migrate-create docker-up docker-down lint lint-fix help test

build:
	@go build -o bin/maestro ./cmd/maestro

run: build
	@./bin/maestro

migrate-create: ## Create a migration (make migrate-create seq=init)
	@if [ -z "$(seq)" ]; then \
		echo "Missing required parameter seq. Example: make migrate-create seq=init"; \
		exit 1; \
	fi; \
	migrate create \
	-ext sql \
	-dir migrations \
	-seq "$(seq)"

docker-up: ## Start containers
	@docker compose up -d --build

docker-down: ## Stop containers
	@docker compose down

lint: ## Run the linter
	@golangci-lint run ./...

lint-fix: ## Auto-fix linter issues
	@golangci-lint run --fix ./...

test: ## Run unit tests with race detection and coverage
	@go test -race -cover ./...

test-integration:
	@go test -tags integration -race ./...

help: ## Show available commands
	@echo "Available commands:"
	@awk 'BEGIN {FS = ":.*## "}; /^[a-zA-Z%_-]+:.*## / {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}' $(MAKEFILE_LIST) | sort