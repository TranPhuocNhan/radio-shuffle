# Module Responsibilities

## Module Status

| Module | Status | Notes |
|---|---|---|
| `auth` | Complete | JWT register/login/refresh/logout; token repository; cross-module adapters |
| `station` | Complete | Full CRUD — reference implementation for new modules |
| `user` | Repository-only | `Repository` interface + SQLC impl; no handler/service yet |
| `syncer` | Complete | Background ingestion from Radio Browser API; no HTTP routes |
| `radiobrowser` | Complete | Public read-only list of synced Radio Browser stations |
| `track` | Complete | Full CRUD scoped to station — `POST/PATCH/DELETE` require auth |
| `playlist` | Stub | `RegisterRoutes` scaffolded; CRUD not yet implemented |
| `stream` | Stub | `RegisterRoutes` scaffolded; CRUD not yet implemented |

## Module File Anatomy

Every fully-implemented HTTP module contains:

| File | Responsibility |
|---|---|
| `module.go` | `NewModule()` constructor + `RegisterRoutes(*gin.RouterGroup)` |
| `handler.go` | HTTP handlers — binds JSON, delegates to service, maps errors to response helpers |
| `service.go` | `Service` interface (exported) + `service` struct (unexported) with business logic |
| `repository.go` | `Repository` interface (port) + all domain types (`XxxRow`, `CreateInput`, `UpdateInput`) |
| `repository_sqlc.go` | SQLC-backed implementation of `Repository` |
| `dto.go` | HTTP request/response structs with `json` and `binding` tags |
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
        auth.NewModule(&cfg, authUsers, authUsers, authTokens, nil)
```

The adapter is the **only** file where two module packages appear in the same import block.

## Per-Module Detail

### `auth`

Handles registration, login, token refresh, and logout.

- Issues short-lived JWT access tokens (HS256, default 15 min)
- Issues long-lived refresh tokens stored as SHA-256 hashes in `refresh_tokens` table
- Depends on `UserReader` and `UserWriter` interfaces (implemented by `auth/adapters/auth_user.go` using `user.Repository`)
- Sentinel errors: `ErrEmailTaken`, `ErrInvalidCredentials`, `ErrInvalidRefreshToken`, `ErrWeakPassword`
- Routes: `POST /auth/register`, `/auth/login`, `/auth/refresh`, `/auth/logout`

### `station`

Full CRUD for radio stations. Use as the reference when implementing other modules.

- Sentinel errors: `ErrNotFound`
- List endpoints use `limit`/`offset` pagination (default 20, max 100)
- PATCH uses read-then-merge strategy in the service layer
- Routes: `POST /stations`, `GET /stations`, `GET /stations/:station_id`, `PATCH /stations/:station_id`, `DELETE /stations/:station_id`

### `user`

Provides `user.Repository` interface and its SQLC implementation. No HTTP surface.
Consumed by `auth` via `auth/adapters/auth_user.go`.

### `syncer`

Ingests Radio Browser stations in paginated batches via `radiobrowser.Client`.

- Runs on a configurable ticker interval (default 6h, set via `SYNC_INTERVAL`)
- Uses `UpsertBatch` with ON CONFLICT DO UPDATE in `radio_browser_stations`
- Exposes `Service.Sync(ctx) (SyncResult, error)` — called directly from `cmd/syncer/main.go`
- No HTTP routes

### `radiobrowser`

Read-only access to the synced Radio Browser stations.

- Routes: `GET /radio-browser/stations`
- List endpoints use `limit`/`offset` pagination (default 20, max 100)

### `track`

Full CRUD for tracks belonging to a station.

- Tracks are always scoped to a `station_id` (FK `ON DELETE CASCADE`)
- Routes are nested under `/stations/:station_id/tracks`
- `GET` endpoints are public; `POST`, `PATCH`, `DELETE` require a valid JWT
- Sentinel errors: `ErrNotFound`
- List endpoints use `limit`/`offset` pagination (default 20, max 100)
- `PATCH` uses read-then-merge strategy in the service layer
- Routes:
  - `GET  /stations/:station_id/tracks`
  - `GET  /stations/:station_id/tracks/:id`
  - `POST /stations/:station_id/tracks` *(auth required)*
  - `PATCH /stations/:station_id/tracks/:id` *(auth required)*
  - `DELETE /stations/:station_id/tracks/:id` *(auth required)*

### `playlist`, `stream`

Currently stubs — `RegisterRoutes` is a no-op. Schema, migrations, and SQLC queries exist.
Implement following `station` and `track` as reference implementations.
