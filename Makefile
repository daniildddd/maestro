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