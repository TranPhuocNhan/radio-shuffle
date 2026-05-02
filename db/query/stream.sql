-- name: CreateStream :one
INSERT INTO streams (user_id, station_id, started_at, ended_at)
VALUES ($1, $2, $3, $4)
RETURNING *;

-- name: EndStreamByID :one
UPDATE streams
SET ended_at = $2
WHERE id = $1
RETURNING *;

-- name: GetStreamByID :one
SELECT * FROM streams WHERE id = $1;

-- name: ListStreamsByUserID :many
SELECT *
FROM streams
WHERE user_id = $1
ORDER BY started_at DESC
LIMIT $2 OFFSET $3;

-- name: CountStreamsByUserID :one
SELECT COUNT(*)::bigint FROM streams WHERE user_id = $1;

