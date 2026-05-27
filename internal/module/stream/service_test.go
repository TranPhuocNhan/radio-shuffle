package stream

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
)

// fakeRepo is an in-memory stub implementing Repository for service tests.
type fakeRepo struct {
	streams []Stream
	nextID  int64
}

func (f *fakeRepo) Create(_ context.Context, in CreateInput) (Stream, error) {
	f.nextID++
	row := Stream{
		ID:        f.nextID,
		UserID:    in.UserID,
		StationID: in.StationID,
		StartedAt: in.StartedAt,
		CreatedAt: time.Now(),
	}
	f.streams = append(f.streams, row)
	return row, nil
}

func (f *fakeRepo) End(_ context.Context, id int64, endedAt time.Time) (Stream, error) {
	for i, s := range f.streams {
		if s.ID == id {
			f.streams[i].EndedAt = &endedAt
			return f.streams[i], nil
		}
	}
	return Stream{}, pgx.ErrNoRows
}

func (f *fakeRepo) GetByID(_ context.Context, id int64) (Stream, error) {
	for _, s := range f.streams {
		if s.ID == id {
			return s, nil
		}
	}
	return Stream{}, pgx.ErrNoRows
}

func (f *fakeRepo) ListByUser(_ context.Context, userID, _, _ int64) ([]Stream, error) {
	var out []Stream
	for _, s := range f.streams {
		if s.UserID == userID {
			out = append(out, s)
		}
	}
	return out, nil
}

func (f *fakeRepo) CountByUser(_ context.Context, userID int64) (int64, error) {
	var count int64
	for _, s := range f.streams {
		if s.UserID == userID {
			count++
		}
	}
	return count, nil
}

func TestService_GetByID_NotFound(t *testing.T) {
	svc := NewService(&fakeRepo{})
	_, err := svc.GetByID(context.Background(), 1, 1)
	if !errors.Is(err, ErrNotFound) {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
}

func TestService_GetByID_Forbidden(t *testing.T) {
	repo := &fakeRepo{streams: []Stream{{ID: 1, UserID: 2}}}
	svc := NewService(repo)
	_, err := svc.GetByID(context.Background(), 1, 1)
	if !errors.Is(err, ErrForbidden) {
		t.Fatalf("expected ErrForbidden, got %v", err)
	}
}

func TestService_End_AlreadyEnded(t *testing.T) {
	ended := time.Now().Add(-time.Minute)
	repo := &fakeRepo{streams: []Stream{{ID: 1, UserID: 1, EndedAt: &ended}}}
	svc := NewService(repo)
	_, err := svc.End(context.Background(), 1, 1)
	if !errors.Is(err, ErrAlreadyEnded) {
		t.Fatalf("expected ErrAlreadyEnded, got %v", err)
	}
}

func TestService_Start_OK(t *testing.T) {
	svc := NewService(&fakeRepo{})
	row, err := svc.Start(context.Background(), 1, 10)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if row.ID == 0 {
		t.Fatalf("expected non-zero ID")
	}
	if row.StationID != 10 {
		t.Fatalf("station_id = %d", row.StationID)
	}
}
