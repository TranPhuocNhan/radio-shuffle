.PHONY: run test tidy sqlc migrate-up migrate-down
include .env
export
run:
	go run ./cmd/api

test:
	go test -race ./...

tidy:
	go mod tidy

sqlc:
	cd db && sqlc generate

migrate-up:
	migrate -path db/migrations -database "$(DATABASE_URL)" up

migrate-down:
	migrate -path db/migrations -database "$(DATABASE_URL)" down 1
