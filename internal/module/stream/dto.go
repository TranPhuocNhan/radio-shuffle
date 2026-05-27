package stream

import "time"

// StartStreamRequest is the HTTP body for POST /streams.
type StartStreamRequest struct {
	StationID int64 `json:"station_id" binding:"required"`
}

// StreamResponse is the JSON shape returned for a stream.
type StreamResponse struct {
	ID        int64      `json:"id"`
	UserID    int64      `json:"user_id"`
	StationID int64      `json:"station_id"`
	StartedAt time.Time  `json:"started_at"`
	EndedAt   *time.Time `json:"ended_at"`
	CreatedAt time.Time  `json:"created_at"`
}

func toStreamResponse(s Stream) StreamResponse {
	return StreamResponse{
		ID:        s.ID,
		UserID:    s.UserID,
		StationID: s.StationID,
		StartedAt: s.StartedAt,
		EndedAt:   s.EndedAt,
		CreatedAt: s.CreatedAt,
	}
}
