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
  ├──► internal/platform/health
  └──► internal/router

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
