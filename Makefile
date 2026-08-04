.PHONY:

run:
	go run cmd/maestro/main.go

migrate-create:
	@if [ -z "$(seq)" ]; then \
		echo "Отсутствует необходимый параметр seq. Пример: make migrate-create seq=init"; \
		exit 1; \
	fi; \
	migrate create \
	-ext sql \
	-dir migrations \
	-seq "$(seq)"

docker-up:
	@docker compose up -d --build

docker-down:
	@docker compose down

help: ## Показать справку по командам
	@echo "Доступные команды:"
	@awk 'BEGIN {FS = ":.*## "}; /^[a-zA-Z%_-]+:.*## / {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}' $(MAKEFILE_LIST) | sort