package radiobrowser

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

// fakeRepo is a minimal in-memory substitute for Repository in handler tests.
type fakeRepo struct {
	lastName     string
	lastCountry  string
	lastLanguage string
	lastLimit    int64
	lastOffset   int64
}

func (f *fakeRepo) List(_ context.Context, name, country, language string, limit, offset int64) ([]Station, error) {
	f.lastName = name
	f.lastCountry = country
	f.lastLanguage = language
	f.lastLimit = limit
	f.lastOffset = offset
	return []Station{}, nil
}

func (f *fakeRepo) Count(_ context.Context, name, country, language string) (int64, error) {
	f.lastName = name
	f.lastCountry = country
	f.lastLanguage = language
	return 0, nil
}

func TestHandler_List_Empty(t *testing.T) {
	gin.SetMode(gin.TestMode)
	repo := &fakeRepo{}
	h := NewHandler(NewService(repo))
	r := gin.New()
	r.GET("/radio-browser/stations", h.List)

	req := httptest.NewRequest(http.MethodGet, "/radio-browser/stations", http.NoBody)
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

func TestHandler_List_Filters(t *testing.T) {
	gin.SetMode(gin.TestMode)
	repo := &fakeRepo{}
	h := NewHandler(NewService(repo))
	r := gin.New()
	r.GET("/radio-browser/stations", h.List)

	req := httptest.NewRequest(http.MethodGet, "/radio-browser/stations?name=Jazz&country=France&language=French", http.NoBody)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, body = %s", w.Code, w.Body.String())
	}
	if repo.lastName != "Jazz" {
		t.Fatalf("expected name filter 'Jazz', got %q", repo.lastName)
	}
	if repo.lastCountry != "France" {
		t.Fatalf("expected country filter 'France', got %q", repo.lastCountry)
	}
	if repo.lastLanguage != "French" {
		t.Fatalf("expected language filter 'French', got %q", repo.lastLanguage)
	}
}
