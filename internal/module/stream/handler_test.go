package stream

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
	streams []Stream
}

func (f *fakeService) Start(_ context.Context, userID, stationID int64) (Stream, error) {
	row := Stream{
		ID:        1,
		UserID:    userID,
		StationID: stationID,
		StartedAt: time.Now(),
		CreatedAt: time.Now(),
	}
	return row, nil
}

func (f *fakeService) End(_ context.Context, id, _ int64) (Stream, error) {
	for _, s := range f.streams {
		if s.ID == id {
			ended := time.Now()
			s.EndedAt = &ended
			return s, nil
		}
	}
	return Stream{}, ErrNotFound
}

func (f *fakeService) GetByID(_ context.Context, id, _ int64) (Stream, error) {
	for _, s := range f.streams {
		if s.ID == id {
			return s, nil
		}
	}
	return Stream{}, ErrNotFound
}

func (f *fakeService) List(_ context.Context, _ int64, _ int64, _ int64) ([]Stream, int64, error) {
	out := make([]Stream, 0, len(f.streams))
	out = append(out, f.streams...)
	return out, int64(len(f.streams)), nil
}

func setupRouter(svc Service) *gin.Engine {
	gin.SetMode(gin.TestMode)
	h := NewHandler(svc)
	r := gin.New()
	r.Use(func(c *gin.Context) {
		c.Set(mw.ContextUserIDKey, int64(1))
		c.Next()
	})
	g := r.Group("/streams")
	g.POST("", httperr.Wrap(h.Start, MapError))
	g.GET("", httperr.Wrap(h.List, MapError))
	g.GET("/:id", httperr.Wrap(h.GetByID, MapError))
	g.PATCH("/:id/end", httperr.Wrap(h.End, MapError))
	return r
}

func TestHandler_List_Empty(t *testing.T) {
	r := setupRouter(&fakeService{})

	req := httptest.NewRequest(http.MethodGet, "/streams", http.NoBody)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, body = %s", w.Code, w.Body.String())
	}
	var envelope struct {
		Success bool  `json:"success"`
		Data    []any `json:"data"`
		Meta    struct {
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

	req := httptest.NewRequest(http.MethodGet, "/streams/999", http.NoBody)
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

func TestHandler_Start_OK(t *testing.T) {
	r := setupRouter(&fakeService{})

	body, _ := json.Marshal(StartStreamRequest{StationID: 10})
	req := httptest.NewRequest(http.MethodPost, "/streams", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Fatalf("status = %d, body = %s", w.Code, w.Body.String())
	}
	var envelope struct {
		Success bool           `json:"success"`
		Data    StreamResponse `json:"data"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &envelope); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if !envelope.Success {
		t.Fatalf("expected success envelope")
	}
	if envelope.Data.StationID != 10 {
		t.Fatalf("station_id = %d", envelope.Data.StationID)
	}
}

func TestHandler_End_NotFound(t *testing.T) {
	r := setupRouter(&fakeService{})

	req := httptest.NewRequest(http.MethodPatch, "/streams/999/end", http.NoBody)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Fatalf("status = %d, body = %s", w.Code, w.Body.String())
	}
}
