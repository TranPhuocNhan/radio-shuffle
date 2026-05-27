package playlist

import (
	"context"
	"errors"

	"github.com/tranphuocnhan/radio-shuffle/pkg/dbsqlc"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

type sqlcRepository struct {
	q    *dbsqlc.Queries
	pool *pgxpool.Pool
}

// NewRepository returns a SQLC-backed Repository.
func NewRepository(pool *pgxpool.Pool) Repository {
	return &sqlcRepository{
		q:    dbsqlc.New(pool),
		pool: pool,
	}
}

// playlistFromDB maps a generated SQLC row to the domain Playlist model.
func playlistFromDB(p dbsqlc.Playlists) Playlist {
	return Playlist{
		ID:          p.ID,
		Name:        p.Name,
		Description: p.Description,
		IsPublic:    p.IsPublic,
		OwnerID:     p.OwnerID,
		CreatedAt:   p.CreatedAt.Time,
		UpdatedAt:   p.UpdatedAt.Time,
	}
}

// trackFromDB maps a generated SQLC row to the domain PlaylistTrack model.
func trackFromDB(t dbsqlc.ListPlaylistTracksRow) PlaylistTrack {
	return PlaylistTrack{
		ID:              t.ID,
		StationID:       t.StationID,
		Title:           t.Title,
		Artist:          t.Artist,
		AudioUrl:        t.AudioUrl,
		DurationSeconds: t.DurationSeconds,
		Position:        t.Position,
		CreatedAt:       t.CreatedAt.Time,
		UpdatedAt:       t.UpdatedAt.Time,
	}
}

func (r *sqlcRepository) Create(ctx context.Context, in CreateInput) (Playlist, error) {
	row, err := r.q.CreatePlaylist(ctx, dbsqlc.CreatePlaylistParams{
		Name:        in.Name,
		Description: in.Description,
		IsPublic:    in.IsPublic,
		OwnerID:     in.OwnerID,
	})
	if err != nil {
		return Playlist{}, err
	}
	return playlistFromDB(row), nil
}

func (r *sqlcRepository) GetByID(ctx context.Context, id int64) (Playlist, error) {
	row, err := r.q.GetPlaylistByID(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return Playlist{}, ErrRepoNotFound
		}
		return Playlist{}, err
	}
	return playlistFromDB(row), nil
}

func (r *sqlcRepository) ListByOwner(ctx context.Context, ownerID, limit, offset int64) ([]Playlist, error) {
	rows, err := r.q.ListPlaylistsOwnedBy(ctx, dbsqlc.ListPlaylistsOwnedByParams{
		OwnerID: ownerID,
		Limit:   int32(limit),
		Offset:  int32(offset),
	})
	if err != nil {
		return nil, err
	}
	out := make([]Playlist, 0, len(rows))
	for _, row := range rows {
		out = append(out, playlistFromDB(row))
	}
	return out, nil
}

func (r *sqlcRepository) CountByOwner(ctx context.Context, ownerID int64) (int64, error) {
	return r.q.CountPlaylistsOwnedBy(ctx, ownerID)
}

func (r *sqlcRepository) Update(ctx context.Context, in UpdateInput) (Playlist, error) {
	row, err := r.q.UpdatePlaylist(ctx, dbsqlc.UpdatePlaylistParams{
		ID:          in.ID,
		Name:        in.Name,
		Description: in.Description,
		IsPublic:    in.IsPublic,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return Playlist{}, ErrRepoNotFound
		}
		return Playlist{}, err
	}
	return playlistFromDB(row), nil
}

func (r *sqlcRepository) Delete(ctx context.Context, id int64) error {
	return r.q.DeletePlaylist(ctx, id)
}

func (r *sqlcRepository) AddTrack(ctx context.Context, playlistID, trackID int64, position int32) error {
	err := r.q.AddTrackToPlaylist(ctx, dbsqlc.AddTrackToPlaylistParams{
		PlaylistID: playlistID,
		TrackID:    trackID,
		Position:   position,
	})
	return mapPgError(err)
}

func (r *sqlcRepository) RemoveTrack(ctx context.Context, playlistID, trackID int64) error {
	_, err := r.q.RemoveTrackFromPlaylist(ctx, dbsqlc.RemoveTrackFromPlaylistParams{
		PlaylistID: playlistID,
		TrackID:    trackID,
	})
	return mapNoRowsOrPgError(err)
}

func (r *sqlcRepository) ListTracks(ctx context.Context, playlistID, limit, offset int64) ([]PlaylistTrack, error) {
	rows, err := r.q.ListPlaylistTracks(ctx, dbsqlc.ListPlaylistTracksParams{
		PlaylistID: playlistID,
		Limit:      int32(limit),
		Offset:     int32(offset),
	})
	if err != nil {
		return nil, err
	}
	out := make([]PlaylistTrack, 0, len(rows))
	for _, row := range rows {
		out = append(out, trackFromDB(row))
	}
	return out, nil
}

func (r *sqlcRepository) CountTracks(ctx context.Context, playlistID int64) (int64, error) {
	return r.q.CountPlaylistTracks(ctx, playlistID)
}

func (r *sqlcRepository) ListTrackIDs(ctx context.Context, playlistID int64) ([]int64, error) {
	return r.q.ListPlaylistTrackIDs(ctx, playlistID)
}

func (r *sqlcRepository) ReorderTracks(ctx context.Context, playlistID int64, items []TrackPositionUpdate) error {
	tx, err := r.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	qtx := r.q.WithTx(tx)
	for _, item := range items {
		_, err := qtx.UpdatePlaylistTrackPosition(ctx, dbsqlc.UpdatePlaylistTrackPositionParams{
			PlaylistID: playlistID,
			TrackID:    item.TrackID,
			Position:   item.Position,
		})
		if err != nil {
			return mapNoRowsOrPgError(err)
		}
	}
	return tx.Commit(ctx)
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
		switch pgErr.Code {
		case "23505":
			return ErrRepoDuplicate
		case "23503":
			return ErrRepoNotFound
		case "23514":
			return ErrRepoInvalid
		}
	}
	return err
}
