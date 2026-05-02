package syncer

import (
	"context"
	"fmt"

	"github.com/tranphuocnhan/radio-shuffle/internal/platform/radiobrowser"
)

const fetchPageSize = 1000

// SyncResult summarises the outcome of one sync run.
type SyncResult struct {
	Fetched  int
	Upserted int64
}

// Service defines the sync use case.
type Service interface {
	Sync(ctx context.Context) (SyncResult, error)
}

type service struct {
	repo   Repository
	client *radiobrowser.Client
}

func newService(repo Repository, client *radiobrowser.Client) Service {
	return &service{repo: repo, client: client}
}

func (s *service) Sync(ctx context.Context) (SyncResult, error) {
	var result SyncResult
	offset := 0

	for {
		stations, err := s.client.FetchStations(ctx, radiobrowser.FetchParams{
			Limit:  fetchPageSize,
			Offset: offset,
		})
		if err != nil {
			return result, fmt.Errorf("fetch stations at offset %d: %w", offset, err)
		}
		if len(stations) == 0 {
			break
		}

		inputs := make([]UpsertInput, len(stations))
		for i, st := range stations {
			inputs[i] = toUpsertInput(st)
		}

		upserted, err := s.repo.UpsertBatch(ctx, inputs)
		if err != nil {
			return result, fmt.Errorf("upsert batch at offset %d: %w", offset, err)
		}

		result.Fetched += len(stations)
		result.Upserted += upserted
		offset += len(stations)

		if len(stations) < fetchPageSize {
			break
		}
	}

	return result, nil
}

func toUpsertInput(st radiobrowser.Station) UpsertInput {
	return UpsertInput{
		StationUUID: st.UUID,
		Name:        st.Name,
		URL:         st.URL,
		URLResolved: nullableStr(st.URLResolved),
		Homepage:    nullableStr(st.Homepage),
		Favicon:     nullableStr(st.Favicon),
		Country:     nullableStr(st.Country),
		CountryCode: nullableStr(st.CountryCode),
		State:       nullableStr(st.State),
		Language:    nullableStr(st.Language),
		Codec:       nullableStr(st.Codec),
		Bitrate:     int32(st.Bitrate),
		Votes:       int32(st.Votes),
		Tags:        nullableStr(st.Tags),
		LastCheckOK: st.LastCheckOK,
	}
}

// nullableStr converts an empty string to nil so the DB stores NULL.
func nullableStr(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}
