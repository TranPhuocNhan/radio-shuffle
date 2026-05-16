package radiobrowser

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

func stationFromDB(s dbsqlc.RadioBrowserStations) Station {
	return Station{
		ID:          s.ID,
		StationUUID: s.Stationuuid,
		Name:        s.Name,
		URL:         s.Url,
		URLResolved: s.UrlResolved,
		Homepage:    s.Homepage,
		Favicon:     s.Favicon,
		Country:     s.Country,
		CountryCode: s.Countrycode,
		State:       s.State,
		Language:    s.Language,
		Codec:       s.Codec,
		Bitrate:     s.Bitrate,
		Votes:       s.Votes,
		Tags:        s.Tags,
		LastCheckOK: s.LastCheckOk,
		SyncedAt:    s.SyncedAt.Time,
		CreatedAt:   s.CreatedAt.Time,
		UpdatedAt:   s.UpdatedAt.Time,
	}
}

func (r *sqlcRepository) List(ctx context.Context, name, country, language string, limit, offset int64) ([]Station, error) {
	rows, err := r.q.ListRadioBrowserStations(ctx, dbsqlc.ListRadioBrowserStationsParams{
		Name:       name,
		Country:    country,
		Language:   language,
		PageLimit:  int32(limit),
		PageOffset: int32(offset),
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

func (r *sqlcRepository) Count(ctx context.Context, name, country, language string) (int64, error) {
	return r.q.CountRadioBrowserStations(ctx, dbsqlc.CountRadioBrowserStationsParams{
		Name:     name,
		Country:  country,
		Language: language,
	})
}
