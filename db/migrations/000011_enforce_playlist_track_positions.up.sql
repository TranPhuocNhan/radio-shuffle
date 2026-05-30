WITH ranked AS (
    SELECT
        playlist_id,
        track_id,
        ROW_NUMBER() OVER (
            PARTITION BY playlist_id
            ORDER BY position ASC, track_id ASC
        ) - 1 AS new_position
    FROM playlist_tracks
)
UPDATE playlist_tracks pt
SET position = ranked.new_position
FROM ranked
WHERE pt.playlist_id = ranked.playlist_id
  AND pt.track_id = ranked.track_id;

ALTER TABLE playlist_tracks
    ADD CONSTRAINT playlist_tracks_position_nonnegative
    CHECK (position >= 0);

ALTER TABLE playlist_tracks
    ADD CONSTRAINT playlist_tracks_playlist_position_unique
    UNIQUE (playlist_id, position)
    DEFERRABLE INITIALLY IMMEDIATE;
