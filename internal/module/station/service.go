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
	Create(ctx context.Context, params CreateStationInput) (Station, error)
	GetByID(ctx context.Context, id int64) (Station, error)
	List(ctx context.Context, filter ListStationsInput) ([]Station, int64, error)
	Update(ctx context.Context, input UpdateStationInput) (Station, error)
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

// UpdateStationInput is the service-layer input for updating a station.
type UpdateStationInput struct {
	ID            int64
	Name          *string
	Genre         *string
	Description   *string
	StreamUrl     *string
	CoverImageUrl *string
	IsPublic      *bool
	OwnerID       *int64
}

// ListStationsInput holds pagination data for listing stations.
type ListStationsInput struct {
	Limit  int64
	Offset int64
}

type service struct {
	repo Repository
}

// NewService constructs a Service.
func NewService(repo Repository) Service {
	return &service{repo: repo}
}

func (s *service) Create(ctx context.Context, params CreateStationInput) (Station, error) {
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
		return Station{}, err
	}
	return row, nil
}

func (s *service) GetByID(ctx context.Context, id int64) (Station, error) {
	row, err := s.repo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return Station{}, ErrNotFound
		}
		return Station{}, err
	}
	return row, nil
}

func (s *service) List(ctx context.Context, filter ListStationsInput) ([]Station, int64, error) {
	total, err := s.repo.Count(ctx)
	if err != nil {
		return nil, 0, err
	}
	rows, err := s.repo.List(ctx, filter.Limit, filter.Offset)
	if err != nil {
		return nil, 0, err
	}
	return rows, total, nil
}

func (s *service) Update(ctx context.Context, input UpdateStationInput) (Station, error) {
	prev, err := s.repo.GetByID(ctx, input.ID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return Station{}, ErrNotFound
		}
		return Station{}, err
	}
	merged := UpdateInput{
		ID:            input.ID,
		Name:          prev.Name,
		Genre:         prev.Genre,
		Description:   prev.Description,
		StreamUrl:     prev.StreamUrl,
		CoverImageUrl: prev.CoverImageUrl,
		IsPublic:      prev.IsPublic,
		OwnerID:       prev.OwnerID,
	}
	if input.Name != nil {
		merged.Name = *input.Name
	}
	if input.Genre != nil {
		merged.Genre = *input.Genre
	}
	if input.Description != nil {
		merged.Description = input.Description
	}
	if input.StreamUrl != nil {
		merged.StreamUrl = *input.StreamUrl
	}
	if input.CoverImageUrl != nil {
		merged.CoverImageUrl = input.CoverImageUrl
	}
	if input.IsPublic != nil {
		merged.IsPublic = *input.IsPublic
	}
	if input.OwnerID != nil {
		merged.OwnerID = input.OwnerID
	}
	row, err := s.repo.Update(ctx, merged)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return Station{}, ErrNotFound
		}
		return Station{}, err
	}
	return row, nil
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
