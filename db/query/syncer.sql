-- name: CreateSyncJob :exec
INSERT INTO sync_jobs (request_id, status, requested_by, scope, created_at)
VALUES ($1, $2, $3, $4, now())
ON CONFLICT DO NOTHING;

-- name: GetSyncJobByID :one
SELECT request_id, status, requested_by, scope, error_message, started_at, finished_at, created_at
FROM sync_jobs
WHERE request_id = $1;

-- name: GetActiveSyncJob :one
SELECT request_id, status, requested_by, scope, error_message, started_at, finished_at, created_at
FROM sync_jobs
WHERE status IN ('pending', 'running')
ORDER BY created_at DESC
LIMIT 1;

-- name: MarkSyncJobRunning :exec
UPDATE sync_jobs
SET status = 'running', started_at = now(), error_message = NULL
WHERE request_id = $1;

-- name: MarkSyncJobCompleted :exec
UPDATE sync_jobs
SET status = 'completed', finished_at = now()
WHERE request_id = $1;

-- name: MarkSyncJobFailed :exec
UPDATE sync_jobs
SET status = 'failed', finished_at = now(), error_message = $2
WHERE request_id = $1;
