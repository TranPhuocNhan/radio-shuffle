# Development Workflows

## Adding a New Module

1. Create `internal/module/{name}/` with all standard files:

   | File | Responsibility |
   |---|---|
   | `module.go` | Route registration: `NewModule()` constructor + `RegisterRoutes(*gin.RouterGroup)` |
   | `handler.go` | HTTP handlers — binds JSON, delegates to service, returns request/service errors |
   | `service.go` | Business logic — `Service` interface (exported) + `service` struct (unexported) |
   | `repository.go` | Port interfaces + domain structs (`XxxRow`, `CreateInput`, `UpdateInput`) |
   | `repository_sqlc.go` | SQLC-backed implementation of the `Repository` interface |
   | `dto.go` | Request/response structs with `json` and `binding` tags |
   | `error_mapper.go` | Module/domain error mapping for `httperr.Wrap` |
   | `handler_test.go` | Handler tests using stubs/fakes and `httptest` |
   | `service_test.go` | Service tests using fakes for repository/external deps |

2. Define the `Repository` interface and domain structs in `repository.go`
3. Write SQL queries in `db/query/{name}.sql` with SQLC annotations
4. Run `cd db && sqlc generate` to produce the Go accessors
5. Implement `repository_sqlc.go` wrapping `dbsqlc.Queries`
6. Define the `Service` interface and business logic in `service.go`
7. Create DTOs in `dto.go` with `json` + `binding` tags
8. Write handlers in `handler.go` as `httperr.AppHandler` functions that return errors
9. Add `error_mapper.go` and wrap routes with `httperr.Wrap(..., MapError)`
10. Register the module in `cmd/api/main.go`
11. Wire concrete repositories/services/handlers in `cmd/api/main.go`; keep `module.go` focused on route registration
12. Write tests for both handler and service layers

## Adding a New Migration

1. Create `db/migrations/{NNNNNN}_{description}.up.sql` and `.down.sql`
2. Number must be strictly sequential (check the last migration number)
3. Update `db/schema.sql` to reflect the new state
4. Run `make migrate-up` to apply
5. Run `cd db && sqlc generate` if the schema change affects queries

> Never modify an already-applied migration — create a new one instead.

## Syncer Event Workflow

1. Admin calls `POST /syncer/trigger` (JWT + admin role required)
2. API publishes `sync.requested` to RabbitMQ
3. Syncer worker consumes the event and runs `Service.Sync`
4. Sync status is stored in `sync_jobs`
5. Admin checks status via `GET /syncer/status/:request_id`

Retries use the TTL-based retry queue (`syncer.jobs.retry`). Failed messages after
`SYNC_MAX_RETRIES` are routed to the DLQ (`syncer.jobs.dlq`).

## API Documentation Workflow

The OpenAPI source of truth is `docs/openapi.yaml`.

When changing HTTP routes, handler DTOs, response DTOs, auth middleware placement,
query parameters, path parameters, or error mappings:

1. Update `docs/openapi.yaml`
2. Update `docs/openapi-report.md` if the endpoint inventory or known mismatches change
3. Keep reusable component schemas aligned with `internal/module/*/dto.go`
4. Re-run the full validation sequence before declaring the change complete

## Full Validation Sequence

MANDATORY before declaring any task complete:

```bash
go build ./...            # 1. compile check
go vet ./...              # 2. static analysis
go test -race ./...       # 3. run all tests
```

### Compilation

```bash
# Type-check the entire project
go build ./...

# Target a single module
go build ./internal/module/station/...
```

- NEVER skip compilation checks
- NEVER declare a task complete if `go build ./...` fails
- If SQLC-generated code is stale, regenerate before building: `cd db && sqlc generate`

### Test Execution

```bash
# Run all tests with race detection
make test
# or equivalently:
go test -race ./...

# Run full validation sequence
make verify

# Run tests for a single module
go test -race ./internal/module/auth/...

# Run a specific test by name
go test -race ./internal/module/station/... -run TestHandler_List_Empty

# Run tests with verbose output
go test -race -v ./internal/module/auth/...
```

- ALWAYS run `go build ./...` before running tests
- ALWAYS use `-race` to catch data races
- If a test needs a live database, it belongs in `test/integration/`, not alongside unit tests

### Integration Tests

```bash
# Start integration dependencies
make integration-up

# Run integration tests
export DATABASE_URL=postgres://root@localhost:5432/radio_shuffle?sslmode=disable
export RABBITMQ_URL=amqp://guest:guest@localhost:5672/
make test-integration

# Tear down dependencies
make integration-down
```

### SQLC Regeneration

When you modify any file under `db/query/` or `db/schema.sql`:

```bash
cd db && sqlc generate
```

Then verify the generated output compiles: `go build ./pkg/dbsqlc/...`

## Make Targets

| Command | Description |
|---|---|
| `make run` | Start the API server (`go run ./cmd/api`) |
| `make verify` | Run compile, vet, and race tests in required order |
| `make test` | Run all tests with race detection |
| `make test-unit` | Run race-tested unit/module tests only |
| `make test-integration` | Run race-tested integration tests in `test/integration` |
| `make integration-up` | Start Postgres + RabbitMQ via Docker Compose |
| `make integration-down` | Stop Docker Compose integration services |
| `make tidy` | Run `go mod tidy` |
| `make sqlc` | Regenerate SQLC code from `db/query/*.sql` |
| `make migrate-up` | Apply all pending database migrations |
| `make migrate-down` | Roll back the last migration |
