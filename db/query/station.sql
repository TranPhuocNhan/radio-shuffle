-- name: CreateStation :one
INSERT INTO stations (
    name, genre, description, stream_url, cover_image_url, is_public, owner_id
) VALUES (
    $1, $2, $3, $4, $5, $6, $7
)
RETURNING *;

-- name: GetStationByID :one
SELECT * FROM stations WHERE id = $1;

-- name: ListStations :many
SELECT *
FROM stations
ORDER BY created_at DESC
LIMIT $1 OFFSET $2;

-- name: CountStations :one
SELECT COUNT(*)::bigint FROM stations;

-- name: UpdateStation :one
UPDATE stations
SET
    name = $2,
    genre = $3,
    description = $4,
    stream_url = $5,
    cover_image_url = $6,
    is_public = $7,
    owner_id = $8,
    updated_at = now()
WHERE id = $1
RETURNING *;

-- name: DeleteStation :exec
DELETE FROM stations WHERE id = $1;
