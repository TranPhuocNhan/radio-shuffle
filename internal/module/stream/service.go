package stream

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
)

// Sentinel errors for HTTP mapping in handlers.
var (
	ErrNotFound     = errors.New("stream not found")
	ErrForbidden    = errors.New("forbidden")
	ErrAlreadyEnded = errors.New("stream already ended")
)

const (
	defaultListLimit = 20
	maxListLimit     = 100
)

// Service defines the stream business operations.
type Service interface {
	Start(ctx context.Context, userID, stationID int64) (Stream, error)
	End(ctx context.Context, id, userID int64) (Stream, error)
	GetByID(ctx context.Context, id, userID int64) (Stream, error)
	List(ctx context.Context, userID, limit, offset int64) ([]Stream, int64, error)
}

type service struct {
	repo Repository
}

// NewService constructs a Service.
func NewService(repo Repository) Service {
	return &service{repo: repo}
}

func (s *service) Start(ctx context.Context, userID, stationID int64) (Stream, error) {
	return s.repo.Create(ctx, CreateInput{
		UserID:    userID,
		StationID: stationID,
		StartedAt: time.Now(),
	})
}

func (s *service) End(ctx context.Context, id, userID int64) (Stream, error) {
	stream, err := s.getByID(ctx, id)
	if err != nil {
		return Stream{}, err
	}
	if stream.UserID != userID {
		return Stream{}, ErrForbidden
	}
	if stream.EndedAt != nil {
		return Stream{}, ErrAlreadyEnded
	}
	ended, err := s.repo.End(ctx, id, time.Now())
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return Stream{}, ErrNotFound
		}
		return Stream{}, err
	}
	return ended, nil
}

func (s *service) GetByID(ctx context.Context, id, userID int64) (Stream, error) {
	stream, err := s.getByID(ctx, id)
	if err != nil {
		return Stream{}, err
	}
	if stream.UserID != userID {
		return Stream{}, ErrForbidden
	}
	return stream, nil
}

func (s *service) List(ctx context.Context, userID, limit, offset int64) ([]Stream, int64, error) {
	total, err := s.repo.CountByUser(ctx, userID)
	if err != nil {
		return nil, 0, err
	}
	items, err := s.repo.ListByUser(ctx, userID, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	return items, total, nil
}

func (s *service) getByID(ctx context.Context, id int64) (Stream, error) {
	stream, err := s.repo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return Stream{}, ErrNotFound
		}
		return Stream{}, err
	}
	return stream, nil
}
