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
type fakeRepo struct{}

func (fakeRepo) List(context.Context, int64, int64) ([]Station, error) {
	return []Station{}, nil
}

func (fakeRepo) Count(context.Context) (int64, error) {
	return 0, nil
}

func TestHandler_List_Empty(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewHandler(NewService(fakeRepo{}))
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
