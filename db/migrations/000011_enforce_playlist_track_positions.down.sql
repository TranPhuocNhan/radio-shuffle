ALTER TABLE playlist_tracks
    DROP CONSTRAINT IF EXISTS playlist_tracks_playlist_position_unique;

ALTER TABLE playlist_tracks
    DROP CONSTRAINT IF EXISTS playlist_tracks_position_nonnegative;
