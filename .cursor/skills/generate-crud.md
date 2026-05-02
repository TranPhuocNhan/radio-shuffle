---
name: generate-crud
description: Scaffolds a complete CRUD feature for a given model — SQL migration, SQLC queries, and module files (handler, service, repository, DTOs, wiring). Use when the user says "generate crud for [ModelName]" or "scaffold [ModelName]".
---

# Skill: Generate CRUD

## Trigger

User says: `generate crud for [ModelName]` or `scaffold [ModelName]`

## Steps

### 1. Confirm the Model

Before generating, confirm:
- The model name (e.g. `Station`, `Track`, `Playlist`, `Stream`, `User`)
- Its fields (ask if not provided; infer from context if obvious)
- Whether it requires auth on write endpoints

### 2. Create SQL Files

#### `db/migration/{next_number}_create_{table}.up.sql`

```sql
CREATE TABLE stations (
    id         BIGSERIAL PRIMARY KEY,
    name       TEXT NOT NULL,
    genre      TEXT NOT NULL DEFAULT '',
    stream_url TEXT NOT NULL,
    owner_id   BIGINT NOT NULL REFERENCES users(id),
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_stations_owner_id ON stations(owner_id);
```

#### `db/migration/{next_number}_create_{table}.down.sql`

```sql
DROP TABLE IF EXISTS stations;
```

#### Update `db/schema.sql`

Append the new table DDL so SQLC can resolve types.

#### `db/query/{model_lower}.sql`

```sql
-- name: GetStation :one
SELECT * FROM stations WHERE id = $1;

-- name: ListStations :many
SELECT * FROM stations ORDER BY created_at DESC LIMIT $1 OFFSET $2;

-- name: CreateStation :one
INSERT INTO stations (name, genre, stream_url, owner_id)
VALUES ($1, $2, $3, $4)
RETURNING *;

-- name: UpdateStation :one
UPDATE stations
SET name = $2, genre = $3, stream_url = $4, updated_at = now()
WHERE id = $1
RETURNING *;

-- name: DeleteStation :exec
DELETE FROM stations WHERE id = $1;
```

### 3. Run SQLC Generate

```bash
sqlc generate
```

This produces type-safe Go code in `pkg/dbsqlc/`.

### 4. Create Module Files

All files go under `internal/module/{model_lower}/`:

#### `dto.go`
- `Create{Model}Request` — fields with `json` and `binding:"required"` tags
- `Update{Model}Request` — pointer fields for partial updates
- `{Model}Response` — fields safe to return (no passwords, no internal fields)
- `New{Model}Response(dbsqlc.{Model}) {Model}Response` constructor function

#### `repository.go`
- Define the `Repository` interface with methods matching the SQLC queries

#### `repository_sqlc.go`
- Implement the `Repository` interface wrapping `*dbsqlc.Queries`
- `NewRepository(db dbsqlc.DBTX) Repository` constructor

#### `service.go`
- Define the `Service` interface
- Implement it with an unexported struct holding the `Repository` interface
- Services call the repository — no direct SQLC usage
- Define sentinel errors: `var ErrNotFound = errors.New("{model} not found")`

#### `handler.go`
- `GET /{model_plural}` — list with `offset` and `limit` query params
- `GET /{model_plural}/:id` — get by ID, 404 if not found
- `POST /{model_plural}` — create, returns 201
- `PATCH /{model_plural}/:id` — partial update, 404 if not found
- `DELETE /{model_plural}/:id` — delete, returns 204
- All handlers use `ShouldBindJSON` and response helpers from `internal/platform/response/`
- Map sentinel errors (`ErrNotFound`, `ErrConflict`) to the correct HTTP status via response helpers

#### `module.go`
- `NewModule(db dbsqlc.DBTX) *Module` — wires repo → service → handler
- `RegisterRoutes(rg *gin.RouterGroup)` — registers all routes

### 5. Register in main.go

In `cmd/server/main.go`, add:
```go
stationMod := station.NewModule(db)
stationMod.RegisterRoutes(api)
```

### 6. Confirm Output

After generating, list the created/modified files and ask:
> "CRUD for `{Model}` is ready. Want me to write tests or review the code?"

## Radio Domain Field Reference

| Model    | Key Fields                                                        |
|----------|-------------------------------------------------------------------|
| Station  | name, genre, stream_url, description, cover_image_url, owner_id  |
| Track    | title, artist, audio_url, duration_seconds, station_id           |
| Playlist | name, description, is_public, owner_id                           |
| Stream   | user_id, station_id, started_at, ended_at                        |
