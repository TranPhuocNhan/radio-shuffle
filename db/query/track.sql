-- name: CreateTrack :one
INSERT INTO tracks (
    station_id, title, artist, audio_url, duration_seconds
) VALUES (
    $1, $2, $3, $4, $5
)
RETURNING *;

-- name: GetTrackByID :one
SELECT * FROM tracks WHERE id = $1;

-- name: ListTracksByStationID :many
SELECT *
FROM tracks
WHERE station_id = $1
ORDER BY created_at DESC
LIMIT $2 OFFSET $3;

-- name: CountTracksByStationID :one
SELECT COUNT(*)::bigint FROM tracks WHERE station_id = $1;

-- name: UpdateTrack :one
UPDATE tracks SET
    title = $2,
    artist = $3,
    audio_url = $4,
    duration_seconds = $5,
    updated_at = now()
WHERE id = $1 AND station_id = $6
RETURNING *;

-- name: DeleteTrack :exec
DELETE FROM tracks WHERE id = $1;
