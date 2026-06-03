# AI Learnings

Evolving discoveries, session summaries, and non-obvious decisions made during development.
Add entries here when you find gotchas, make architectural choices that aren't obvious from the code, or complete a significant session.

---

## Entry format

```
### YYYY-MM-DD — <short title>
<What was learned or decided, and why.>
```

---

## Entries

### 2026-05-31 — Centralized HTTP Observability
Request completion logging belongs in `internal/platform/mw.LoggerStructured`, while handler/service error logging belongs in `internal/platform/httperr.Handle`. This keeps module handlers free of logging side effects, preserves the handler → service → repository boundary, and gives production debugging a single `request_id` that correlates the request log, error log, and panic log.

Internal errors should not be echoed to API clients. Return the generic `INTERNAL_ERROR` envelope and keep concrete error details in structured logs.

### 2026-05-31 — Idempotent Radio Browser Upserts
`ON CONFLICT DO UPDATE` still writes a new row version in PostgreSQL even when every assigned value is unchanged. The Radio Browser syncer now uses `IS DISTINCT FROM` in the conflict update `WHERE` clause, so unchanged station payloads do not update `synced_at`/`updated_at`, generate dead tuples, or inflate the sync result's changed-row count.

Repeated manual triggers can still create multiple full crawls with different request IDs. The syncer now rejects new triggers while any job is `pending` or `running`, so trigger spam cannot enqueue overlapping Radio Browser crawls.

Radio Browser can expose multiple UUIDs for the same stream URL. The sync insert now skips a new UUID when that URL is already present locally, keeping the local station set keyed by stream identity instead of only by external UUID.

### 2026-06-02 — Syncer Event Boundary
The syncer is event-driven, but RabbitMQ remains infrastructure rather than a business boundary. `internal/module/syncer` owns `SyncCommand`, sync job statuses, overlap rejection, and processing decisions; `internal/platform/mq` only declares topology, publishes JSON payloads, and consumes deliveries.

The API process is a producer only: it creates a `pending` sync job and publishes `sync.requested`. The standalone `cmd/syncer` process is the consumer: it marks jobs `running`, executes Radio Browser ingestion, and marks jobs `completed` or `failed`. This keeps long-running crawls out of the HTTP request lifecycle while preserving the handler -> service -> repository boundary.
