package track

import (
	"context"
	"errors"

	"github.com/tranphuocnhan/radio-shuffle/pkg/dbsqlc"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

type sqlcRepository struct {
	q *dbsqlc.Queries
}

// NewRepository returns a SQLC-backed Repository.
func NewRepository(pool *pgxpool.Pool) Repository {
	return &sqlcRepository{q: dbsqlc.New(pool)}
}

// rowFromDB maps a generated SQLC row to the domain Track model.
// This is the only place in the module where dbsqlc.Tracks is touched.
func rowFromDB(t dbsqlc.Tracks) Track {
	return Track{
		ID:              t.ID,
		StationID:       t.StationID,
		Title:           t.Title,
		Artist:          t.Artist,
		AudioUrl:        t.AudioUrl,
		DurationSeconds: t.DurationSeconds,
		CreatedAt:       t.CreatedAt.Time,
		UpdatedAt:       t.UpdatedAt.Time,
	}
}

func (r *sqlcRepository) Create(ctx context.Context, in CreateInput) (Track, error) {
	t, err := r.q.CreateTrack(ctx, dbsqlc.CreateTrackParams{
		StationID:       in.StationID,
		Title:           in.Title,
		Artist:          in.Artist,
		AudioUrl:        in.AudioUrl,
		DurationSeconds: in.DurationSeconds,
	})
	if err != nil {
		return Track{}, mapPgError(err)
	}
	return rowFromDB(t), nil
}

func (r *sqlcRepository) GetByID(ctx context.Context, id int64) (Track, error) {
	t, err := r.q.GetTrackByID(ctx, id)
	if err != nil {
		return Track{}, mapNoRowsOrPgError(err)
	}
	return rowFromDB(t), nil
}

func (r *sqlcRepository) ListByStation(ctx context.Context, stationID, limit, offset int64) ([]Track, error) {
	rows, err := r.q.ListTracksByStationID(ctx, dbsqlc.ListTracksByStationIDParams{
		StationID: stationID,
		Limit:     int32(limit),
		Offset:    int32(offset),
	})
	if err != nil {
		return nil, err
	}
	out := make([]Track, 0, len(rows))
	for _, t := range rows {
		out = append(out, rowFromDB(t))
	}
	return out, nil
}

func (r *sqlcRepository) CountByStation(ctx context.Context, stationID int64) (int64, error) {
	return r.q.CountTracksByStationID(ctx, stationID)
}

func (r *sqlcRepository) Update(ctx context.Context, in UpdateInput) (Track, error) {
	t, err := r.q.UpdateTrack(ctx, dbsqlc.UpdateTrackParams{
		ID:              in.ID,
		StationID:       in.StationID,
		Title:           in.Title,
		Artist:          in.Artist,
		AudioUrl:        in.AudioUrl,
		DurationSeconds: in.DurationSeconds,
	})
	if err != nil {
		return Track{}, mapNoRowsOrPgError(err)
	}
	return rowFromDB(t), nil
}

func (r *sqlcRepository) Delete(ctx context.Context, id int64) error {
	return r.q.DeleteTrack(ctx, id)
}

func mapNoRowsOrPgError(err error) error {
	if errors.Is(err, pgx.ErrNoRows) {
		return ErrRepoNotFound
	}
	return mapPgError(err)
}

func mapPgError(err error) error {
	if err == nil {
		return nil
	}
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		if pgErr.Code == "23503" && pgErr.ConstraintName == "tracks_station_id_fkey" {
			return ErrRepoStationNotFound
		}
	}
	return err
}
