package track

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
)

// Sentinel errors for HTTP mapping in handlers.
var ErrNotFound = errors.New("track not found")

const (
	defaultListLimit = 20
	maxListLimit     = 100
)

// CreateTrackInput is the service-layer input for creating a track.
type CreateTrackInput struct {
	StationID       int64
	Title           string
	Artist          string
	AudioUrl        string
	DurationSeconds int32
}

// UpdateTrackRequest is the service-layer input for a partial update.
// Nil pointer fields are left unchanged (read-then-merge strategy).
type UpdateTrackRequest struct {
	Title           *string
	Artist          *string
	AudioUrl        *string
	DurationSeconds *int32
}

// Service defines the track business operations.
// All methods operate on the domain Track model — no HTTP or DB types cross this boundary.
type Service interface {
	Create(ctx context.Context, in CreateTrackInput) (Track, error)
	GetByID(ctx context.Context, id int64) (Track, error)
	List(ctx context.Context, stationID, limit, offset int64) ([]Track, int64, error)
	Update(ctx context.Context, id, stationID int64, req UpdateTrackRequest) (Track, error)
	Delete(ctx context.Context, id int64) error
}

type service struct {
	repo Repository
}

// NewService constructs a Service.
func NewService(repo Repository) Service {
	return &service{repo: repo}
}

func (s *service) Create(ctx context.Context, in CreateTrackInput) (Track, error) {
	return s.repo.Create(ctx, CreateInput{
		StationID:       in.StationID,
		Title:           in.Title,
		Artist:          in.Artist,
		AudioUrl:        in.AudioUrl,
		DurationSeconds: in.DurationSeconds,
	})
}

func (s *service) GetByID(ctx context.Context, id int64) (Track, error) {
	track, err := s.repo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return Track{}, ErrNotFound
		}
		return Track{}, err
	}
	return track, nil
}

func (s *service) List(ctx context.Context, stationID, limit, offset int64) ([]Track, int64, error) {
	total, err := s.repo.CountByStation(ctx, stationID)
	if err != nil {
		return nil, 0, err
	}
	tracks, err := s.repo.ListByStation(ctx, stationID, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	return tracks, total, nil
}

func (s *service) Update(ctx context.Context, id, stationID int64, req UpdateTrackRequest) (Track, error) {
	prev, err := s.repo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return Track{}, ErrNotFound
		}
		return Track{}, err
	}
	// Read-then-merge: only overwrite fields supplied by the caller.
	merged := UpdateInput{
		ID:              id,
		StationID:       stationID,
		Title:           prev.Title,
		Artist:          prev.Artist,
		AudioUrl:        prev.AudioUrl,
		DurationSeconds: prev.DurationSeconds,
	}
	if req.Title != nil {
		merged.Title = *req.Title
	}
	if req.Artist != nil {
		merged.Artist = *req.Artist
	}
	if req.AudioUrl != nil {
		merged.AudioUrl = *req.AudioUrl
	}
	if req.DurationSeconds != nil {
		merged.DurationSeconds = *req.DurationSeconds
	}
	track, err := s.repo.Update(ctx, merged)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return Track{}, ErrNotFound
		}
		return Track{}, err
	}
	return track, nil
}

func (s *service) Delete(ctx context.Context, id int64) error {
	_, err := s.repo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrNotFound
		}
		return err
	}
	return s.repo.Delete(ctx, id)
}
