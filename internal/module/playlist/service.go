package playlist

import (
	"context"
	"errors"
)

// Sentinel errors for HTTP mapping in handlers.
var (
	ErrNotFound          = errors.New("playlist not found")
	ErrForbidden         = errors.New("forbidden")
	ErrTrackNotFound     = errors.New("track not found")
	ErrDuplicateTrack    = errors.New("duplicate track")
	ErrDuplicatePosition = errors.New("duplicate position")
	ErrInvalidPosition   = errors.New("invalid position")
	ErrInvalidTrackOrder = errors.New("invalid track reorder payload")
)

const (
	defaultListLimit = 20
	maxListLimit     = 100
)

// CreatePlaylistInput is the service-layer input for creating a playlist.
type CreatePlaylistInput struct {
	Name        string
	Description *string
	IsPublic    bool
}

// UpdatePlaylistInput is the service-layer input for a partial update.
// Nil pointer fields are left unchanged (read-then-merge strategy).
type UpdatePlaylistInput struct {
	Name        *string
	Description *string
	IsPublic    *bool
}

// TrackPositionPatch is a single playlist track position patch.
type TrackPositionPatch struct {
	TrackID  int64
	Position int32
}

// Service defines the playlist business operations.
type Service interface {
	Create(ctx context.Context, ownerID int64, in CreatePlaylistInput) (Playlist, error)
	GetByID(ctx context.Context, id, requesterID int64) (Playlist, error)
	List(ctx context.Context, ownerID, limit, offset int64) ([]Playlist, int64, error)
	Update(ctx context.Context, id, ownerID int64, in UpdatePlaylistInput) (Playlist, error)
	Delete(ctx context.Context, id, ownerID int64) error
	AddTrack(ctx context.Context, playlistID, ownerID, trackID int64, position int32) error
	RemoveTrack(ctx context.Context, playlistID, ownerID, trackID int64) error
	ListTracks(ctx context.Context, playlistID, requesterID, limit, offset int64) ([]PlaylistTrack, int64, error)
	ReorderTracks(ctx context.Context, playlistID, ownerID int64, items []TrackPositionPatch) error
}

type service struct {
	repo Repository
}

// NewService constructs a Service.
func NewService(repo Repository) Service {
	return &service{repo: repo}
}

func (s *service) Create(ctx context.Context, ownerID int64, in CreatePlaylistInput) (Playlist, error) {
	return s.repo.Create(ctx, CreateInput{
		Name:        in.Name,
		Description: in.Description,
		IsPublic:    in.IsPublic,
		OwnerID:     ownerID,
	})
}

func (s *service) GetByID(ctx context.Context, id, requesterID int64) (Playlist, error) {
	playlist, err := s.getReadablePlaylist(ctx, id, requesterID)
	if err != nil {
		return Playlist{}, err
	}
	return playlist, nil
}

func (s *service) List(ctx context.Context, ownerID, limit, offset int64) ([]Playlist, int64, error) {
	total, err := s.repo.CountByOwner(ctx, ownerID)
	if err != nil {
		return nil, 0, err
	}
	items, err := s.repo.ListByOwner(ctx, ownerID, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	return items, total, nil
}

func (s *service) Update(ctx context.Context, id, ownerID int64, in UpdatePlaylistInput) (Playlist, error) {
	prev, err := s.getOwnedPlaylist(ctx, id, ownerID)
	if err != nil {
		return Playlist{}, err
	}
	merged := UpdateInput{
		ID:          id,
		Name:        prev.Name,
		Description: prev.Description,
		IsPublic:    prev.IsPublic,
	}
	if in.Name != nil {
		merged.Name = *in.Name
	}
	if in.Description != nil {
		merged.Description = in.Description
	}
	if in.IsPublic != nil {
		merged.IsPublic = *in.IsPublic
	}
	playlist, err := s.repo.Update(ctx, merged)
	if err != nil {
		if errors.Is(err, ErrRepoNotFound) {
			return Playlist{}, ErrNotFound
		}
		return Playlist{}, err
	}
	return playlist, nil
}

func (s *service) Delete(ctx context.Context, id, ownerID int64) error {
	_, err := s.getOwnedPlaylist(ctx, id, ownerID)
	if err != nil {
		return err
	}
	return s.repo.Delete(ctx, id)
}

func (s *service) AddTrack(ctx context.Context, playlistID, ownerID, trackID int64, position int32) error {
	_, err := s.getOwnedPlaylist(ctx, playlistID, ownerID)
	if err != nil {
		return err
	}
	if position < 0 {
		return ErrInvalidPosition
	}
	if err := s.repo.AddTrack(ctx, playlistID, trackID, position); err != nil {
		switch {
		case errors.Is(err, ErrRepoNotFound):
			return ErrTrackNotFound
		case errors.Is(err, ErrRepoDuplicate):
			return ErrDuplicateTrack
		case errors.Is(err, ErrRepoDuplicatePosition):
			return ErrDuplicatePosition
		case errors.Is(err, ErrRepoInvalid):
			return ErrInvalidPosition
		default:
			return err
		}
	}
	return nil
}

func (s *service) RemoveTrack(ctx context.Context, playlistID, ownerID, trackID int64) error {
	_, err := s.getOwnedPlaylist(ctx, playlistID, ownerID)
	if err != nil {
		return err
	}
	if err := s.repo.RemoveTrack(ctx, playlistID, trackID); err != nil {
		if errors.Is(err, ErrRepoNotFound) {
			return ErrTrackNotFound
		}
		return err
	}
	return nil
}

func (s *service) ListTracks(ctx context.Context, playlistID, requesterID, limit, offset int64) ([]PlaylistTrack, int64, error) {
	_, err := s.getReadablePlaylist(ctx, playlistID, requesterID)
	if err != nil {
		return nil, 0, err
	}
	total, err := s.repo.CountTracks(ctx, playlistID)
	if err != nil {
		return nil, 0, err
	}
	items, err := s.repo.ListTracks(ctx, playlistID, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	return items, total, nil
}

func (s *service) ReorderTracks(ctx context.Context, playlistID, ownerID int64, items []TrackPositionPatch) error {
	_, err := s.getOwnedPlaylist(ctx, playlistID, ownerID)
	if err != nil {
		return err
	}
	if err := validateReorderItems(items); err != nil {
		return err
	}
	updates := make([]TrackPositionUpdate, 0, len(items))
	for _, item := range items {
		updates = append(updates, TrackPositionUpdate{TrackID: item.TrackID, Position: item.Position})
	}
	if err := s.repo.ReorderTracks(ctx, playlistID, updates); err != nil {
		switch {
		case errors.Is(err, ErrRepoNotFound):
			return ErrTrackNotFound
		case errors.Is(err, ErrRepoInvalid):
			return ErrInvalidPosition
		case errors.Is(err, ErrRepoInvalidTrackSet):
			return ErrInvalidTrackOrder
		case errors.Is(err, ErrRepoDuplicatePosition):
			return ErrInvalidTrackOrder
		}
		return err
	}
	return nil
}

func (s *service) getByID(ctx context.Context, id int64) (Playlist, error) {
	playlist, err := s.repo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, ErrRepoNotFound) {
			return Playlist{}, ErrNotFound
		}
		return Playlist{}, err
	}
	return playlist, nil
}

func (s *service) getOwnedPlaylist(ctx context.Context, id, ownerID int64) (Playlist, error) {
	playlist, err := s.getByID(ctx, id)
	if err != nil {
		return Playlist{}, err
	}
	if playlist.OwnerID != ownerID {
		return Playlist{}, ErrForbidden
	}
	return playlist, nil
}

func (s *service) getReadablePlaylist(ctx context.Context, id, requesterID int64) (Playlist, error) {
	playlist, err := s.getByID(ctx, id)
	if err != nil {
		return Playlist{}, err
	}
	if !canRead(playlist, requesterID) {
		return Playlist{}, ErrForbidden
	}
	return playlist, nil
}

func canRead(playlist Playlist, requesterID int64) bool {
	return playlist.OwnerID == requesterID || playlist.IsPublic
}

func validateReorderItems(items []TrackPositionPatch) error {
	seenTrackIDs := make(map[int64]struct{}, len(items))
	seenPositions := make(map[int32]struct{}, len(items))
	for _, item := range items {
		if item.Position < 0 {
			return ErrInvalidPosition
		}
		if _, ok := seenTrackIDs[item.TrackID]; ok {
			return ErrInvalidTrackOrder
		}
		if _, ok := seenPositions[item.Position]; ok {
			return ErrInvalidTrackOrder
		}
		seenTrackIDs[item.TrackID] = struct{}{}
		seenPositions[item.Position] = struct{}{}
	}
	return nil
}
