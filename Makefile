.PHONY: up
up:
	docker compose up --force-recreate --build -d

.PHONY: verbose-up
verbose-up:
	docker compose up --force-recreate --build

.PHONY: down
down:
	docker compose down -v

.PHONY: db-create-migration
db-create-migration:
	bin/goose -dir ./migrations/ create -s $(migration_name) sql
	# make db-create-migration migration_name=test_create пример использования

.PHONY: db-migrate-up
db-migrate-up:
	bin/goose -dir ./migrations/ postgres postgresql://intern_hiring_user:intern_hiring_password@0.0.0.0:5433/intern_hiring_db?sslmode=disable up

.PHONY: db-migrate-down
db-migrate-down:
	bin/goose -dir ./migrations/ postgres postgresql://intern_hiring_user:intern_hiring_password@0.0.0.0:5433/intern_hiring_db?sslmode=disable down

.PHONY: db-migrate-down-all
db-migrate-down-all:
	bin/goose -dir ./migrations/ postgres postgresql://intern_hiring_user:intern_hiring_password@0.0.0.0:5433/intern_hiring_db?sslmode=disable down-to 0

.PHONY: restart
restart: down up

.PHONY: clean
clean:
	docker compose down -v --remove-orphans
	docker builder prune -f

.PHONY: logs
logs:
	docker compose logs -f --tail=150

.PHONY: bin-deps
bin-deps:
	@mkdir -p bin
	@if [ ! -f bin/goose ]; then \
		echo "Installing goose..."; \
        go install github.com/pressly/goose/v3/cmd/goose@latest; \
        cp $$GOPATH/bin/goose bin/; \
    else \
        echo "goose already installed in ./bin"; \
    fi
