# Development Workflows

## Adding a New Module

1. Create `internal/module/{name}/` with all standard files:

   | File | Responsibility |
   |---|---|
   | `module.go` | DI wiring: `NewModule()` constructor + `RegisterRoutes(*gin.RouterGroup)` |
   | `handler.go` | HTTP handlers — binds JSON, delegates to service, maps errors to response helpers |
   | `service.go` | Business logic — `Service` interface (exported) + `service` struct (unexported) |
   | `repository.go` | Port interfaces + domain structs (`XxxRow`, `CreateInput`, `UpdateInput`) |
   | `repository_sqlc.go` | SQLC-backed implementation of the `Repository` interface |
   | `dto.go` | Request/response structs with `json` and `binding` tags |
   | `handler_test.go` | Handler tests using stubs/fakes and `httptest` |
   | `service_test.go` | Service tests using fakes for repository/external deps |

2. Define the `Repository` interface and domain structs in `repository.go`
3. Write SQL queries in `db/query/{name}.sql` with SQLC annotations
4. Run `cd db && sqlc generate` to produce the Go accessors
5. Implement `repository_sqlc.go` wrapping `dbsqlc.Queries`
6. Define the `Service` interface and business logic in `service.go`
7. Create DTOs in `dto.go` with `json` + `binding` tags
8. Write handlers in `handler.go` using `response.*` helpers
9. Wire everything in `module.go` with `NewModule()` + `RegisterRoutes()`
10. Register the module in `cmd/api/main.go`
11. Write tests for both handler and service layers

## Adding a New Migration

1. Create `db/migrations/{NNNNNN}_{description}.up.sql` and `.down.sql`
2. Number must be strictly sequential (check the last migration number)
3. Update `db/schema.sql` to reflect the new state
4. Run `make migrate-up` to apply
5. Run `cd db && sqlc generate` if the schema change affects queries

> Never modify an already-applied migration — create a new one instead.

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
| `make test` | Run all tests with race detection |
| `make tidy` | Run `go mod tidy` |
| `make sqlc` | Regenerate SQLC code from `db/query/*.sql` |
| `make migrate-up` | Apply all pending database migrations |
| `make migrate-down` | Roll back the last migration |
