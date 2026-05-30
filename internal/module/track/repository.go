package track

import (
	"context"
	"errors"
	"time"
)

var (
	// ErrRepoNotFound indicates that the repository could not find a requested row.
	ErrRepoNotFound = errors.New("track repository: not found")
	// ErrRepoStationNotFound indicates that a track references a missing station.
	ErrRepoStationNotFound = errors.New("track repository: station not found")
)

// Track is the domain model for a track.
// This is the only type that crosses layer boundaries inside this module.
// HTTP DTOs and DB rows map to/from Track at their respective edges.
type Track struct {
	ID              int64
	StationID       int64
	Title           string
	Artist          string
	AudioUrl        string
	DurationSeconds int32
	CreatedAt       time.Time
	UpdatedAt       time.Time
}

// CreateInput is data required to insert a track (repository layer).
type CreateInput struct {
	StationID       int64
	Title           string
	Artist          string
	AudioUrl        string
	DurationSeconds int32
}

// UpdateInput is the full column set passed to SQLC UpdateTrack after merge.
// StationID is included for the ownership check in the SQL WHERE clause.
type UpdateInput struct {
	ID              int64
	StationID       int64
	Title           string
	Artist          string
	AudioUrl        string
	DurationSeconds int32
}

// Repository loads and mutates tracks in the database.
// All methods return the domain Track model — TrackRow never escapes this package.
type Repository interface {
	Create(ctx context.Context, in CreateInput) (Track, error)
	GetByID(ctx context.Context, id int64) (Track, error)
	ListByStation(ctx context.Context, stationID, limit, offset int64) ([]Track, error)
	CountByStation(ctx context.Context, stationID int64) (int64, error)
	Update(ctx context.Context, in UpdateInput) (Track, error)
	Delete(ctx context.Context, id int64) error
}
