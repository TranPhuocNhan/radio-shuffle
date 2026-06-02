# Module Responsibilities

## Module Status

| Module | Status | Notes |
|---|---|---|
| `auth` | Complete | JWT register/login/refresh/logout; token repository; cross-module adapters |
| `station` | Complete | Full CRUD — reference implementation for new modules |
| `user` | Repository-only | `Repository` interface + SQLC impl; no handler/service yet |
| `syncer` | Complete | Background ingestion from Radio Browser API; admin trigger/status endpoints |
| `radiobrowser` | Complete | Public read-only list of synced Radio Browser stations |
| `track` | Complete | Full CRUD scoped to station — `POST/PATCH/DELETE` require auth |
| `playlist` | Complete | CRUD + playlist tracks (add/remove/reorder), owner-only writes; auth required for reads |
| `stream` | Complete | Start/end streams + user history; auth required |

## Module File Anatomy

Every fully-implemented HTTP module contains:

| File | Responsibility |
|---|---|
| `module.go` | `NewModule()` constructor + `RegisterRoutes(*gin.RouterGroup)` |
| `handler.go` | HTTP handlers — binds JSON, delegates to service, returns request/service errors |
| `service.go` | `Service` interface (exported) + `service` struct (unexported) with business logic |
| `repository.go` | `Repository` interface (port) + all domain types (`XxxRow`, `CreateInput`, `UpdateInput`) |
| `repository_sqlc.go` | SQLC-backed implementation of `Repository` |
| `dto.go` | HTTP request/response structs with `json` and `binding` tags |
| `error_mapper.go` | Maps module/domain errors to response helpers for `httperr.Wrap` |
| `handler_test.go` | Handler unit tests using stub services and `httptest` |
| `service_test.go` | Service unit tests using fake repositories |

Non-HTTP modules (e.g., `syncer`) expose their service directly rather than routes:

```go
// module.go
func (m *Module) Service() Service { return m.svc }
```

## Module Isolation Rules

- Modules never import each other's packages at the service or handler level
- Cross-module dependencies are expressed as interfaces defined in the *consuming* module's `repository.go`
- An adapter in `{consumer}/adapters/` bridges the consuming interface to the providing module's repository
- `cmd/api/main.go` is the only place where module packages appear together

## Cross-Module Adapter Pattern

```
auth needs user data
  └── auth/repository.go defines: UserReader, UserWriter
  └── auth/adapters/auth_user.go implements those interfaces using user.Repository
  └── cmd/api/main.go wires it:
        userRepo  := user.NewRepository(pool)
        authUsers := authadapters.NewAuthUserAdapter(userRepo)
        authSvc   := auth.NewService(authUsers, authUsers, authTokens, authCfg, nil)
        auth.NewModule(auth.NewHandler(authSvc))
```

The adapter is the **only** file where two module packages appear in the same import block.

## Per-Module Detail

### `auth`

Handles registration, login, token refresh, and logout.

- Issues short-lived JWT access tokens (HS256, default 15 min)
- Issues long-lived refresh tokens stored as SHA-256 hashes in `refresh_tokens` table
- Depends on `UserReader` and `UserWriter` interfaces (implemented by `auth/adapters/auth_user.go` using `user.Repository`)
- Sentinel errors: `ErrEmailTaken`, `ErrInvalidCredentials`, `ErrInvalidRefreshToken`, `ErrWeakPassword`
- Handlers return errors and routes use `httperr.Wrap(..., MapError)`
- Routes: `POST /auth/register`, `/auth/login`, `/auth/refresh`, `/auth/logout`

### `station`

Full CRUD for radio stations. Use as the reference when implementing other modules.

- Sentinel errors: `ErrNotFound`
- Repository maps DB no-row errors to `ErrRepoNotFound`; service maps that to `ErrNotFound`
- Handlers return errors and routes use `httperr.Wrap(..., MapError)`
- List endpoints use `limit`/`offset` pagination (default 20, max 100)
- PATCH uses read-then-merge strategy in the service layer
- Follow endpoints require auth and are wired through the station module auth middleware
- Routes: `POST /stations`, `GET /stations`, `GET /stations/:station_id`, `PATCH /stations/:station_id`, `DELETE /stations/:station_id`, `POST/DELETE /stations/:station_id/follow`, `GET /stations/:station_id/following`, `GET /stations/:station_id/followers/count`, `GET /stations/followed`

### `user`

Provides `user.Repository` interface and its SQLC implementation. No HTTP surface.
Consumed by `auth` via `auth/adapters/auth_user.go`.

### `syncer`

Owns the Radio Browser ingestion use case and sync job state machine.

- Splits into two runtime surfaces:
  - API producer path: `Handler.Trigger` -> `JobService.Trigger` -> `CreateSyncJob(pending)` -> publish `sync.requested`
  - Worker consumer path: `cmd/syncer` -> `JobProcessor.Process` -> `Service.Sync` -> Radio Browser fetch/upsert
- Consumes `sync.requested` events from RabbitMQ through the standalone `cmd/syncer` process
- Allows only one pending/running sync job at a time; overlapping trigger requests return conflict
- Tracks job statuses as `pending`, `running`, `completed`, and `failed`
- Uses `SyncCommand` as the domain command and maps it to/from `mq.SyncRequestedMessage` at the MQ boundary
- Uses manual RabbitMQ ack/nack in the worker; failed jobs are retried through `syncer.jobs.retry` and exhausted/invalid messages go to `syncer.jobs.dlq`
- Uses `UpsertBatch` with `ON CONFLICT DO UPDATE` in `radio_browser_stations`, but skips conflict updates when the incoming row is identical and skips new UUIDs that reuse an existing stream URL
- `SyncResult.Upserted` counts rows inserted or changed, not every fetched API record
- Exposes `Service.Sync(ctx) (SyncResult, error)` for the worker ingestion use case
- Exposes `JobService.Trigger` and `JobService.Status` for admin HTTP endpoints
- API handlers return errors and routes use `httperr.Wrap(..., MapError)`
- Sync job repository no-row errors map to `ErrRepoJobNotFound`; job service maps that to `ErrJobNotFound`
- Admin endpoints:
  - `POST /syncer/trigger`
  - `GET /syncer/status/:request_id`

### `radiobrowser`

Read-only access to the synced Radio Browser stations.

- Routes: `GET /radio-browser/stations`
- List endpoints use `limit`/`offset` pagination (default 20, max 100)
- Handlers return errors and routes use `httperr.Wrap(..., MapError)`; unexpected repository/service errors fall through to platform internal error handling

### `track`

Full CRUD for tracks belonging to a station.

- Tracks are always scoped to a `station_id` (FK `ON DELETE CASCADE`)
- Routes are nested under `/stations/:station_id/tracks`
- `GET` endpoints are public; `POST`, `PATCH`, `DELETE` require a valid JWT
- Sentinel errors: `ErrNotFound`
- Repository maps DB no-row errors to `ErrRepoNotFound`; service maps that to `ErrNotFound`
- Invalid station references on track create map to `ErrStationNotFound`
- Handlers return errors and routes use `httperr.Wrap(..., MapError)`
- List endpoints use `limit`/`offset` pagination (default 20, max 100)
- `PATCH` uses read-then-merge strategy in the service layer
- Routes:
  - `GET  /stations/:station_id/tracks`
  - `GET  /stations/:station_id/tracks/:id`
  - `POST /stations/:station_id/tracks` *(auth required)*
  - `PATCH /stations/:station_id/tracks/:id` *(auth required)*
  - `DELETE /stations/:station_id/tracks/:id` *(auth required)*

### `playlist`

Owner-managed playlists with tracks and reorder support.

- All reads require auth; non-owners can read only public playlists
- Writes are owner-only (create/update/delete, add/remove/reorder tracks)
- List endpoints use `limit`/`offset` pagination (default 20, max 100)
- Playlist track positions are required, non-negative, and unique within a playlist; reorder requests must include the exact existing track set
- Routes:
  - `POST /playlists`
  - `GET  /playlists`
  - `GET  /playlists/:id`
  - `PATCH /playlists/:id`
  - `DELETE /playlists/:id`
  - `POST /playlists/:id/tracks`
  - `GET  /playlists/:id/tracks`
  - `DELETE /playlists/:id/tracks/:track_id`
  - `PATCH /playlists/:id/tracks/reorder`

### `stream`

Tracks user listening sessions.

- All routes require auth; users only see their own streams
- Repository maps DB no-row errors to `ErrRepoNotFound`; service maps that to `ErrNotFound`
- Invalid station references on stream start map to `ErrStationNotFound`
- Handlers return errors and routes use `httperr.Wrap(..., MapError)`
- List endpoints use `limit`/`offset` pagination (default 20, max 100)
- Endpoints:
  - `POST /streams`
  - `GET  /streams`
  - `GET  /streams/:id`
  - `PATCH /streams/:id/end`
