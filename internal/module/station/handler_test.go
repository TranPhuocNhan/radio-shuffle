package station

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5"
	"github.com/tranphuocnhan/radio-shuffle/internal/platform/mw"
)

// stubRepo is a configurable in-memory substitute for Repository in handler tests.
type stubRepo struct {
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

func (s stubRepo) Create(ctx context.Context, in CreateInput) (Station, error) {
	if s.create != nil {
		return s.create(ctx, in)
	}
	return Station{}, errors.New("unexpected Create")
}

func (s stubRepo) GetByID(ctx context.Context, id int64) (Station, error) {
	if s.getByID != nil {
		return s.getByID(ctx, id)
	}
	return Station{}, pgx.ErrNoRows
}

func (s stubRepo) List(ctx context.Context, limit int64, offset int64) ([]Station, error) {
	if s.list != nil {
		return s.list(ctx, limit, offset)
	}
	return []Station{}, nil
}

func (s stubRepo) Count(ctx context.Context) (int64, error) {
	if s.count != nil {
		return s.count(ctx)
	}
	return 0, nil
}

func (s stubRepo) Update(ctx context.Context, in UpdateInput) (Station, error) {
	if s.update != nil {
		return s.update(ctx, in)
	}
	return Station{}, errors.New("unexpected Update")
}

func (s stubRepo) Delete(ctx context.Context, id int64) error {
	if s.delete != nil {
		return s.delete(ctx, id)
	}
	return errors.New("unexpected Delete")
}

func (s stubRepo) Follow(ctx context.Context, userID int64, stationID int64) error {
	if s.follow != nil {
		return s.follow(ctx, userID, stationID)
	}
	return errors.New("unexpected Follow")
}

func (s stubRepo) Unfollow(ctx context.Context, userID int64, stationID int64) error {
	if s.unfollow != nil {
		return s.unfollow(ctx, userID, stationID)
	}
	return errors.New("unexpected Unfollow")
}

func (s stubRepo) IsFollowing(ctx context.Context, userID int64, stationID int64) (bool, error) {
	if s.isFollowing != nil {
		return s.isFollowing(ctx, userID, stationID)
	}
	return false, errors.New("unexpected IsFollowing")
}

func (s stubRepo) GetFollowedStations(ctx context.Context, userID int64, limit int64, offset int64) ([]Station, error) {
	if s.getFollowedStations != nil {
		return s.getFollowedStations(ctx, userID, limit, offset)
	}
	return []Station{}, nil
}

func (s stubRepo) CountFollowedStations(ctx context.Context, userID int64) (int64, error) {
	if s.countFollowedStations != nil {
		return s.countFollowedStations(ctx, userID)
	}
	return 0, nil
}

func (s stubRepo) CountFollowers(ctx context.Context, stationID int64) (int64, error) {
	if s.countFollowers != nil {
		return s.countFollowers(ctx, stationID)
	}
	return 0, errors.New("unexpected CountFollowers")
}

func withUser(id int64) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Set(mw.ContextUserIDKey, id)
		c.Next()
	}
}

func TestHandler_List_Empty(t *testing.T) {
	gin.SetMode(gin.TestMode)
	repo := stubRepo{}
	h := NewHandler(NewService(repo))
	r := gin.New()
	r.GET("/stations", h.List)

	req := httptest.NewRequest(http.MethodGet, "/stations", http.NoBody)
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
	gin.SetMode(gin.TestMode)
	repo := stubRepo{}
	h := NewHandler(NewService(repo))
	r := gin.New()
	r.GET("/stations/:station_id", h.GetByID)

	req := httptest.NewRequest(http.MethodGet, "/stations/999", http.NoBody)
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

func TestHandler_Follow_OK(t *testing.T) {
	gin.SetMode(gin.TestMode)
	repo := stubRepo{
		getByID: func(context.Context, int64) (Station, error) {
			return Station{ID: 10}, nil
		},
		follow: func(context.Context, int64, int64) error { return nil },
	}
	h := NewHandler(NewService(repo))
	r := gin.New()
	r.Use(withUser(5))
	r.POST("/stations/:station_id/follow", h.Follow)

	req := httptest.NewRequest(http.MethodPost, "/stations/10/follow", http.NoBody)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusNoContent {
		t.Fatalf("status = %d, body = %s", w.Code, w.Body.String())
	}
}

func TestHandler_IsFollowing_OK(t *testing.T) {
	gin.SetMode(gin.TestMode)
	repo := stubRepo{
		getByID: func(context.Context, int64) (Station, error) {
			return Station{ID: 10}, nil
		},
		isFollowing: func(context.Context, int64, int64) (bool, error) { return true, nil },
	}
	h := NewHandler(NewService(repo))
	r := gin.New()
	r.Use(withUser(5))
	r.GET("/stations/:station_id/following", h.IsFollowing)

	req := httptest.NewRequest(http.MethodGet, "/stations/10/following", http.NoBody)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, body = %s", w.Code, w.Body.String())
	}
	var envelope struct {
		Success bool `json:"success"`
		Data    struct {
			Following bool `json:"following"`
		} `json:"data"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &envelope); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if !envelope.Success || !envelope.Data.Following {
		t.Fatalf("expected following true")
	}
}

func TestHandler_ListFollowed_Empty(t *testing.T) {
	gin.SetMode(gin.TestMode)
	repo := stubRepo{
		countFollowedStations: func(context.Context, int64) (int64, error) { return 0, nil },
		getFollowedStations: func(context.Context, int64, int64, int64) ([]Station, error) {
			return []Station{}, nil
		},
	}
	h := NewHandler(NewService(repo))
	r := gin.New()
	r.Use(withUser(5))
	r.GET("/stations/followed", h.ListFollowed)

	req := httptest.NewRequest(http.MethodGet, "/stations/followed", http.NoBody)
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

func TestHandler_CountFollowers_OK(t *testing.T) {
	gin.SetMode(gin.TestMode)
	repo := stubRepo{
		getByID: func(context.Context, int64) (Station, error) {
			return Station{ID: 10}, nil
		},
		countFollowers: func(context.Context, int64) (int64, error) { return 2, nil },
	}
	h := NewHandler(NewService(repo))
	r := gin.New()
	r.Use(withUser(5))
	r.GET("/stations/:station_id/followers/count", h.CountFollowers)

	req := httptest.NewRequest(http.MethodGet, "/stations/10/followers/count", http.NoBody)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, body = %s", w.Code, w.Body.String())
	}
	var envelope struct {
		Success bool `json:"success"`
		Data    struct {
			Count int64 `json:"count"`
		} `json:"data"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &envelope); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if !envelope.Success || envelope.Data.Count != 2 {
		t.Fatalf("expected count 2, got %d", envelope.Data.Count)
	}
}

