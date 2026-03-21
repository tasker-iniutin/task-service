COMPOSE ?= docker compose
GOOSE ?= goose
MIGRATIONS_DIR ?= ./migrations

POSTGRES_USER ?= postgres
POSTGRES_PASSWORD ?= postgres
POSTGRES_DB ?= app
POSTGRES_PORT ?= 5432
DATABASE_URL ?= postgres://$(POSTGRES_USER):$(POSTGRES_PASSWORD)@localhost:$(POSTGRES_PORT)/$(POSTGRES_DB)?sslmode=disable

.PHONY: db-up db-down db-reset db-logs db-psql migrate-up migrate-down migrate-status migrate-create

db-up:
	$(COMPOSE) up -d postgres

db-down:
	$(COMPOSE) down

db-reset:
	$(COMPOSE) down -v

db-logs:
	$(COMPOSE) logs -f postgres

db-psql:
	$(COMPOSE) exec postgres psql -U $(POSTGRES_USER) -d $(POSTGRES_DB)

migrate-up:
	$(GOOSE) -dir $(MIGRATIONS_DIR) postgres "$(DATABASE_URL)" up

migrate-down:
	$(GOOSE) -dir $(MIGRATIONS_DIR) postgres "$(DATABASE_URL)" down

migrate-status:
	$(GOOSE) -dir $(MIGRATIONS_DIR) postgres "$(DATABASE_URL)" status

migrate-create:
ifndef NAME
	$(error NAME is required, use: make migrate-create NAME=create_tasks)
endif
	$(GOOSE) -dir $(MIGRATIONS_DIR) create $(NAME) sql
