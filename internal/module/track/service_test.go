package track

import (
	"context"
	"errors"
	"testing"
	"time"
)

// fakeRepo is an in-memory stub implementing Repository for service tests.
type fakeRepo struct {
	tracks []Track
	nextID int64
}

func (f *fakeRepo) Create(_ context.Context, in CreateInput) (Track, error) {
	f.nextID++
	row := Track{
		ID:              f.nextID,
		StationID:       in.StationID,
		Title:           in.Title,
		Artist:          in.Artist,
		AudioUrl:        in.AudioUrl,
		DurationSeconds: in.DurationSeconds,
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
	}
	f.tracks = append(f.tracks, row)
	return row, nil
}

func (f *fakeRepo) GetByID(_ context.Context, id int64) (Track, error) {
	for _, t := range f.tracks {
		if t.ID == id {
			return t, nil
		}
	}
	return Track{}, ErrRepoNotFound
}

func (f *fakeRepo) ListByStation(_ context.Context, stationID, _, _ int64) ([]Track, error) {
	var out []Track
	for _, t := range f.tracks {
		if t.StationID == stationID {
			out = append(out, t)
		}
	}
	return out, nil
}

func (f *fakeRepo) CountByStation(_ context.Context, stationID int64) (int64, error) {
	var count int64
	for _, t := range f.tracks {
		if t.StationID == stationID {
			count++
		}
	}
	return count, nil
}

func (f *fakeRepo) Update(_ context.Context, in UpdateInput) (Track, error) {
	for i, t := range f.tracks {
		if t.ID == in.ID {
			f.tracks[i].Title = in.Title
			f.tracks[i].Artist = in.Artist
			f.tracks[i].AudioUrl = in.AudioUrl
			f.tracks[i].DurationSeconds = in.DurationSeconds
			f.tracks[i].UpdatedAt = time.Now()
			return f.tracks[i], nil
		}
	}
	return Track{}, ErrRepoNotFound
}

func (f *fakeRepo) Delete(_ context.Context, id int64) error {
	for i, t := range f.tracks {
		if t.ID == id {
			f.tracks = append(f.tracks[:i], f.tracks[i+1:]...)
			return nil
		}
	}
	return ErrRepoNotFound
}

func TestService_GetByID_NotFound(t *testing.T) {
	svc := NewService(&fakeRepo{})
	_, err := svc.GetByID(context.Background(), 99)
	if !errors.Is(err, ErrNotFound) {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
}

func TestService_Create_OK(t *testing.T) {
	svc := NewService(&fakeRepo{})
	row, err := svc.Create(context.Background(), CreateTrackInput{
		StationID:       1,
		Title:           "My Track",
		Artist:          "Some Artist",
		AudioUrl:        "https://example.com/track.mp3",
		DurationSeconds: 240,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if row.ID == 0 {
		t.Fatal("expected non-zero ID")
	}
	if row.Title != "My Track" {
		t.Fatalf("title = %q", row.Title)
	}
	if row.StationID != 1 {
		t.Fatalf("station_id = %d", row.StationID)
	}
}

func TestService_Update_NotFound(t *testing.T) {
	svc := NewService(&fakeRepo{})
	title := "New Title"
	_, err := svc.Update(context.Background(), 99, 1, UpdateTrackInput{Title: &title})
	if !errors.Is(err, ErrNotFound) {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
}

func TestService_Update_MergesFields(t *testing.T) {
	repo := &fakeRepo{}
	svc := NewService(repo)

	// Seed a track.
	created, err := svc.Create(context.Background(), CreateTrackInput{
		StationID:       1,
		Title:           "Original Title",
		Artist:          "Original Artist",
		AudioUrl:        "https://example.com/original.mp3",
		DurationSeconds: 120,
	})
	if err != nil {
		t.Fatalf("create: %v", err)
	}

	// Patch only the title; other fields must remain unchanged.
	newTitle := "Updated Title"
	updated, err := svc.Update(context.Background(), created.ID, 1, UpdateTrackInput{
		Title: &newTitle,
	})
	if err != nil {
		t.Fatalf("update: %v", err)
	}
	if updated.Title != "Updated Title" {
		t.Fatalf("title = %q", updated.Title)
	}
	if updated.Artist != "Original Artist" {
		t.Fatalf("artist should be unchanged, got %q", updated.Artist)
	}
	if updated.DurationSeconds != 120 {
		t.Fatalf("duration_seconds should be unchanged, got %d", updated.DurationSeconds)
	}
}
