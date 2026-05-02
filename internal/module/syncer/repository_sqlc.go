package syncer

import (
	"context"
	"fmt"

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

func (r *sqlcRepository) UpsertBatch(ctx context.Context, stations []UpsertInput) (int64, error) {
	var upserted int64
	for _, s := range stations {
		err := r.q.UpsertRadioBrowserStation(ctx, dbsqlc.UpsertRadioBrowserStationParams{
			Stationuuid: s.StationUUID,
			Name:        s.Name,
			Url:         s.URL,
			UrlResolved: s.URLResolved,
			Homepage:    s.Homepage,
			Favicon:     s.Favicon,
			Country:     s.Country,
			Countrycode: s.CountryCode,
			State:       s.State,
			Language:    s.Language,
			Codec:       s.Codec,
			Bitrate:     s.Bitrate,
			Votes:       s.Votes,
			Tags:        s.Tags,
			LastCheckOk: s.LastCheckOK,
		})
		if err != nil {
			return upserted, fmt.Errorf("upsert station %s: %w", s.StationUUID, err)
		}
		upserted++
	}
	return upserted, nil
}
