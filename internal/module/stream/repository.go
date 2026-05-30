package stream

import (
	"context"
	"errors"
	"time"
)

var (
	// ErrRepoNotFound indicates that the repository could not find a requested stream row.
	ErrRepoNotFound = errors.New("stream repository: not found")
	// ErrRepoStationNotFound indicates that a stream references a missing station.
	ErrRepoStationNotFound = errors.New("stream repository: station not found")
)

// Stream is the domain model for a listening session.
// This is the only type that crosses layer boundaries inside this module.
type Stream struct {
	ID        int64
	UserID    int64
	StationID int64
	StartedAt time.Time
	EndedAt   *time.Time
	CreatedAt time.Time
}

// CreateInput is data required to insert a stream (repository layer).
type CreateInput struct {
	UserID    int64
	StationID int64
	StartedAt time.Time
}

// Repository loads and mutates streams in the database.
type Repository interface {
	Create(ctx context.Context, in CreateInput) (Stream, error)
	End(ctx context.Context, id int64, endedAt time.Time) (Stream, error)
	GetByID(ctx context.Context, id int64) (Stream, error)
	ListByUser(ctx context.Context, userID, limit, offset int64) ([]Stream, error)
	CountByUser(ctx context.Context, userID int64) (int64, error)
}
