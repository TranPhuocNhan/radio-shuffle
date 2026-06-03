-- name: UpsertRadioBrowserStation :execrows
INSERT INTO radio_browser_stations (
    stationuuid, name, url, url_resolved, homepage, favicon,
    country, countrycode, state, language, codec,
    bitrate, votes, tags, last_check_ok
) SELECT
    $1, $2, $3, $4, $5, $6,
    $7, $8, $9, $10, $11,
    $12, $13, $14, $15
WHERE NOT EXISTS (
    SELECT 1
    FROM radio_browser_stations
    WHERE url = $3
      AND stationuuid <> $1
)
ON CONFLICT (stationuuid) DO UPDATE SET
    name          = EXCLUDED.name,
    url           = EXCLUDED.url,
    url_resolved  = EXCLUDED.url_resolved,
    homepage      = EXCLUDED.homepage,
    favicon       = EXCLUDED.favicon,
    country       = EXCLUDED.country,
    countrycode   = EXCLUDED.countrycode,
    state         = EXCLUDED.state,
    language      = EXCLUDED.language,
    codec         = EXCLUDED.codec,
    bitrate       = EXCLUDED.bitrate,
    votes         = EXCLUDED.votes,
    tags          = EXCLUDED.tags,
    last_check_ok = EXCLUDED.last_check_ok,
    synced_at     = now(),
    updated_at    = now()
WHERE (
    radio_browser_stations.name,
    radio_browser_stations.url,
    radio_browser_stations.url_resolved,
    radio_browser_stations.homepage,
    radio_browser_stations.favicon,
    radio_browser_stations.country,
    radio_browser_stations.countrycode,
    radio_browser_stations.state,
    radio_browser_stations.language,
    radio_browser_stations.codec,
    radio_browser_stations.bitrate,
    radio_browser_stations.votes,
    radio_browser_stations.tags,
    radio_browser_stations.last_check_ok
) IS DISTINCT FROM (
    EXCLUDED.name,
    EXCLUDED.url,
    EXCLUDED.url_resolved,
    EXCLUDED.homepage,
    EXCLUDED.favicon,
    EXCLUDED.country,
    EXCLUDED.countrycode,
    EXCLUDED.state,
    EXCLUDED.language,
    EXCLUDED.codec,
    EXCLUDED.bitrate,
    EXCLUDED.votes,
    EXCLUDED.tags,
    EXCLUDED.last_check_ok
);

-- name: ListRadioBrowserStations :many
SELECT *
FROM radio_browser_stations
WHERE (sqlc.arg(name) = '' OR name ILIKE '%' || sqlc.arg(name) || '%')
  AND (sqlc.arg(country) = '' OR country ILIKE sqlc.arg(country))
  AND (sqlc.arg(language) = '' OR language ILIKE sqlc.arg(language))
ORDER BY votes DESC, name ASC
LIMIT sqlc.arg(page_limit) OFFSET sqlc.arg(page_offset);

-- name: CountRadioBrowserStations :one
SELECT COUNT(*)::bigint
FROM radio_browser_stations
WHERE (sqlc.arg(name) = '' OR name ILIKE '%' || sqlc.arg(name) || '%')
  AND (sqlc.arg(country) = '' OR country ILIKE sqlc.arg(country))
  AND (sqlc.arg(language) = '' OR language ILIKE sqlc.arg(language));
