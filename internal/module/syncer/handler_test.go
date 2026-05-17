package syncer

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/tranphuocnhan/radio-shuffle/internal/platform/mw"
)

type stubJobService struct {
	trigger func(context.Context, string) (SyncJob, error)
	status  func(context.Context, string) (SyncJob, error)
}

func (s stubJobService) Trigger(ctx context.Context, requestedBy string) (SyncJob, error) {
	if s.trigger == nil {
		return SyncJob{}, errors.New("unexpected Trigger")
	}
	return s.trigger(ctx, requestedBy)
}

func (s stubJobService) Status(ctx context.Context, requestID string) (SyncJob, error) {
	if s.status == nil {
		return SyncJob{}, errors.New("unexpected Status")
	}
	return s.status(ctx, requestID)
}

func withUser(id int64) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Set(mw.ContextUserIDKey, id)
		c.Set(mw.ContextUserRoleKey, "admin")
		c.Next()
	}
}

func TestHandler_Trigger_OK(t *testing.T) {
	gin.SetMode(gin.TestMode)
	svc := stubJobService{
		trigger: func(context.Context, string) (SyncJob, error) {
			return SyncJob{RequestID: "req-1"}, nil
		},
	}
	h := NewHandler(svc)
	r := gin.New()
	r.Use(withUser(42))
	r.POST("/syncer/trigger", h.Trigger)

	req := httptest.NewRequest(http.MethodPost, "/syncer/trigger", http.NoBody)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusAccepted {
		t.Fatalf("status = %d, body = %s", w.Code, w.Body.String())
	}
	var envelope struct {
		Success bool `json:"success"`
		Data    struct {
			RequestID string `json:"request_id"`
			StatusURL string `json:"status_url"`
		} `json:"data"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &envelope); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if !envelope.Success {
		t.Fatalf("expected success envelope")
	}
	if envelope.Data.RequestID != "req-1" {
		t.Fatalf("request_id = %q", envelope.Data.RequestID)
	}
}

func TestHandler_Status_OK(t *testing.T) {
	gin.SetMode(gin.TestMode)
	now := time.Now()
	svc := stubJobService{
		status: func(context.Context, string) (SyncJob, error) {
			return SyncJob{RequestID: "req-1", Status: SyncJobCompleted, RequestedBy: "admin:1", Scope: defaultScopeBytes, CreatedAt: now}, nil
		},
	}
	h := NewHandler(svc)
	r := gin.New()
	r.Use(withUser(42))
	r.GET("/syncer/status/:request_id", h.Status)

	req := httptest.NewRequest(http.MethodGet, "/syncer/status/req-1", http.NoBody)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, body = %s", w.Code, w.Body.String())
	}
	var envelope struct {
		Success bool `json:"success"`
		Data    struct {
			RequestID string `json:"request_id"`
			Status    string `json:"status"`
		} `json:"data"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &envelope); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if !envelope.Success || envelope.Data.Status != "completed" {
		t.Fatalf("unexpected response: %+v", envelope)
	}
}

func TestHandler_Status_NotFound(t *testing.T) {
	gin.SetMode(gin.TestMode)
	svc := stubJobService{
		status: func(context.Context, string) (SyncJob, error) { return SyncJob{}, ErrJobNotFound },
	}
	h := NewHandler(svc)
	r := gin.New()
	r.Use(withUser(42))
	r.GET("/syncer/status/:request_id", h.Status)

	req := httptest.NewRequest(http.MethodGet, "/syncer/status/missing", http.NoBody)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Fatalf("status = %d, body = %s", w.Code, w.Body.String())
	}
}

