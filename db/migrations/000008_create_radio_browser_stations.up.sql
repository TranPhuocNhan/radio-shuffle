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
