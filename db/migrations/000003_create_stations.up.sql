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
