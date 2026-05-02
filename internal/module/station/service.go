package station

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
)

// Sentinel errors for HTTP mapping in handlers.
var ErrNotFound = errors.New("station not found")

const (
	defaultListLimit = 20
	maxListLimit     = 100
)

// Service defines the station business operations.
type Service interface {
	Create(ctx context.Context, params CreateStationInput) (StationRow, error)
	GetByID(ctx context.Context, id int64) (StationResponse, error)
	List(ctx context.Context, limit, offset int64) ([]StationResponse, int64, error)
	Update(ctx context.Context, id int64, req UpdateStationRequest) (StationResponse, error)
	Delete(ctx context.Context, id int64) error
}

// CreateStationInput is the service-layer input for creating a station.
type CreateStationInput struct {
	Name          string
	Genre         string
	Description   *string
	StreamUrl     string
	CoverImageUrl *string
	IsPublic      *bool
	OwnerID       *int64
}

type service struct {
	repo Repository
}

// NewService constructs a Service.
func NewService(repo Repository) Service {
	return &service{repo: repo}
}

func (s *service) Create(ctx context.Context, params CreateStationInput) (StationRow, error) {
	isPublic := true
	if params.IsPublic != nil {
		isPublic = *params.IsPublic
	}
	row, err := s.repo.Create(ctx, CreateInput{
		Name:          params.Name,
		Genre:         params.Genre,
		Description:   params.Description,
		StreamUrl:     params.StreamUrl,
		CoverImageUrl: params.CoverImageUrl,
		IsPublic:      isPublic,
		OwnerID:       params.OwnerID,
	})
	if err != nil {
		return StationRow{}, err
	}
	return row, nil
}

func (s *service) GetByID(ctx context.Context, id int64) (StationResponse, error) {
	row, err := s.repo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return StationResponse{}, ErrNotFound
		}
		return StationResponse{}, err
	}
	return toStationResponse(row), nil
}

func (s *service) List(ctx context.Context, limit, offset int64) ([]StationResponse, int64, error) {
	total, err := s.repo.Count(ctx)
	if err != nil {
		return nil, 0, err
	}
	rows, err := s.repo.List(ctx, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	out := make([]StationResponse, 0, len(rows))
	for _, r := range rows {
		out = append(out, toStationResponse(r))
	}
	return out, total, nil
}

func (s *service) Update(ctx context.Context, id int64, req UpdateStationRequest) (StationResponse, error) {
	prev, err := s.repo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return StationResponse{}, ErrNotFound
		}
		return StationResponse{}, err
	}
	merged := UpdateInput{
		ID:            id,
		Name:          prev.Name,
		Genre:         prev.Genre,
		Description:   prev.Description,
		StreamUrl:     prev.StreamUrl,
		CoverImageUrl: prev.CoverImageUrl,
		IsPublic:      prev.IsPublic,
		OwnerID:       prev.OwnerID,
	}
	if req.Name != nil {
		merged.Name = *req.Name
	}
	if req.Genre != nil {
		merged.Genre = *req.Genre
	}
	if req.Description != nil {
		merged.Description = req.Description
	}
	if req.StreamUrl != nil {
		merged.StreamUrl = *req.StreamUrl
	}
	if req.CoverImageUrl != nil {
		merged.CoverImageUrl = req.CoverImageUrl
	}
	if req.IsPublic != nil {
		merged.IsPublic = *req.IsPublic
	}
	if req.OwnerID != nil {
		merged.OwnerID = req.OwnerID
	}
	row, err := s.repo.Update(ctx, merged)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return StationResponse{}, ErrNotFound
		}
		return StationResponse{}, err
	}
	return toStationResponse(row), nil
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
