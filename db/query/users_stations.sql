-- name: FollowStation :exec
INSERT INTO users_stations (user_id, station_id)
VALUES ($1, $2)
ON CONFLICT DO NOTHING;

-- name: UnfollowStation :exec
DELETE FROM users_stations
WHERE user_id = $1 AND station_id = $2;

-- name: IsFollowingStation :one
SELECT EXISTS(
    SELECT 1
    FROM users_stations
    WHERE user_id = $1 AND station_id = $2
);

-- name: GetStationsByUserID :many
SELECT s.*
FROM stations s
JOIN users_stations us ON s.id = us.station_id
WHERE us.user_id = $1
ORDER BY s.created_at DESC
LIMIT $2 OFFSET $3;

-- name: GetFollowersByStationID :many
SELECT u.*
FROM users u
JOIN users_stations us ON u.id = us.user_id
WHERE us.station_id = $1
ORDER BY u.created_at DESC;

-- name: CountStationsByUserID :one
SELECT COUNT(*)
FROM users_stations
WHERE user_id = $1;

-- name: CountFollowersByStationID :one
SELECT COUNT(*)
FROM users_stations
WHERE station_id = $1;

