package track

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
)

// fakeService is an in-memory stub implementing Service for handler tests.
type fakeService struct {
	tracks []Track
}

func (f *fakeService) Create(_ context.Context, in CreateTrackInput) (Track, error) {
	row := Track{
		ID:              1,
		StationID:       in.StationID,
		Title:           in.Title,
		Artist:          in.Artist,
		AudioUrl:        in.AudioUrl,
		DurationSeconds: in.DurationSeconds,
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
	}
	return row, nil
}

func (f *fakeService) GetByID(_ context.Context, id int64) (Track, error) {
	for _, t := range f.tracks {
		if t.ID == id {
			return t, nil
		}
	}
	return Track{}, ErrNotFound
}

func (f *fakeService) List(_ context.Context, _ int64, _ int64, _ int64) ([]Track, int64, error) {
	out := make([]Track, 0, len(f.tracks))
	for _, t := range f.tracks {
		out = append(out, t)
	}
	return out, int64(len(f.tracks)), nil
}

func (f *fakeService) Update(_ context.Context, id, _ int64, _ UpdateTrackRequest) (Track, error) {
	for _, t := range f.tracks {
		if t.ID == id {
			return t, nil
		}
	}
	return Track{}, ErrNotFound
}

func (f *fakeService) Delete(_ context.Context, id int64) error {
	for _, t := range f.tracks {
		if t.ID == id {
			return nil
		}
	}
	return ErrNotFound
}

func setupRouter(svc Service) *gin.Engine {
	gin.SetMode(gin.TestMode)
	h := NewHandler(svc)
	r := gin.New()
	g := r.Group("/stations/:station_id/tracks")
	g.GET("", h.List)
	g.GET("/:id", h.GetByID)
	g.POST("", h.Create)
	g.PATCH("/:id", h.Update)
	g.DELETE("/:id", h.Delete)
	return r
}

func TestHandler_List_Empty(t *testing.T) {
	r := setupRouter(&fakeService{})

	req := httptest.NewRequest(http.MethodGet, "/stations/1/tracks", http.NoBody)
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
	if envelope.Data == nil {
		t.Fatalf("expected data array, got nil")
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

	req := httptest.NewRequest(http.MethodGet, "/stations/1/tracks/999", http.NoBody)
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

	body, _ := json.Marshal(CreateTrackRequest{
		Title:           "Test Track",
		Artist:          "Test Artist",
		AudioUrl:        "https://example.com/audio.mp3",
		DurationSeconds: 180,
	})
	req := httptest.NewRequest(http.MethodPost, "/stations/1/tracks", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Fatalf("status = %d, body = %s", w.Code, w.Body.String())
	}
	var envelope struct {
		Success bool          `json:"success"`
		Data    TrackResponse `json:"data"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &envelope); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if !envelope.Success {
		t.Fatalf("expected success envelope")
	}
	if envelope.Data.Title != "Test Track" {
		t.Fatalf("title = %q", envelope.Data.Title)
	}
}

func TestHandler_Delete_NotFound(t *testing.T) {
	r := setupRouter(&fakeService{})

	req := httptest.NewRequest(http.MethodDelete, "/stations/1/tracks/999", http.NoBody)
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
	if envelope.Error.Code != "NOT_FOUND" {
		t.Fatalf("code = %q", envelope.Error.Code)
	}
}
