-include .env
export

.DEFAULT_GOAL := help

.PHONY: build run migrate-action migrate-create jmx-exporter docker-up docker-down ps lint lint-fix lint-actions lint-dockerfile lint-trivy mocks test test-integration validate-swagger

build:
	@go build -o bin/maestro ./cmd/maestro

run: build
	@./bin/maestro

JMX_EXPORTER_VERSION := 1.6.0
JMX_EXPORTER_SHA256 := a95983fd96e865d2bcdf911cc500e7c82808c27ab9fd226bf96732b6c3d8c46e

jmx-exporter: ## Download jmx_exporter agent jar
	@mkdir -p deploy/jmx-exporter
	@curl -fL --retry 3 --retry-delay 2 -o deploy/jmx-exporter/jmx_prometheus_javaagent.jar \
		https://github.com/prometheus/jmx_exporter/releases/download/$(JMX_EXPORTER_VERSION)/jmx_prometheus_javaagent-$(JMX_EXPORTER_VERSION).jar
	@echo "$(JMX_EXPORTER_SHA256)  deploy/jmx-exporter/jmx_prometheus_javaagent.jar" | shasum -a 256 -c -

migrate-create: ## PostgreSQL: Create a new schema version
	@if [ -z "$(seq)" ]; then \
		echo "Missing required parameter seq. Example: make migrate-create seq=init"; \
		exit 1; \
	fi; \
	docker compose run --rm migrate \
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
	docker compose run --rm migrate \
		-path /migrations \
		-database postgres://${POSTGRES_USER}:${POSTGRES_PASSWORD}@postgres:5432/${POSTGRES_DB}?sslmode=disable \
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

lint-actions: ## Lint GitHub Actions workflows with actionlint
	@docker compose run --rm actionlint -color

lint-dockerfile: ## Lint Dockerfile with hadolint
	@docker compose run --rm hadolint cmd/maestro/Dockerfile

lint-trivy: ## Scan filesystem for vulnerabilities with Trivy
	@docker compose run --rm trivy fs --severity HIGH,CRITICAL --ignore-unfixed --exit-code 1 .

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