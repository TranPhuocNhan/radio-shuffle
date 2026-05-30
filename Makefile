.PHONY: run verify test test-unit test-integration integration-up integration-down tidy sqlc migrate-up migrate-down
include .env
export
run:
	go run ./cmd/api

verify:
	go build ./...
	go vet ./...
	go test -race ./...

test:
	go test -race ./...

test-unit:
	go test -race ./internal/... ./cmd/... ./pkg/...

test-integration:
	go test -race ./test/integration/...

integration-up:
	docker compose up -d postgres rabbitmq

integration-down:
	docker compose down

tidy:
	go mod tidy

sqlc:
	cd db && sqlc generate

migrate-up:
	migrate -path db/migrations -database "$(DATABASE_URL)" up

migrate-down:
	migrate -path db/migrations -database "$(DATABASE_URL)" down 1
