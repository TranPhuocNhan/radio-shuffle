package station

import (
	"context"
	"time"
)

// Station is the domain representation of a station.
type Station struct {
	ID            int64
	Name          string
	Genre         string
	Description   *string
	StreamUrl     string
	CoverImageUrl *string
	IsPublic      bool
	OwnerID       *int64
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

// CreateInput is data required to insert a station.
type CreateInput struct {
	Name          string
	Genre         string
	Description   *string
	StreamUrl     string
	CoverImageUrl *string
	IsPublic      bool
	OwnerID       *int64
}

// UpdateInput is the full column set passed to SQLC UpdateStation after merge.
type UpdateInput struct {
	ID            int64
	Name          string
	Genre         string
	Description   *string
	StreamUrl     string
	CoverImageUrl *string
	IsPublic      bool
	OwnerID       *int64
}

// Repository loads and mutates stations in the database.
type Repository interface {
	Create(ctx context.Context, in CreateInput) (Station, error)
	GetByID(ctx context.Context, id int64) (Station, error)
	List(ctx context.Context, limit, offset int64) ([]Station, error)
	Count(ctx context.Context) (int64, error)
	Update(ctx context.Context, in UpdateInput) (Station, error)
	Delete(ctx context.Context, id int64) error
}
