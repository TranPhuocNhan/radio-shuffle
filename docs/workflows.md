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

The syncer is event-driven so the HTTP API does not block on a full Radio
Browser crawl. The API is the event producer; `cmd/syncer` is the event consumer.

### Trigger Path

1. Admin calls `POST /syncer/trigger` (JWT + admin role required)
2. `syncer.Handler.Trigger` extracts the admin user ID from auth context
3. `syncer.JobService.Trigger` checks `sync_jobs` for an active `pending` or `running` job
4. If no active job exists, the job service creates a new `pending` job with scope `{"mode":"full"}`
5. `syncer.Publisher` maps the domain `SyncCommand` to `mq.SyncRequestedMessage`
6. The API publishes `sync.requested` to the `sync.events` RabbitMQ exchange
7. The API returns `202 Accepted` with the `request_id` and status URL

If publishing fails after the row is created, the job is marked `failed` and the
trigger request returns the publish error.

### Worker Path

1. `cmd/syncer` consumes from `syncer.jobs` with manual acknowledgements and `PrefetchCount=1`
2. The worker decodes `mq.SyncRequestedMessage` and maps it back to `syncer.SyncCommand`
3. `syncer.JobProcessor.Process` loads or creates the matching `sync_jobs` row
4. Completed jobs are treated as duplicates and acknowledged without re-running
5. Running jobs are treated as in-progress and acknowledged without re-running
6. Pending or failed jobs are marked `running`
7. `syncer.Service.Sync` fetches Radio Browser stations in batches and calls `Repository.UpsertBatch`
8. On success, the job is marked `completed` and the message is acknowledged
9. On business failure, the job is marked `failed` and the message is sent to retry or DLQ

### Status Model

| Status | Meaning |
|---|---|
| `pending` | API accepted the request and published, or the worker created the row from an event |
| `running` | Worker claimed the job and is fetching/upserting Radio Browser stations |
| `completed` | Worker finished all batches successfully |
| `failed` | Publish failed or worker processing failed; `error_message` contains the failure |

Admin status checks use `GET /syncer/status/:request_id`, which reads the
`sync_jobs` row and maps missing jobs to `404 NOT_FOUND`.

### Retry And DLQ Behavior

Only one sync job can be pending or running at a time. Additional trigger
requests return `409 CONFLICT` until the active job completes or fails.

Radio Browser station writes are idempotent: duplicate `stationuuid` rows are only
updated when persisted fields differ from the incoming API payload. New UUIDs are
also skipped when the same stream URL already exists locally.

Retries use the TTL-based retry queue (`syncer.jobs.retry`). Failed messages after
`SYNC_MAX_RETRIES` are routed to the DLQ (`syncer.jobs.dlq`).

The retry queue is bound with routing key `sync.requested.retry`. It has
`x-message-ttl=SYNC_RETRY_TTL_MS` and dead-letters expired messages back to
`sync.events` with routing key `sync.requested`, so delayed retries re-enter the
main worker queue. The worker increments the `x-retry-count` header before
publishing to retry or DLQ.

Invalid JSON messages cannot be converted into a sync command, so they are sent
directly to the DLQ and then acknowledged.

### Event Configuration

| Environment variable | Default | Purpose |
|---|---|---|
| `SYNC_EVENTS_EXCHANGE` | `sync.events` | Direct exchange used for sync events |
| `SYNC_QUEUE` | `syncer.jobs` | Main worker queue |
| `SYNC_RETRY_QUEUE` | `syncer.jobs.retry` | TTL retry queue |
| `SYNC_DLQ` | `syncer.jobs.dlq` | Dead-letter queue |
| `SYNC_RETRY_TTL_MS` | `60000` | Delay before retry messages return to the main queue |
| `SYNC_MAX_RETRIES` | `5` | Maximum worker retries before DLQ |

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
