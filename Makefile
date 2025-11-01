include .env
MIGRAIONS_PATH=./cmd/migrate/migrations

.PHONY: migrate
migrate:
	goose -dir $(MIGRAIONS_PATH) postgres "$(DB_ADDR)" up

.PHONY: migrate-down
migrate-down:
	goose -dir $(MIGRAIONS_PATH) postgres "$(DB_ADDR)" down

.PHONY: migrate-create
migrate-create:
	goose -dir $(MIGRAIONS_PATH) create $(name) sql

.PHONY: run
run:
	go run ./cmd/server/main.go

.PHONY: gen-docs
gen-docs:
	swag init -g ./cmd/api/main.go -o ./cmd/docs && swag fmt