package stream

import (
	"context"
	"time"

	"github.com/tranphuocnhan/radio-shuffle/pkg/dbsqlc"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

type sqlcRepository struct {
	q *dbsqlc.Queries
}

// NewRepository returns a SQLC-backed Repository.
func NewRepository(pool *pgxpool.Pool) Repository {
	return &sqlcRepository{q: dbsqlc.New(pool)}
}

// streamFromDB maps a generated SQLC row to the domain Stream model.
func streamFromDB(s dbsqlc.Streams) Stream {
	var endedAt *time.Time
	if s.EndedAt.Valid {
		ended := s.EndedAt.Time
		endedAt = &ended
	}
	return Stream{
		ID:        s.ID,
		UserID:    s.UserID,
		StationID: s.StationID,
		StartedAt: s.StartedAt.Time,
		EndedAt:   endedAt,
		CreatedAt: s.CreatedAt.Time,
	}
}

func (r *sqlcRepository) Create(ctx context.Context, in CreateInput) (Stream, error) {
	row, err := r.q.CreateStream(ctx, dbsqlc.CreateStreamParams{
		UserID:    in.UserID,
		StationID: in.StationID,
		StartedAt: pgtype.Timestamptz{Time: in.StartedAt, Valid: true},
		EndedAt:   pgtype.Timestamptz{Valid: false},
	})
	if err != nil {
		return Stream{}, err
	}
	return streamFromDB(row), nil
}

func (r *sqlcRepository) End(ctx context.Context, id int64, endedAt time.Time) (Stream, error) {
	row, err := r.q.EndStreamByID(ctx, dbsqlc.EndStreamByIDParams{
		ID:      id,
		EndedAt: pgtype.Timestamptz{Time: endedAt, Valid: true},
	})
	if err != nil {
		return Stream{}, err
	}
	return streamFromDB(row), nil
}

func (r *sqlcRepository) GetByID(ctx context.Context, id int64) (Stream, error) {
	row, err := r.q.GetStreamByID(ctx, id)
	if err != nil {
		return Stream{}, err
	}
	return streamFromDB(row), nil
}

func (r *sqlcRepository) ListByUser(ctx context.Context, userID, limit, offset int64) ([]Stream, error) {
	rows, err := r.q.ListStreamsByUserID(ctx, dbsqlc.ListStreamsByUserIDParams{
		UserID: userID,
		Limit:  int32(limit),
		Offset: int32(offset),
	})
	if err != nil {
		return nil, err
	}
	out := make([]Stream, 0, len(rows))
	for _, row := range rows {
		out = append(out, streamFromDB(row))
	}
	return out, nil
}

func (r *sqlcRepository) CountByUser(ctx context.Context, userID int64) (int64, error) {
	return r.q.CountStreamsByUserID(ctx, userID)
}
