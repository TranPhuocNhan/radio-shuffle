package station

import (
	"context"
	"errors"
	"testing"

	"github.com/jackc/pgx/v5"
)

type serviceStubRepo struct {
	create                func(context.Context, CreateInput) (Station, error)
	getByID               func(context.Context, int64) (Station, error)
	list                  func(context.Context, int64, int64) ([]Station, error)
	count                 func(context.Context) (int64, error)
	update                func(context.Context, UpdateInput) (Station, error)
	delete                func(context.Context, int64) error
	follow                func(context.Context, int64, int64) error
	unfollow              func(context.Context, int64, int64) error
	isFollowing           func(context.Context, int64, int64) (bool, error)
	getFollowedStations   func(context.Context, int64, int64, int64) ([]Station, error)
	countFollowedStations func(context.Context, int64) (int64, error)
	countFollowers        func(context.Context, int64) (int64, error)
}

func (s serviceStubRepo) Create(ctx context.Context, in CreateInput) (Station, error) {
	if s.create != nil {
		return s.create(ctx, in)
	}
	return Station{}, errors.New("unexpected Create")
}

func (s serviceStubRepo) GetByID(ctx context.Context, id int64) (Station, error) {
	if s.getByID != nil {
		return s.getByID(ctx, id)
	}
	return Station{}, errors.New("unexpected GetByID")
}

func (s serviceStubRepo) List(ctx context.Context, limit int64, offset int64) ([]Station, error) {
	if s.list != nil {
		return s.list(ctx, limit, offset)
	}
	return nil, errors.New("unexpected List")
}

func (s serviceStubRepo) Count(ctx context.Context) (int64, error) {
	if s.count != nil {
		return s.count(ctx)
	}
	return 0, errors.New("unexpected Count")
}

func (s serviceStubRepo) Update(ctx context.Context, in UpdateInput) (Station, error) {
	if s.update != nil {
		return s.update(ctx, in)
	}
	return Station{}, errors.New("unexpected Update")
}

func (s serviceStubRepo) Delete(ctx context.Context, id int64) error {
	if s.delete != nil {
		return s.delete(ctx, id)
	}
	return errors.New("unexpected Delete")
}

func (s serviceStubRepo) Follow(ctx context.Context, userID int64, stationID int64) error {
	if s.follow != nil {
		return s.follow(ctx, userID, stationID)
	}
	return errors.New("unexpected Follow")
}

func (s serviceStubRepo) Unfollow(ctx context.Context, userID int64, stationID int64) error {
	if s.unfollow != nil {
		return s.unfollow(ctx, userID, stationID)
	}
	return errors.New("unexpected Unfollow")
}

func (s serviceStubRepo) IsFollowing(ctx context.Context, userID int64, stationID int64) (bool, error) {
	if s.isFollowing != nil {
		return s.isFollowing(ctx, userID, stationID)
	}
	return false, errors.New("unexpected IsFollowing")
}

func (s serviceStubRepo) GetFollowedStations(ctx context.Context, userID int64, limit int64, offset int64) ([]Station, error) {
	if s.getFollowedStations != nil {
		return s.getFollowedStations(ctx, userID, limit, offset)
	}
	return nil, errors.New("unexpected GetFollowedStations")
}

func (s serviceStubRepo) CountFollowedStations(ctx context.Context, userID int64) (int64, error) {
	if s.countFollowedStations != nil {
		return s.countFollowedStations(ctx, userID)
	}
	return 0, errors.New("unexpected CountFollowedStations")
}

func (s serviceStubRepo) CountFollowers(ctx context.Context, stationID int64) (int64, error) {
	if s.countFollowers != nil {
		return s.countFollowers(ctx, stationID)
	}
	return 0, errors.New("unexpected CountFollowers")
}

func TestService_Create_DefaultsIsPublic(t *testing.T) {
	repo := serviceStubRepo{
		create: func(_ context.Context, in CreateInput) (Station, error) {
			if !in.IsPublic {
				return Station{}, errors.New("expected IsPublic true")
			}
			return Station{ID: 1, Name: in.Name, IsPublic: in.IsPublic}, nil
		},
	}
	svc := NewService(repo)

	got, err := svc.Create(context.Background(), CreateStationInput{
		Name:      "Test",
		StreamUrl: "http://stream.test",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.ID != 1 || !got.IsPublic {
		t.Fatalf("unexpected station: %+v", got)
	}
}

func TestService_GetByID_NotFound(t *testing.T) {
	repo := serviceStubRepo{
		getByID: func(context.Context, int64) (Station, error) { return Station{}, pgx.ErrNoRows },
	}
	svc := NewService(repo)

	_, err := svc.GetByID(context.Background(), 10)
	if !errors.Is(err, ErrNotFound) {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
}

func TestService_List_ReturnsRowsAndTotal(t *testing.T) {
	repo := serviceStubRepo{
		count: func(context.Context) (int64, error) { return 2, nil },
		list: func(context.Context, int64, int64) ([]Station, error) {
			return []Station{{ID: 1}, {ID: 2}}, nil
		},
	}
	svc := NewService(repo)

	rows, total, err := svc.List(context.Background(), ListStationsInput{Limit: 10, Offset: 0})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if total != 2 || len(rows) != 2 {
		t.Fatalf("unexpected result: total=%d rows=%d", total, len(rows))
	}
}

func TestService_Update_MergesFields(t *testing.T) {
	repo := serviceStubRepo{
		getByID: func(context.Context, int64) (Station, error) {
			return Station{ID: 9, Name: "Old", Genre: "rock", StreamUrl: "http://old", IsPublic: true}, nil
		},
		update: func(_ context.Context, in UpdateInput) (Station, error) {
			if in.Name != "New" || in.Genre != "rock" || in.StreamUrl != "http://old" {
				return Station{}, errors.New("merge mismatch")
			}
			return Station{ID: in.ID, Name: in.Name}, nil
		},
	}
	svc := NewService(repo)
	name := "New"
	got, err := svc.Update(context.Background(), UpdateStationInput{ID: 9, Name: &name})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.ID != 9 || got.Name != "New" {
		t.Fatalf("unexpected station: %+v", got)
	}
}

func TestService_Update_NotFound(t *testing.T) {
	repo := serviceStubRepo{
		getByID: func(context.Context, int64) (Station, error) { return Station{}, pgx.ErrNoRows },
	}
	svc := NewService(repo)

	_, err := svc.Update(context.Background(), UpdateStationInput{ID: 10})
	if !errors.Is(err, ErrNotFound) {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
}

func TestService_Delete_NotFound(t *testing.T) {
	repo := serviceStubRepo{
		getByID: func(context.Context, int64) (Station, error) { return Station{}, pgx.ErrNoRows },
	}
	svc := NewService(repo)

	err := svc.Delete(context.Background(), 10)
	if !errors.Is(err, ErrNotFound) {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
}

func TestService_Delete_OK(t *testing.T) {
	deleted := false
	repo := serviceStubRepo{
		getByID: func(context.Context, int64) (Station, error) { return Station{ID: 10}, nil },
		delete: func(context.Context, int64) error { deleted = true; return nil },
	}
	svc := NewService(repo)

	if err := svc.Delete(context.Background(), 10); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !deleted {
		t.Fatalf("expected delete to be called")
	}
}

func TestService_Follow_NotFound(t *testing.T) {
	repo := serviceStubRepo{
		getByID: func(context.Context, int64) (Station, error) { return Station{}, pgx.ErrNoRows },
	}
	svc := NewService(repo)

	err := svc.Follow(context.Background(), 5, 10)
	if !errors.Is(err, ErrNotFound) {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
}

func TestService_Follow_OK(t *testing.T) {
	called := false
	repo := serviceStubRepo{
		getByID: func(context.Context, int64) (Station, error) { return Station{ID: 10}, nil },
		follow: func(context.Context, int64, int64) error { called = true; return nil },
	}
	svc := NewService(repo)

	if err := svc.Follow(context.Background(), 5, 10); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !called {
		t.Fatalf("expected follow to be called")
	}
}

func TestService_Unfollow_NotFound(t *testing.T) {
	repo := serviceStubRepo{
		getByID: func(context.Context, int64) (Station, error) { return Station{}, pgx.ErrNoRows },
	}
	svc := NewService(repo)

	err := svc.Unfollow(context.Background(), 5, 10)
	if !errors.Is(err, ErrNotFound) {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
}

func TestService_Unfollow_OK(t *testing.T) {
	called := false
	repo := serviceStubRepo{
		getByID: func(context.Context, int64) (Station, error) { return Station{ID: 10}, nil },
		unfollow: func(context.Context, int64, int64) error { called = true; return nil },
	}
	svc := NewService(repo)

	if err := svc.Unfollow(context.Background(), 5, 10); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !called {
		t.Fatalf("expected unfollow to be called")
	}
}

func TestService_IsFollowing(t *testing.T) {
	repo := serviceStubRepo{
		getByID: func(context.Context, int64) (Station, error) { return Station{ID: 10}, nil },
		isFollowing: func(context.Context, int64, int64) (bool, error) { return true, nil },
	}
	svc := NewService(repo)

	following, err := svc.IsFollowing(context.Background(), 5, 10)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !following {
		t.Fatalf("expected following true")
	}
}

func TestService_GetFollowedStations(t *testing.T) {
	repo := serviceStubRepo{
		getFollowedStations: func(context.Context, int64, int64, int64) ([]Station, error) {
			return []Station{{ID: 1}}, nil
		},
	}
	svc := NewService(repo)

	rows, err := svc.GetFollowedStations(context.Background(), 5, 10, 0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(rows) != 1 {
		t.Fatalf("expected 1 station, got %d", len(rows))
	}
}

func TestService_CountFollowedStations(t *testing.T) {
	repo := serviceStubRepo{
		countFollowedStations: func(context.Context, int64) (int64, error) { return 3, nil },
	}
	svc := NewService(repo)

	count, err := svc.CountFollowedStations(context.Background(), 5)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if count != 3 {
		t.Fatalf("expected 3, got %d", count)
	}
}

func TestService_CountFollowers(t *testing.T) {
	repo := stubRepo{
		getByID: func(context.Context, int64) (Station, error) { return Station{ID: 10}, nil },
		countFollowers: func(context.Context, int64) (int64, error) { return 4, nil },
	}
	svc := NewService(repo)

	count, err := svc.CountFollowers(context.Background(), 10)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if count != 4 {
		t.Fatalf("expected 4, got %d", count)
	}
}


