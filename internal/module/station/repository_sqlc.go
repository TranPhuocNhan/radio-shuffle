package station

import (
	"context"

	"github.com/tranphuocnhan/radio-shuffle/pkg/dbsqlc"

	"github.com/jackc/pgx/v5/pgxpool"
)

type sqlcRepository struct {
	q *dbsqlc.Queries
}

// NewRepository returns a SQLC-backed Repository.
func NewRepository(pool *pgxpool.Pool) Repository {
	return &sqlcRepository{q: dbsqlc.New(pool)}
}

func stationFromDB(s dbsqlc.Stations) Station {
	return Station{
		ID:            s.ID,
		Name:          s.Name,
		Genre:         s.Genre,
		Description:   s.Description,
		StreamUrl:     s.StreamUrl,
		CoverImageUrl: s.CoverImageUrl,
		IsPublic:      s.IsPublic,
		OwnerID:       s.OwnerID,
		CreatedAt:     s.CreatedAt.Time,
		UpdatedAt:     s.UpdatedAt.Time,
	}
}

func (r *sqlcRepository) Create(ctx context.Context, in CreateInput) (Station, error) {
	s, err := r.q.CreateStation(ctx, dbsqlc.CreateStationParams{
		Name:          in.Name,
		Genre:         in.Genre,
		Description:   in.Description,
		StreamUrl:     in.StreamUrl,
		CoverImageUrl: in.CoverImageUrl,
		IsPublic:      in.IsPublic,
		OwnerID:       in.OwnerID,
	})
	if err != nil {
		return Station{}, err
	}
	return stationFromDB(s), nil
}

func (r *sqlcRepository) GetByID(ctx context.Context, id int64) (Station, error) {
	s, err := r.q.GetStationByID(ctx, id)
	if err != nil {
		return Station{}, err
	}
	return stationFromDB(s), nil
}

func (r *sqlcRepository) List(ctx context.Context, limit, offset int64) ([]Station, error) {
	rows, err := r.q.ListStations(ctx, dbsqlc.ListStationsParams{
		Limit:  int32(limit),
		Offset: int32(offset),
	})
	if err != nil {
		return nil, err
	}
	out := make([]Station, 0, len(rows))
	for _, s := range rows {
		out = append(out, stationFromDB(s))
	}
	return out, nil
}

func (r *sqlcRepository) Count(ctx context.Context) (int64, error) {
	return r.q.CountStations(ctx)
}

func (r *sqlcRepository) Update(ctx context.Context, in UpdateInput) (Station, error) {
	s, err := r.q.UpdateStation(ctx, dbsqlc.UpdateStationParams{
		ID:            in.ID,
		Name:          in.Name,
		Genre:         in.Genre,
		Description:   in.Description,
		StreamUrl:     in.StreamUrl,
		CoverImageUrl: in.CoverImageUrl,
		IsPublic:      in.IsPublic,
		OwnerID:       in.OwnerID,
	})
	if err != nil {
		return Station{}, err
	}
	return stationFromDB(s), nil
}

func (r *sqlcRepository) Delete(ctx context.Context, id int64) error {
	return r.q.DeleteStation(ctx, id)
}

func (r *sqlcRepository) Follow(ctx context.Context, userID, stationID int64) error {
	return r.q.FollowStation(ctx, dbsqlc.FollowStationParams{
		UserID:    userID,
		StationID: stationID,
	})
}
func (r *sqlcRepository) Unfollow(ctx context.Context, userID, stationID int64) error {
	return r.q.UnfollowStation(ctx, dbsqlc.UnfollowStationParams{
		UserID:    userID,
		StationID: stationID,
	})
}
func (r *sqlcRepository) IsFollowing(ctx context.Context, userID, stationID int64) (bool, error) {
	follow, err := r.q.IsFollowingStation(ctx, dbsqlc.IsFollowingStationParams{
		UserID:    userID,
		StationID: stationID,
	})
	if err != nil {
		return false, err
	}
	return follow, nil
}
func (r *sqlcRepository) GetFollowedStations(ctx context.Context, userID int64, limit int64, offset int64) ([]Station, error) {
	rows, err := r.q.GetStationsByUserID(ctx, dbsqlc.GetStationsByUserIDParams{
		UserID: userID,
		Limit:  int32(limit),
		Offset: int32(offset),
	})
	if err != nil {
		return nil, err
	}
	out := make([]Station, 0, len(rows))
	for _, row := range rows {
		out = append(out, stationFromDB(row))
	}
	return out, nil
}
func (r *sqlcRepository) CountFollowedStations(ctx context.Context, userID int64) (int64, error) {
	count, err := r.q.CountStationsByUserID(ctx, userID)
	if err != nil {
		return 0, err
	}
	return int64(count), nil
}
func (r *sqlcRepository) CountFollowers(ctx context.Context, stationID int64) (int64, error) {
	count, err := r.q.CountFollowersByStationID(ctx, stationID)
	if err != nil {
		return 0, err
	}
	return int64(count), nil
}
