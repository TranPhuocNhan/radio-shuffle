package playlist

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/tranphuocnhan/radio-shuffle/internal/platform/httperr"
	"github.com/tranphuocnhan/radio-shuffle/internal/platform/mw"

	"github.com/gin-gonic/gin"
)

// fakeService is an in-memory stub implementing Service for handler tests.
type fakeService struct {
	playlists []Playlist
	tracks    []PlaylistTrack
}

func (f *fakeService) Create(_ context.Context, ownerID int64, in CreatePlaylistInput) (Playlist, error) {
	row := Playlist{
		ID:          1,
		Name:        in.Name,
		Description: nil,
		IsPublic:    in.IsPublic,
		OwnerID:     ownerID,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
	if in.Description != nil {
		row.Description = in.Description
	}
	return row, nil
}

func (f *fakeService) GetByID(_ context.Context, id, _ int64) (Playlist, error) {
	for _, p := range f.playlists {
		if p.ID == id {
			return p, nil
		}
	}
	return Playlist{}, ErrNotFound
}

func (f *fakeService) List(_ context.Context, _ int64, _, _ int64) ([]Playlist, int64, error) {
	out := make([]Playlist, 0, len(f.playlists))
	out = append(out, f.playlists...)
	return out, int64(len(f.playlists)), nil
}

func (f *fakeService) Update(_ context.Context, id, _ int64, _ UpdatePlaylistInput) (Playlist, error) {
	for _, p := range f.playlists {
		if p.ID == id {
			return p, nil
		}
	}
	return Playlist{}, ErrNotFound
}

func (f *fakeService) Delete(_ context.Context, id, _ int64) error {
	for _, p := range f.playlists {
		if p.ID == id {
			return nil
		}
	}
	return ErrNotFound
}

func (f *fakeService) AddTrack(_ context.Context, _, _, _ int64, _ int32) error {
	return nil
}

func (f *fakeService) RemoveTrack(_ context.Context, _, _, _ int64) error {
	return ErrTrackNotFound
}

func (f *fakeService) ListTracks(_ context.Context, _, _ int64, _, _ int64) ([]PlaylistTrack, int64, error) {
	out := make([]PlaylistTrack, 0, len(f.tracks))
	out = append(out, f.tracks...)
	return out, int64(len(f.tracks)), nil
}

func (f *fakeService) ReorderTracks(_ context.Context, _, _ int64, _ []TrackPositionPatch) error {
	return nil
}

func setupRouter(svc Service) *gin.Engine {
	gin.SetMode(gin.TestMode)
	h := NewHandler(svc)
	r := gin.New()
	r.Use(func(c *gin.Context) {
		c.Set(mw.ContextUserIDKey, int64(1))
		c.Next()
	})
	g := r.Group("/playlists")
	g.POST("", httperr.Wrap(h.create, MapError))
	g.GET("", httperr.Wrap(h.list, MapError))
	g.GET("/:id", httperr.Wrap(h.getByID, MapError))
	g.PATCH("/:id", httperr.Wrap(h.update, MapError))
	g.DELETE("/:id", httperr.Wrap(h.delete, MapError))
	g.POST("/:id/tracks", httperr.Wrap(h.addTrack, MapError))
	g.GET("/:id/tracks", httperr.Wrap(h.listTracks, MapError))
	g.DELETE("/:id/tracks/:track_id", httperr.Wrap(h.removeTrack, MapError))
	g.PATCH("/:id/tracks/reorder", httperr.Wrap(h.reorderTracks, MapError))
	return r
}

func TestHandler_List_Empty(t *testing.T) {
	r := setupRouter(&fakeService{})

	req := httptest.NewRequest(http.MethodGet, "/playlists", http.NoBody)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, body = %s", w.Code, w.Body.String())
	}
	var envelope struct {
		Success bool  `json:"success"`
		Data    []any `json:"data"`
		Meta    struct {
			Page  int   `json:"page"`
			Limit int   `json:"limit"`
			Total int64 `json:"total"`
		} `json:"meta"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &envelope); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if !envelope.Success {
		t.Fatalf("expected success envelope")
	}
	if len(envelope.Data) != 0 {
		t.Fatalf("expected empty list, got %d items", len(envelope.Data))
	}
	if envelope.Meta.Total != 0 {
		t.Fatalf("expected meta.total 0, got %d", envelope.Meta.Total)
	}
}

func TestHandler_GetByID_NotFound(t *testing.T) {
	r := setupRouter(&fakeService{})

	req := httptest.NewRequest(http.MethodGet, "/playlists/999", http.NoBody)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Fatalf("status = %d, body = %s", w.Code, w.Body.String())
	}
	var envelope struct {
		Success bool `json:"success"`
		Error   struct {
			Code string `json:"code"`
		} `json:"error"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &envelope); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if envelope.Success {
		t.Fatalf("expected failing envelope")
	}
	if envelope.Error.Code != "NOT_FOUND" {
		t.Fatalf("code = %q", envelope.Error.Code)
	}
}

func TestHandler_Create_OK(t *testing.T) {
	r := setupRouter(&fakeService{})

	body, _ := json.Marshal(CreatePlaylistRequest{
		Name:     "Chill Mix",
		IsPublic: true,
	})
	req := httptest.NewRequest(http.MethodPost, "/playlists", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Fatalf("status = %d, body = %s", w.Code, w.Body.String())
	}
	var envelope struct {
		Success bool             `json:"success"`
		Data    PlaylistResponse `json:"data"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &envelope); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if !envelope.Success {
		t.Fatalf("expected success envelope")
	}
	if envelope.Data.Name != "Chill Mix" {
		t.Fatalf("name = %q", envelope.Data.Name)
	}
}
