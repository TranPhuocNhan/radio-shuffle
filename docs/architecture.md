# Architecture

## Root Folders

| Path | Purpose |
|---|---|
| `cmd/api/` | HTTP API server entry point |
| `cmd/syncer/` | Background Radio Browser syncer entry point |
| `internal/module/` | Domain modules (business features) |
| `internal/platform/` | Shared infrastructure (no business logic) |
| `internal/router/` | `RouteRegistrar` interface + `Server` wrapper |
| `pkg/dbsqlc/` | SQLC-generated DB accessors — **never edit by hand** |
| `db/` | Schema, migrations, SQLC queries, config |
| `test/integration/` | Integration tests that exercise database-backed workflows |
| `docs/` | Living documentation |
| `logs/` | Runtime log files (`.gitkeep`) |

## Internal Structure

```
internal/
├── module/
│   ├── auth/                   ← Authentication & JWT token management
│   │   ├── adapters/           ← Cross-module adapter (auth ↔ user)
│   │   ├── handler.go
│   │   ├── service.go
│   │   ├── repository.go       ← Port interfaces + domain types
│   │   ├── repository_sqlc.go  ← SQLC-backed implementation
│   │   ├── dto.go
│   │   └── module.go           ← DI wiring, RegisterRoutes()
│   ├── station/                ← Full CRUD (reference implementation)
│   ├── user/                   ← Repository-only; consumed by auth via adapter
│   ├── syncer/                 ← Background ingestion from Radio Browser API
│   ├── radiobrowser/            ← Read-only access to synced Radio Browser stations
│   ├── track/                  ← Station-scoped track CRUD
│   ├── playlist/               ← Owner-managed playlists and playlist tracks
│   └── stream/                 ← User listening sessions
├── platform/
│   ├── apidocs/                ← gin-swagger UI wiring + embedded OpenAPI spec endpoint
│   ├── config/                 ← Env/config loading via viper
│   ├── database/               ← pgxpool connection factory
│   ├── health/                 ← /health and /ready endpoints
│   ├── mw/                     ← Middleware: RequestID, Recovery, Logger, AuthRequired
│   ├── mq/                     ← RabbitMQ client + topology setup + message payloads
│   ├── radiobrowser/           ← External Radio Browser API client
│   └── response/               ← Standardized JSON envelope helpers
└── router/                     ← RouteRegistrar interface, Server wrapper
```

## Data Layer

```
db/
├── schema.sql                  ← Full schema snapshot (authoritative for SQLC)
├── sqlc.yaml                   ← SQLC v2 config (postgresql + pgx/v5)
├── query/                      ← One .sql file per domain concern
│   ├── auth.sql                ← Refresh token queries
│   ├── user.sql
│   ├── station.sql
│   ├── track.sql
│   ├── playlist.sql
│   ├── stream.sql
│   └── radio_browser_station.sql
└── migrations/                 ← Numbered golang-migrate up/down files
    ├── 000001_create_users.{up,down}.sql
    ├── ...
    └── 000008_create_radio_browser_stations.{up,down}.sql
```

## Dependency Graph

```
cmd/api ──► internal/module/* ──► pkg/dbsqlc (generated)
  │               │
  │               ├──► internal/platform/response
  │               └──► internal/platform/mw
  │
  ├──► internal/platform/config
  ├──► internal/platform/database
  ├──► internal/platform/apidocs ──► docs/openapi.yaml (embedded)
  ├──► internal/platform/health
  └──► internal/router

cmd/api ──► internal/module/syncer ──► internal/platform/mq ──► RabbitMQ

cmd/syncer ──► internal/module/syncer ──► internal/platform/radiobrowser
  │                │                            └──► pkg/dbsqlc
  │                └──► internal/platform/mq ──► RabbitMQ
  ├──► internal/platform/config
  └──► internal/platform/database
```

**Boundary rules:**

- `internal/module/*` packages never import each other directly
- `internal/platform/*` packages never contain business logic
- `pkg/dbsqlc/` is accessed only through repository implementations — never from handlers or services
- All cross-module wiring is done exclusively in `cmd/api/main.go`

## Syncer Event-Driven Architecture

The Radio Browser syncer is intentionally asynchronous. The HTTP API only accepts
admin intent, persists a sync job, and publishes an event. The background syncer
process consumes that event and performs the long-running ingestion work.

```
Admin HTTP request
  → cmd/api
  → syncer.Handler.Trigger
  → syncer.JobService.Trigger
  → syncer.Repository.CreateSyncJob(status=pending)
  → syncer.Publisher.PublishSyncCommand
  → platform/mq.Publish(exchange=sync.events, routing_key=sync.requested)
  → RabbitMQ queue syncer.jobs

RabbitMQ delivery
  → cmd/syncer
  → platform/mq.Consume(queue=syncer.jobs, prefetch=1)
  → syncer.JobProcessor.Process
  → syncer.Repository.MarkSyncJobRunning
  → syncer.Service.Sync
  → platform/radiobrowser.Client.FetchStations
  → syncer.Repository.UpsertBatch
  → syncer.Repository.MarkSyncJobCompleted or MarkSyncJobFailed
  → message Ack, retry publish, or DLQ publish
```

### Event Topology

| Component | Default | Purpose |
|---|---|---|
| Exchange | `sync.events` | Durable direct exchange for sync events |
| Main routing key | `sync.requested` | Published by the API and consumed by the worker |
| Main queue | `syncer.jobs` | Work queue for sync requests |
| Retry routing key | `sync.requested.retry` | Worker publishes failed attempts here |
| Retry queue | `syncer.jobs.retry` | TTL queue; expired messages dead-letter back to `sync.requested` |
| DLQ routing key | `sync.requested.dlq` | Terminal failure route |
| DLQ | `syncer.jobs.dlq` | Holds invalid or exhausted sync messages |

`sync.requested` payloads are `internal/platform/mq.SyncRequestedMessage`:

```json
{
  "request_id": "hex-encoded-request-id",
  "requested_by": "admin-user-id",
  "scope": {"mode": "full"},
  "requested_at": "2026-06-02T00:00:00Z"
}
```

The API adds `x-request-id` and `x-retry-count` headers when publishing. The
worker increments `x-retry-count` when scheduling a retry.

### Delivery Semantics

- `cmd/api` creates the `sync_jobs` row before publishing the event.
- If publish fails, the API marks that job `failed` and returns the publish error.
- `cmd/syncer` uses manual acknowledgement; successful, duplicate, and in-progress messages are acknowledged.
- Business failures mark the job `failed`, then the message is republished to the retry queue until `SYNC_MAX_RETRIES` is reached.
- Retry delay is implemented by the retry queue TTL (`SYNC_RETRY_TTL_MS`) and dead-letter routing back to the main queue.
- Invalid JSON messages bypass business processing and are sent directly to the DLQ.
- `PrefetchCount=1` keeps one in-flight sync per worker process; the repository also rejects overlapping pending/running jobs.

The event bus is infrastructure. Business rules still live in `internal/module/syncer`;
RabbitMQ details stay in `internal/platform/mq` and the syncer publisher/adapter.

### MQ Package Anatomy

`internal/platform/mq` is organized by RabbitMQ responsibility:

| File | Responsibility |
|---|---|
| `config.go` | RabbitMQ URL, exchange, queue, routing key, retry, DLQ, and prefetch settings |
| `client.go` | `Client` lifecycle: dial, open channel, declare topology, set QoS, close |
| `topology.go` | Durable direct exchange, main queue, retry queue, DLQ, and bindings |
| `publisher.go` | JSON publish helper using the configured exchange and caller-provided routing key |
| `consumer.go` | Main queue consumer setup with manual acknowledgements |
| `message.go` | Infrastructure-level event payload structs such as `SyncRequestedMessage` |

Module-owned publishers/adapters, such as `internal/module/syncer/publisher.go`
and `internal/module/syncer/mq_adapter.go`, translate domain commands to these
infrastructure payloads. Do not put syncer business rules in `internal/platform/mq`.

### Syncer Binary Anatomy

`cmd/syncer` is organized by process responsibility:

| File | Responsibility |
|---|---|
| `main.go` | Process entrypoint: logger setup, config load, signal context, app lifecycle |
| `app.go` | Worker app and RabbitMQ consume loop |
| `wiring.go` | Database, RabbitMQ, Radio Browser client, repository, service, and processor wiring |
| `message_handler.go` | Decode `sync.requested`, invoke `JobProcessor`, decide ack/retry/DLQ outcomes, and keep delivery helper logic near the message flow |
| `logging.go` | Syncer binary logger setup |

## Request Flow

```
HTTP request
  → Gin router
  → platform/mw  (RequestID → Recovery → Logger → [AuthRequired])
  → module/handler       bind JSON, parse params
  → module/service       business logic, sentinel errors
  → module/repository    interface call
  → repository_sqlc      dbsqlc.Queries → PostgreSQL

  ← repository_sqlc      dbsqlc row → domain type
  ← service              repository errors → domain sentinel errors
  ← httperr.Handle       module MapError → response.NotFound / response.Conflict / …
  ← response.*           JSON envelope
```

Request completion logs are emitted by `mw.LoggerStructured()` with `request_id`, method, path, route, query, status, latency, client IP, user agent, optional `user_id`, and response size. Handler/service errors are logged only in `httperr.Handle`, so module handlers do not need their own request logs. Panic recovery logs the same request context plus a stack trace.

## Playlist Track Ordering

Playlist track positions are unique and non-negative within each playlist.

- `playlist_tracks.position` is required by the API and `position >= 0` is enforced by the database.
- `(playlist_id, position)` is unique, so a playlist cannot contain two tracks at the same position.
- Reorder requests must include the exact existing track set for that playlist.
- Playlist reorder validation and updates run inside one repository transaction with row locks to avoid validating stale track membership.

## Response Envelope

All responses use `internal/platform/response` helpers — never construct manually.

```json
// Success
{"success": true,  "data": {...},  "error": null}

// Success with pagination
{"success": true,  "data": [...],  "error": null, "meta": {"page":1,"limit":20,"total":42}}

// Error
{"success": false, "data": null,   "error": {"code": "NOT_FOUND", "message": "..."}}
```

Error codes: `VALIDATION_ERROR` · `UNAUTHORIZED` · `FORBIDDEN` · `NOT_FOUND` · `CONFLICT` · `INTERNAL_ERROR`

`INTERNAL_ERROR` responses use a generic `"internal error"` message for clients. The original error detail is available in structured server logs, correlated by `request_id`.

## Tech Stack

| Concern | Technology |
|---|---|
| Language | Go 1.23 |
| HTTP framework | Gin (`github.com/gin-gonic/gin`) |
| Database | PostgreSQL (pgx/v5, pgxpool) |
| SQL codegen | SQLC v2 (`db/sqlc.yaml`) |
| Auth | JWT (`golang-jwt/jwt/v5`), bcrypt |
| Config | Viper + `.env` |
| Migrations | golang-migrate (numbered up/down SQL) |
| Testing | `testing` + `testify` + `httptest` |
| Logging | `log/slog` (structured JSON) |
| CORS | `gin-contrib/cors` |
| External API | Radio Browser (`internal/platform/radiobrowser/`) |
| Messaging | RabbitMQ (`internal/platform/mq/`, `amqp091-go`) |
