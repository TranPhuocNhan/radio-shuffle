package syncer

import (
	"context"
	"errors"
	"testing"

	"github.com/tranphuocnhan/radio-shuffle/internal/platform/radiobrowser"
)

// fakeRepo is an in-memory Repository stub.
type fakeRepo struct {
	upserted []UpsertInput
	err      error
}

func (f *fakeRepo) UpsertBatch(_ context.Context, stations []UpsertInput) (int64, error) {
	if f.err != nil {
		return 0, f.err
	}
	f.upserted = append(f.upserted, stations...)
	return int64(len(stations)), nil
}

// fakeClient replaces the radiobrowser.Client in tests.
type fakeClient struct {
	pages [][]radiobrowser.Station
	calls int
}

func (f *fakeClient) FetchStations(_ context.Context, params radiobrowser.FetchParams) ([]radiobrowser.Station, error) {
	if f.calls >= len(f.pages) {
		return nil, nil
	}
	page := f.pages[f.calls]
	f.calls++
	return page, nil
}

// clientFetcher is a minimal interface so tests can inject a fake client.
type clientFetcher interface {
	FetchStations(ctx context.Context, params radiobrowser.FetchParams) ([]radiobrowser.Station, error)
}

// testService is a thin variant of service that accepts a configurable page size
// and a clientFetcher, so pagination can be exercised without 1000-station fixtures.
type testService struct {
	repo     Repository
	client   clientFetcher
	pageSize int
}

func (s *testService) Sync(ctx context.Context) (SyncResult, error) {
	ps := s.pageSize
	if ps <= 0 {
		ps = fetchPageSize
	}
	var result SyncResult
	offset := 0

	for {
		stations, err := s.client.FetchStations(ctx, radiobrowser.FetchParams{
			Limit:  ps,
			Offset: offset,
		})
		if err != nil {
			return result, err
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
			return result, err
		}
		result.Fetched += len(stations)
		result.Upserted += upserted
		offset += len(stations)

		if len(stations) < ps {
			break
		}
	}
	return result, nil
}

func TestService_Sync_TwoPages(t *testing.T) {
	// pageSize=3 so the first full page triggers a second fetch,
	// and the second page (2 items) signals end of results.
	page1 := makeStations(3, 0)
	page2 := makeStations(2, 3)

	repo := &fakeRepo{}
	svc := &testService{
		repo:     repo,
		client:   &fakeClient{pages: [][]radiobrowser.Station{page1, page2}},
		pageSize: 3,
	}

	result, err := svc.Sync(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Fetched != 5 {
		t.Errorf("fetched: got %d, want 5", result.Fetched)
	}
	if result.Upserted != 5 {
		t.Errorf("upserted: got %d, want 5", result.Upserted)
	}
	if len(repo.upserted) != 5 {
		t.Errorf("repo received %d stations, want 5", len(repo.upserted))
	}
}

func TestService_Sync_EmptyResponse(t *testing.T) {
	repo := &fakeRepo{}
	svc := &testService{
		repo:     repo,
		client:   &fakeClient{pages: [][]radiobrowser.Station{}},
		pageSize: 3,
	}

	result, err := svc.Sync(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Fetched != 0 || result.Upserted != 0 {
		t.Errorf("expected zero result, got %+v", result)
	}
}

func TestService_Sync_RepoError(t *testing.T) {
	repo := &fakeRepo{err: errors.New("db error")}
	svc := &testService{
		repo:     repo,
		client:   &fakeClient{pages: [][]radiobrowser.Station{makeStations(2, 0)}},
		pageSize: 3,
	}

	_, err := svc.Sync(context.Background())
	if err == nil {
		t.Fatal("expected error from repo, got nil")
	}
}

func TestToUpsertInput_EmptyStringsBecomesNil(t *testing.T) {
	st := radiobrowser.Station{
		UUID:        "uuid-1",
		Name:        "Radio X",
		URL:         "http://stream.test",
		URLResolved: "",
		Country:     "",
		LastCheckOK: true,
	}
	got := toUpsertInput(st)
	if got.URLResolved != nil {
		t.Errorf("URLResolved: expected nil, got %v", got.URLResolved)
	}
	if got.Country != nil {
		t.Errorf("Country: expected nil, got %v", got.Country)
	}
	if !got.LastCheckOK {
		t.Errorf("LastCheckOK: expected true, got false")
	}
}

func makeStations(n, startID int) []radiobrowser.Station {
	stations := make([]radiobrowser.Station, n)
	for i := range stations {
		stations[i] = radiobrowser.Station{
			UUID:    "uuid-" + string(rune('a'+startID+i)),
			Name:    "Station",
			URL:     "http://stream.test",
			Bitrate: 128,
		}
	}
	return stations
}
