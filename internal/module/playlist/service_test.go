package playlist

import (
	"context"
	"errors"
	"testing"
	"time"
)

// fakeRepo is an in-memory stub implementing Repository for service tests.
type fakeRepo struct {
	playlists map[int64]Playlist
	tracks    map[int64][]PlaylistTrack
	nextID    int64
}

func (f *fakeRepo) Create(_ context.Context, in CreateInput) (Playlist, error) {
	f.nextID++
	row := Playlist{
		ID:          f.nextID,
		Name:        in.Name,
		Description: nil,
		IsPublic:    in.IsPublic,
		OwnerID:     in.OwnerID,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
	if in.Description != nil {
		row.Description = in.Description
	}
	if f.playlists == nil {
		f.playlists = make(map[int64]Playlist)
	}
	f.playlists[row.ID] = row
	return row, nil
}

func (f *fakeRepo) GetByID(_ context.Context, id int64) (Playlist, error) {
	if f.playlists == nil {
		return Playlist{}, ErrRepoNotFound
	}
	row, ok := f.playlists[id]
	if !ok {
		return Playlist{}, ErrRepoNotFound
	}
	return row, nil
}

func (f *fakeRepo) ListByOwner(_ context.Context, ownerID, _, _ int64) ([]Playlist, error) {
	var out []Playlist
	for _, p := range f.playlists {
		if p.OwnerID == ownerID {
			out = append(out, p)
		}
	}
	return out, nil
}

func (f *fakeRepo) CountByOwner(_ context.Context, ownerID int64) (int64, error) {
	var count int64
	for _, p := range f.playlists {
		if p.OwnerID == ownerID {
			count++
		}
	}
	return count, nil
}

func (f *fakeRepo) Update(_ context.Context, in UpdateInput) (Playlist, error) {
	row, ok := f.playlists[in.ID]
	if !ok {
		return Playlist{}, ErrRepoNotFound
	}
	row.Name = in.Name
	row.IsPublic = in.IsPublic
	row.Description = in.Description
	row.UpdatedAt = time.Now()
	f.playlists[in.ID] = row
	return row, nil
}

func (f *fakeRepo) Delete(_ context.Context, id int64) error {
	if _, ok := f.playlists[id]; !ok {
		return ErrRepoNotFound
	}
	delete(f.playlists, id)
	return nil
}

func (f *fakeRepo) AddTrack(_ context.Context, playlistID, trackID int64, position int32) error {
	if position < 0 {
		return ErrRepoInvalid
	}
	if f.tracks == nil {
		f.tracks = make(map[int64][]PlaylistTrack)
	}
	for _, item := range f.tracks[playlistID] {
		if item.ID == trackID {
			return ErrRepoDuplicate
		}
	}
	f.tracks[playlistID] = append(f.tracks[playlistID], PlaylistTrack{
		ID:       trackID,
		Position: position,
	})
	return nil
}

func (f *fakeRepo) RemoveTrack(_ context.Context, playlistID, trackID int64) error {
	rows := f.tracks[playlistID]
	for i, t := range rows {
		if t.ID == trackID {
			f.tracks[playlistID] = append(rows[:i], rows[i+1:]...)
			return nil
		}
	}
	return ErrRepoNotFound
}

func (f *fakeRepo) ListTracks(_ context.Context, playlistID, limit, offset int64) ([]PlaylistTrack, error) {
	rows := f.tracks[playlistID]
	if offset >= int64(len(rows)) {
		return []PlaylistTrack{}, nil
	}
	end := offset + limit
	if end > int64(len(rows)) {
		end = int64(len(rows))
	}
	return rows[offset:end], nil
}

func (f *fakeRepo) CountTracks(_ context.Context, playlistID int64) (int64, error) {
	return int64(len(f.tracks[playlistID])), nil
}

func (f *fakeRepo) ListTrackIDs(_ context.Context, playlistID int64) ([]int64, error) {
	rows := f.tracks[playlistID]
	out := make([]int64, 0, len(rows))
	for _, row := range rows {
		out = append(out, row.ID)
	}
	return out, nil
}

func (f *fakeRepo) ReorderTracks(_ context.Context, playlistID int64, items []TrackPositionUpdate) error {
	rows := f.tracks[playlistID]
	for _, item := range items {
		updated := false
		for i, t := range rows {
			if t.ID == item.TrackID {
				rows[i].Position = item.Position
				updated = true
				break
			}
		}
		if !updated {
			return ErrRepoNotFound
		}
	}
	f.tracks[playlistID] = rows
	return nil
}

func TestService_GetByID_Private_NotOwner(t *testing.T) {
	repo := &fakeRepo{playlists: map[int64]Playlist{
		1: {ID: 1, OwnerID: 1, IsPublic: false},
	}}
	svc := NewService(repo)
	_, err := svc.GetByID(context.Background(), 1, 2)
	if !errors.Is(err, ErrForbidden) {
		t.Fatalf("expected ErrForbidden, got %v", err)
	}
}

func TestService_GetByID_Public_NotOwner(t *testing.T) {
	repo := &fakeRepo{playlists: map[int64]Playlist{
		1: {ID: 1, OwnerID: 1, IsPublic: true},
	}}
	svc := NewService(repo)
	row, err := svc.GetByID(context.Background(), 1, 2)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if row.ID != 1 {
		t.Fatalf("id = %d", row.ID)
	}
}

func TestService_Update_MergesFields(t *testing.T) {
	repo := &fakeRepo{}
	svc := NewService(repo)

	created, err := svc.Create(context.Background(), 1, CreatePlaylistInput{
		Name:     "Original",
		IsPublic: false,
	})
	if err != nil {
		t.Fatalf("create: %v", err)
	}

	isPublic := true
	updated, err := svc.Update(context.Background(), created.ID, 1, UpdatePlaylistInput{IsPublic: &isPublic})
	if err != nil {
		t.Fatalf("update: %v", err)
	}
	if updated.Name != "Original" {
		t.Fatalf("name should be unchanged, got %q", updated.Name)
	}
	if !updated.IsPublic {
		t.Fatalf("expected is_public true")
	}
}

func TestService_ReorderTracks_NotFound(t *testing.T) {
	repo := &fakeRepo{playlists: map[int64]Playlist{
		1: {ID: 1, OwnerID: 1, IsPublic: false},
	}, tracks: map[int64][]PlaylistTrack{1: {{ID: 10, Position: 1}}}}
	svc := NewService(repo)

	err := svc.ReorderTracks(context.Background(), 1, 1, []TrackPositionPatch{
		{TrackID: 99, Position: 1},
	})
	if !errors.Is(err, ErrInvalidTrackOrder) {
		t.Fatalf("expected ErrInvalidTrackOrder, got %v", err)
	}
}

func TestService_AddTrack_Duplicate(t *testing.T) {
	repo := &fakeRepo{
		playlists: map[int64]Playlist{1: {ID: 1, OwnerID: 1}},
		tracks:    map[int64][]PlaylistTrack{1: {{ID: 3, Position: 0}}},
	}
	svc := NewService(repo)
	err := svc.AddTrack(context.Background(), 1, 1, 3, 1)
	if !errors.Is(err, ErrDuplicateTrack) {
		t.Fatalf("expected ErrDuplicateTrack, got %v", err)
	}
}

func TestService_ReorderTracks_DuplicatePosition(t *testing.T) {
	repo := &fakeRepo{
		playlists: map[int64]Playlist{1: {ID: 1, OwnerID: 1}},
		tracks:    map[int64][]PlaylistTrack{1: {{ID: 3, Position: 0}, {ID: 4, Position: 1}}},
	}
	svc := NewService(repo)
	err := svc.ReorderTracks(context.Background(), 1, 1, []TrackPositionPatch{
		{TrackID: 3, Position: 1},
		{TrackID: 4, Position: 1},
	})
	if !errors.Is(err, ErrInvalidTrackOrder) {
		t.Fatalf("expected ErrInvalidTrackOrder, got %v", err)
	}
}
