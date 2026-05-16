CREATE TABLE users_stations (
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    station_id BIGINT NOT NULL REFERENCES stations(id) ON DELETE CASCADE,
    PRIMARY KEY (user_id, station_id)
);
CREATE INDEX idx_users_stations_station_id ON users_stations(station_id);