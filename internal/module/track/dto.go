package track

import "time"

// CreateTrackRequest is the HTTP body for POST /stations/:station_id/tracks.
type CreateTrackRequest struct {
	Title           string `json:"title"            binding:"required"`
	Artist          string `json:"artist"`
	AudioUrl        string `json:"audio_url"`
	DurationSeconds int32  `json:"duration_seconds"`
}

// UpdateTrackRequestBody is the HTTP body for PATCH /stations/:station_id/tracks/:id.
// All fields are optional; omitted fields keep existing values.
type UpdateTrackRequestBody struct {
	Title           *string `json:"title"`
	Artist          *string `json:"artist"`
	AudioUrl        *string `json:"audio_url"`
	DurationSeconds *int32  `json:"duration_seconds"`
}

// TrackResponse is the JSON shape returned for a single track.
type TrackResponse struct {
	ID              int64     `json:"id"`
	StationID       int64     `json:"station_id"`
	Title           string    `json:"title"`
	Artist          string    `json:"artist"`
	AudioUrl        string    `json:"audio_url"`
	DurationSeconds int32     `json:"duration_seconds"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}

func toTrackResponse(t Track) TrackResponse {
	return TrackResponse{
		ID:              t.ID,
		StationID:       t.StationID,
		Title:           t.Title,
		Artist:          t.Artist,
		AudioUrl:        t.AudioUrl,
		DurationSeconds: t.DurationSeconds,
		CreatedAt:       t.CreatedAt,
		UpdatedAt:       t.UpdatedAt,
	}
}
