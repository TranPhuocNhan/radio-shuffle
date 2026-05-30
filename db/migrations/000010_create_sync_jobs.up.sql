CREATE TABLE sync_jobs (
    request_id TEXT PRIMARY KEY,
    status TEXT NOT NULL,
    requested_by TEXT NOT NULL,
    scope JSONB NOT NULL,
    error_message TEXT,
    started_at TIMESTAMPTZ,
    finished_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX idx_sync_jobs_status ON sync_jobs(status);

