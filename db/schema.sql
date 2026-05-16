-- Canonical schema snapshot for sqlc — keep in sync with db/migrations

CREATE TABLE users (
    id BIGSERIAL PRIMARY KEY,
    email TEXT NOT NULL UNIQUE,
    password_hash BYTEA NOT NULL,
    role TEXT NOT NULL DEFAULT 'user',
    CONSTRAINT users_role_ck CHECK (role IN ('admin', 'user')),
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE refresh_tokens (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    token_hash BYTEA NOT NULL UNIQUE,
    expires_at TIMESTAMPTZ NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX idx_refresh_tokens_user_id ON refresh_tokens(user_id);

CREATE TABLE stations (
    id BIGSERIAL PRIMARY KEY,
    name TEXT NOT NULL,
    genre TEXT NOT NULL DEFAULT '',
    description TEXT DEFAULT '',
    stream_url TEXT NOT NULL DEFAULT '',
    cover_image_url TEXT DEFAULT '',
    is_public BOOLEAN NOT NULL DEFAULT TRUE,
    owner_id BIGINT REFERENCES users(id) ON DELETE SET NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX idx_stations_owner_id ON stations(owner_id);
CREATE INDEX idx_stations_is_public ON stations(is_public);
CREATE INDEX idx_stations_created_at ON stations(created_at DESC);

CREATE TABLE tracks (
    id BIGSERIAL PRIMARY KEY,
    station_id BIGINT NOT NULL REFERENCES stations(id) ON DELETE CASCADE,
    title TEXT NOT NULL,
    artist TEXT NOT NULL DEFAULT '',
    audio_url TEXT NOT NULL DEFAULT '',
    duration_seconds INT NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX idx_tracks_station_id ON tracks(station_id);

CREATE TABLE playlists (
    id BIGSERIAL PRIMARY KEY,
    name TEXT NOT NULL,
    description TEXT DEFAULT '',
    is_public BOOLEAN NOT NULL DEFAULT FALSE,
    owner_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX idx_playlists_owner ON playlists(owner_id);
CREATE INDEX idx_playlists_public ON playlists(is_public);

CREATE TABLE playlist_tracks (
    playlist_id BIGINT NOT NULL REFERENCES playlists(id) ON DELETE CASCADE,
    track_id BIGINT NOT NULL REFERENCES tracks(id) ON DELETE CASCADE,
    position INT NOT NULL DEFAULT 0,
    PRIMARY KEY (playlist_id, track_id)
);

CREATE TABLE streams (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    station_id BIGINT NOT NULL REFERENCES stations(id) ON DELETE CASCADE,
    started_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    ended_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX idx_streams_user ON streams(user_id);
CREATE INDEX idx_streams_station ON streams(station_id);
CREATE INDEX idx_streams_started ON streams(started_at DESC);

CREATE TABLE radio_browser_stations (
    id              BIGSERIAL PRIMARY KEY,
    stationuuid     TEXT UNIQUE NOT NULL,
    name            TEXT NOT NULL,
    url             TEXT NOT NULL,
    url_resolved    TEXT,
    homepage        TEXT,
    favicon         TEXT,
    country         TEXT,
    countrycode     TEXT,
    state           TEXT,
    language        TEXT,
    codec           TEXT,
    bitrate         INT NOT NULL DEFAULT 0,
    votes           INT NOT NULL DEFAULT 0,
    tags            TEXT,
    last_check_ok   BOOLEAN NOT NULL DEFAULT FALSE,
    synced_at       TIMESTAMPTZ NOT NULL DEFAULT now(),
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX idx_rb_stations_countrycode ON radio_browser_stations (countrycode);
CREATE INDEX idx_rb_stations_tags ON radio_browser_stations USING gin (to_tsvector('simple', COALESCE(tags, '')));

CREATE TABLE users_stations (
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    station_id BIGINT NOT NULL REFERENCES stations(id) ON DELETE CASCADE,
    PRIMARY KEY (user_id, station_id)
);
CREATE INDEX idx_users_stations_station_id ON users_stations(station_id);