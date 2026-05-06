-- name: InsertRefreshToken :one
INSERT INTO refresh_tokens (user_id, token_hash, expires_at)
VALUES ($1, $2, $3)
RETURNING *;

-- name: GetRefreshTokenByHash :one
SELECT * FROM refresh_tokens WHERE token_hash = $1;

-- name: DeleteRefreshTokenByHash :exec
DELETE FROM refresh_tokens WHERE token_hash = $1;

-- name: DeleteAllRefreshTokensForUser :exec
DELETE FROM refresh_tokens WHERE user_id = $1;

