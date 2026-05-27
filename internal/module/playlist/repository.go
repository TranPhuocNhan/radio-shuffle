package playlist

import (
	"context"
	"errors"
	"time"
)

var (
	// ErrRepoNotFound indicates that the repository could not find a requested row.
	ErrRepoNotFound = errors.New("playlist repository: not found")
	// ErrRepoDuplicate indicates duplicate key violation in repository operations.
	ErrRepoDuplicate = errors.New("playlist repository: duplicate")
	// ErrRepoInvalid indicates check/validation violation enforced by database.
	ErrRepoInvalid = errors.New("playlist repository: invalid input")
)

// Playlist is the domain model for a playlist.
// This is the only type that crosses layer boundaries inside this module.
type Playlist struct {
	ID          int64
	Name        string
	Description *string
	IsPublic    bool
	OwnerID     int64
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

// PlaylistTrack is the domain model for a track in a playlist.
type PlaylistTrack struct {
	ID              int64
	StationID       int64
	Title           string
	Artist          string
	AudioUrl        string
	DurationSeconds int32
	Position        int32
	CreatedAt       time.Time
	UpdatedAt       time.Time
}

// CreateInput is data required to insert a playlist (repository layer).
type CreateInput struct {
	Name        string
	Description *string
	IsPublic    bool
	OwnerID     int64
}

// UpdateInput is the full column set passed to SQLC UpdatePlaylist after merge.
type UpdateInput struct {
	ID          int64
	Name        string
	Description *string
	IsPublic    bool
}

// TrackPositionUpdate is a single playlist track position update.
type TrackPositionUpdate struct {
	TrackID  int64
	Position int32
}

// Repository loads and mutates playlists in the database.
type Repository interface {
	Create(ctx context.Context, in CreateInput) (Playlist, error)
	GetByID(ctx context.Context, id int64) (Playlist, error)
	ListByOwner(ctx context.Context, ownerID, limit, offset int64) ([]Playlist, error)
	CountByOwner(ctx context.Context, ownerID int64) (int64, error)
	Update(ctx context.Context, in UpdateInput) (Playlist, error)
	Delete(ctx context.Context, id int64) error

	AddTrack(ctx context.Context, playlistID, trackID int64, position int32) error
	RemoveTrack(ctx context.Context, playlistID, trackID int64) error
	ListTracks(ctx context.Context, playlistID, limit, offset int64) ([]PlaylistTrack, error)
	ListTrackIDs(ctx context.Context, playlistID int64) ([]int64, error)
	CountTracks(ctx context.Context, playlistID int64) (int64, error)
	ReorderTracks(ctx context.Context, playlistID int64, items []TrackPositionUpdate) error
}
