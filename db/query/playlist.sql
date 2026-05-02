-- name: CreatePlaylist :one
INSERT INTO playlists (name, description, is_public, owner_id)
VALUES ($1, $2, $3, $4)
RETURNING *;

-- name: GetPlaylistByID :one
SELECT * FROM playlists WHERE id = $1;

-- name: ListPlaylistsOwnedBy :many
SELECT *
FROM playlists
WHERE owner_id = $1
ORDER BY created_at DESC
LIMIT $2 OFFSET $3;

-- name: ListPublicPlaylists :many
SELECT *
FROM playlists
WHERE is_public = TRUE
ORDER BY created_at DESC
LIMIT $1 OFFSET $2;

-- name: CountPlaylistsOwnedBy :one
SELECT COUNT(*)::bigint FROM playlists WHERE owner_id = $1;

-- name: UpdatePlaylist :one
UPDATE playlists SET
    name = $2,
    description = $3,
    is_public = $4,
    updated_at = now()
WHERE id = $1
RETURNING *;

-- name: DeletePlaylist :exec
DELETE FROM playlists WHERE id = $1;

-- name: AddTrackToPlaylist :exec
INSERT INTO playlist_tracks (playlist_id, track_id, position)
VALUES ($1, $2, $3)
ON CONFLICT DO NOTHING;

-- name: ListPlaylistTracks :many
SELECT t.*
FROM playlist_tracks pt
JOIN tracks t ON t.id = pt.track_id
WHERE pt.playlist_id = $1
ORDER BY pt.position ASC, pt.track_id ASC;
