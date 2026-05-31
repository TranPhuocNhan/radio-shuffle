# radio-shuffle (backend skeleton)

Radio Shuffle is a Go backend skeleton for a radio/playlist service. It ships a Gin HTTP API, a modular monolith layout, and SQLC-ready data access for PostgreSQL (with SQLite used in tests). Migrations, shared middleware, and a consistent response envelope are included so you can add endpoints and features incrementally.

## Project overview

- Modular packages under `internal/module/` with handler → service → repository layering
- Domain modules for **user**, **station**, **track**, **playlist**, **stream**, plus a **syncer** for Radio Browser ingestion
- SQLC workflow with authoritative SQL in `db/query/` and generated code in `pkg/dbsqlc/`
- Migrations via `db/migrations/` and a full schema snapshot in `db/schema.sql`
- Shared platform helpers for config, database pool, auth middleware, and response envelopes

## Layout (high level)

```
cmd/api/               HTTP entrypoint, graceful shutdown
internal/platform/    config, database pool, health, middleware, response envelope
internal/router/      Gin route registration
pkg/dbsqlc/           SQL accessors (maintain parity with db/query/*.sql)
db/schema.sql          Full schema snapshot for tooling
db/migrations/         golang-migrate numbered up/down migrations
db/query/              Authoritative *.sql snippets for regeneration
logs/                   runtime logs (.gitkeep)
```

Mandatory domain modules (**user**, **station**, **track**, **playlist**, **stream**) are modeled in migrations + SQL; HTTP handlers/services can be added incrementally on top.

## Env

See `.env.example`. Required variables for boot: **DATABASE_URL**, **RABBITMQ_URL**, **JWT_SIGNING_KEY**.

## Run

```
cp .env.example .env
# start Postgres and RabbitMQ locally; then:
make integration-up
make migrate-up
go mod tidy
make run
```

Smoke: `curl -s localhost:8080/api/v1/ping`.

Swagger UI: `http://localhost:8080/swagger`

## Responses

Success envelope mirrors the project prompt:
`success`, `data`, `error`, optional `meta` with pagination `{page,limit,total}` — see `internal/platform/response/respond.go`.

## Commands

See `Makefile` (`run`, `test`, `tidy`, `sqlc`, `migrate-up`, `migrate-down`).
