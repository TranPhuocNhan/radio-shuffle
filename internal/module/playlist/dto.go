package playlist

import "time"

// CreatePlaylistRequest is the HTTP body for POST /playlists.
type CreatePlaylistRequest struct {
	Name        string  `json:"name" binding:"required"`
	Description *string `json:"description"`
	IsPublic    bool    `json:"is_public"`
}

// UpdatePlaylistRequest is the HTTP body for PATCH /playlists/:id.
// All fields are optional; omitted fields keep existing values.
type UpdatePlaylistRequest struct {
	Name        *string `json:"name"`
	Description *string `json:"description"`
	IsPublic    *bool   `json:"is_public"`
}

// PlaylistResponse is the JSON shape returned for a single playlist.
type PlaylistResponse struct {
	ID          int64     `json:"id"`
	Name        string    `json:"name"`
	Description *string   `json:"description"`
	IsPublic    bool      `json:"is_public"`
	OwnerID     int64     `json:"owner_id"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// AddPlaylistTrackRequest is the HTTP body for POST /playlists/:id/tracks.
type AddPlaylistTrackRequest struct {
	TrackID  int64 `json:"track_id" binding:"required"`
	Position int32 `json:"position"`
}

// ReorderPlaylistTracksRequest is the HTTP body for PATCH /playlists/:id/tracks/reorder.
type ReorderPlaylistTracksRequest struct {
	Items []PlaylistTrackPosition `json:"items" binding:"required"`
}

// PlaylistTrackPosition is a single track position in a reorder request.
type PlaylistTrackPosition struct {
	TrackID  int64 `json:"track_id" binding:"required"`
	Position int32 `json:"position"`
}

// PlaylistTrackResponse is the JSON shape returned for a playlist track.
type PlaylistTrackResponse struct {
	ID              int64     `json:"id"`
	StationID       int64     `json:"station_id"`
	Title           string    `json:"title"`
	Artist          string    `json:"artist"`
	AudioUrl        string    `json:"audio_url"`
	DurationSeconds int32     `json:"duration_seconds"`
	Position        int32     `json:"position"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}

func toPlaylistResponse(p Playlist) PlaylistResponse {
	return PlaylistResponse{
		ID:          p.ID,
		Name:        p.Name,
		Description: p.Description,
		IsPublic:    p.IsPublic,
		OwnerID:     p.OwnerID,
		CreatedAt:   p.CreatedAt,
		UpdatedAt:   p.UpdatedAt,
	}
}

func toPlaylistTrackResponse(t PlaylistTrack) PlaylistTrackResponse {
	return PlaylistTrackResponse{
		ID:              t.ID,
		StationID:       t.StationID,
		Title:           t.Title,
		Artist:          t.Artist,
		AudioUrl:        t.AudioUrl,
		DurationSeconds: t.DurationSeconds,
		Position:        t.Position,
		CreatedAt:       t.CreatedAt,
		UpdatedAt:       t.UpdatedAt,
	}
}
