package integration

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/tranphuocnhan/radio-shuffle/internal/module/playlist"
	"github.com/tranphuocnhan/radio-shuffle/pkg/dbsqlc"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type accessClaims struct {
	jwt.RegisteredClaims
	Role string `json:"role"`
}

type envelope[T any] struct {
	Success bool `json:"success"`
	Data    T    `json:"data"`
	Error   struct {
		Code    string `json:"code"`
		Message string `json:"message"`
	} `json:"error"`
	Meta struct {
		Page  int   `json:"page"`
		Limit int   `json:"limit"`
		Total int64 `json:"total"`
	} `json:"meta"`
}

func setupTestPool(t *testing.T) (*pgxpool.Pool, func()) {
	t.Helper()
	url := os.Getenv("DATABASE_URL")
	if url == "" {
		t.Skip("DATABASE_URL is required for integration tests")
	}

	adminCfg, err := pgxpool.ParseConfig(url)
	if err != nil {
		t.Fatalf("parse DATABASE_URL: %v", err)
	}
	adminPool, err := pgxpool.NewWithConfig(context.Background(), adminCfg)
	if err != nil {
		t.Fatalf("admin pool: %v", err)
	}

	schemaName := fmt.Sprintf("test_%d", time.Now().UnixNano())
	if _, err := adminPool.Exec(context.Background(), fmt.Sprintf("CREATE SCHEMA %s", schemaName)); err != nil {
		adminPool.Close()
		t.Fatalf("create schema: %v", err)
	}

	cfg, err := pgxpool.ParseConfig(url)
	if err != nil {
		adminPool.Close()
		t.Fatalf("parse DATABASE_URL: %v", err)
	}
	if cfg.ConnConfig.RuntimeParams == nil {
		cfg.ConnConfig.RuntimeParams = map[string]string{}
	}
	cfg.ConnConfig.RuntimeParams["search_path"] = schemaName
	pool, err := pgxpool.NewWithConfig(context.Background(), cfg)
	if err != nil {
		adminPool.Close()
		t.Fatalf("test pool: %v", err)
	}

	schemaSQL, err := os.ReadFile("db/schema.sql")
	if err != nil {
		pool.Close()
		adminPool.Close()
		t.Fatalf("read schema: %v", err)
	}
	if _, err := pool.Exec(context.Background(), string(schemaSQL)); err != nil {
		pool.Close()
		adminPool.Close()
		t.Fatalf("apply schema: %v", err)
	}

	cleanup := func() {
		pool.Close()
		_, _ = adminPool.Exec(context.Background(), fmt.Sprintf("DROP SCHEMA %s CASCADE", schemaName))
		adminPool.Close()
	}
	return pool, cleanup
}

func setupRouter(t *testing.T, pool *pgxpool.Pool, signingKey []byte, issuer string) *gin.Engine {
	t.Helper()
	gin.SetMode(gin.TestMode)
	r := gin.New()
	api := r.Group("/api/v1")
	playlist.NewModule(pool, signingKey, issuer).RegisterRoutes(api)
	return r
}

func seedUserStationTrack(t *testing.T, pool *pgxpool.Pool) (dbsqlc.Users, dbsqlc.Stations, dbsqlc.Tracks) {
	t.Helper()
	q := dbsqlc.New(pool)
	user, err := q.CreateUser(context.Background(), dbsqlc.CreateUserParams{
		Email:        fmt.Sprintf("user_%d@example.com", time.Now().UnixNano()),
		PasswordHash: []byte("hash"),
		Role:         "user",
	})
	if err != nil {
		t.Fatalf("create user: %v", err)
	}
	station, err := q.CreateStation(context.Background(), dbsqlc.CreateStationParams{
		Name:          "Station",
		Genre:         "",
		Description:   stringPtr(""),
		StreamUrl:     "https://example.com/stream",
		CoverImageUrl: stringPtr(""),
		IsPublic:      true,
		OwnerID:       int64Ptr(user.ID),
	})
	if err != nil {
		t.Fatalf("create station: %v", err)
	}
	track, err := q.CreateTrack(context.Background(), dbsqlc.CreateTrackParams{
		StationID:       station.ID,
		Title:           "Track 1",
		Artist:          "Artist",
		AudioUrl:        "https://example.com/audio.mp3",
		DurationSeconds: 180,
	})
	if err != nil {
		t.Fatalf("create track: %v", err)
	}
	return user, station, track
}

func authHeader(t *testing.T, userID int64, signingKey []byte, issuer string) string {
	t.Helper()
	claims := accessClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   fmt.Sprintf("%d", userID),
			Issuer:    issuer,
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(5 * time.Minute)),
		},
		Role: "user",
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	str, err := token.SignedString(signingKey)
	if err != nil {
		t.Fatalf("sign token: %v", err)
	}
	return "Bearer " + str
}

func stringPtr(s string) *string {
	return &s
}

func int64Ptr(v int64) *int64 {
	return &v
}

func TestIntegration_Playlists_CRUD_And_Tracks(t *testing.T) {
	pool, cleanup := setupTestPool(t)
	defer cleanup()

	signingKey := []byte("test-signing-key")
	issuer := "radio-shuffle"
	router := setupRouter(t, pool, signingKey, issuer)

	user, _, track := seedUserStationTrack(t, pool)
	authz := authHeader(t, user.ID, signingKey, issuer)

	// Create playlist.
	createBody, _ := json.Marshal(map[string]any{
		"name":        "Chill Mix",
		"description": "focus",
		"is_public":   true,
	})
	createReq := httptest.NewRequest(http.MethodPost, "/api/v1/playlists", bytes.NewReader(createBody))
	createReq.Header.Set("Authorization", authz)
	createReq.Header.Set("Content-Type", "application/json")
	createResp := httptest.NewRecorder()
	router.ServeHTTP(createResp, createReq)
	if createResp.Code != http.StatusCreated {
		t.Fatalf("create status = %d, body = %s", createResp.Code, createResp.Body.String())
	}
	var created envelope[playlist.PlaylistResponse]
	if err := json.Unmarshal(createResp.Body.Bytes(), &created); err != nil {
		t.Fatalf("decode create: %v", err)
	}
	if !created.Success {
		t.Fatalf("expected success on create")
	}
	playlistID := created.Data.ID
	if playlistID == 0 {
		t.Fatalf("expected non-zero playlist id")
	}

	// List playlists.
	listReq := httptest.NewRequest(http.MethodGet, "/api/v1/playlists?limit=10&offset=0", http.NoBody)
	listReq.Header.Set("Authorization", authz)
	listResp := httptest.NewRecorder()
	router.ServeHTTP(listResp, listReq)
	if listResp.Code != http.StatusOK {
		t.Fatalf("list status = %d, body = %s", listResp.Code, listResp.Body.String())
	}
	var listed envelope[[]playlist.PlaylistResponse]
	if err := json.Unmarshal(listResp.Body.Bytes(), &listed); err != nil {
		t.Fatalf("decode list: %v", err)
	}
	if listed.Meta.Total != 1 {
		t.Fatalf("expected total 1, got %d", listed.Meta.Total)
	}

	// Get playlist by ID.
	getReq := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/api/v1/playlists/%d", playlistID), http.NoBody)
	getReq.Header.Set("Authorization", authz)
	getResp := httptest.NewRecorder()
	router.ServeHTTP(getResp, getReq)
	if getResp.Code != http.StatusOK {
		t.Fatalf("get status = %d, body = %s", getResp.Code, getResp.Body.String())
	}
	var fetched envelope[playlist.PlaylistResponse]
	if err := json.Unmarshal(getResp.Body.Bytes(), &fetched); err != nil {
		t.Fatalf("decode get: %v", err)
	}
	if fetched.Data.ID != playlistID {
		t.Fatalf("expected playlist id %d, got %d", playlistID, fetched.Data.ID)
	}

	// Update playlist.
	updateBody, _ := json.Marshal(map[string]any{"name": "Updated Mix"})
	updateReq := httptest.NewRequest(http.MethodPatch, fmt.Sprintf("/api/v1/playlists/%d", playlistID), bytes.NewReader(updateBody))
	updateReq.Header.Set("Authorization", authz)
	updateReq.Header.Set("Content-Type", "application/json")
	updateResp := httptest.NewRecorder()
	router.ServeHTTP(updateResp, updateReq)
	if updateResp.Code != http.StatusOK {
		t.Fatalf("update status = %d, body = %s", updateResp.Code, updateResp.Body.String())
	}
	var updated envelope[playlist.PlaylistResponse]
	if err := json.Unmarshal(updateResp.Body.Bytes(), &updated); err != nil {
		t.Fatalf("decode update: %v", err)
	}
	if updated.Data.Name != "Updated Mix" {
		t.Fatalf("expected updated name, got %q", updated.Data.Name)
	}

	// Add track.
	addTrackBody, _ := json.Marshal(map[string]any{"track_id": track.ID, "position": 1})
	addTrackReq := httptest.NewRequest(http.MethodPost, fmt.Sprintf("/api/v1/playlists/%d/tracks", playlistID), bytes.NewReader(addTrackBody))
	addTrackReq.Header.Set("Authorization", authz)
	addTrackReq.Header.Set("Content-Type", "application/json")
	addTrackResp := httptest.NewRecorder()
	router.ServeHTTP(addTrackResp, addTrackReq)
	if addTrackResp.Code != http.StatusNoContent {
		t.Fatalf("add track status = %d, body = %s", addTrackResp.Code, addTrackResp.Body.String())
	}

	// List tracks.
	listTracksReq := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/api/v1/playlists/%d/tracks?limit=10&offset=0", playlistID), http.NoBody)
	listTracksReq.Header.Set("Authorization", authz)
	listTracksResp := httptest.NewRecorder()
	router.ServeHTTP(listTracksResp, listTracksReq)
	if listTracksResp.Code != http.StatusOK {
		t.Fatalf("list tracks status = %d, body = %s", listTracksResp.Code, listTracksResp.Body.String())
	}
	var tracks envelope[[]playlist.PlaylistTrackResponse]
	if err := json.Unmarshal(listTracksResp.Body.Bytes(), &tracks); err != nil {
		t.Fatalf("decode list tracks: %v", err)
	}
	if tracks.Meta.Total != 1 {
		t.Fatalf("expected track total 1, got %d", tracks.Meta.Total)
	}
	if tracks.Data[0].Position != 1 {
		t.Fatalf("expected track position 1, got %d", tracks.Data[0].Position)
	}

	// Reorder tracks.
	reorderBody, _ := json.Marshal(map[string]any{"items": []map[string]any{{"track_id": track.ID, "position": 2}}})
	reorderReq := httptest.NewRequest(http.MethodPatch, fmt.Sprintf("/api/v1/playlists/%d/tracks/reorder", playlistID), bytes.NewReader(reorderBody))
	reorderReq.Header.Set("Authorization", authz)
	reorderReq.Header.Set("Content-Type", "application/json")
	reorderResp := httptest.NewRecorder()
	router.ServeHTTP(reorderResp, reorderReq)
	if reorderResp.Code != http.StatusNoContent {
		t.Fatalf("reorder status = %d, body = %s", reorderResp.Code, reorderResp.Body.String())
	}

	// Confirm reorder.
	confirmReq := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/api/v1/playlists/%d/tracks?limit=10&offset=0", playlistID), http.NoBody)
	confirmReq.Header.Set("Authorization", authz)
	confirmResp := httptest.NewRecorder()
	router.ServeHTTP(confirmResp, confirmReq)
	if confirmResp.Code != http.StatusOK {
		t.Fatalf("confirm status = %d, body = %s", confirmResp.Code, confirmResp.Body.String())
	}
	var confirmed envelope[[]playlist.PlaylistTrackResponse]
	if err := json.Unmarshal(confirmResp.Body.Bytes(), &confirmed); err != nil {
		t.Fatalf("decode confirm: %v", err)
	}
	if confirmed.Data[0].Position != 2 {
		t.Fatalf("expected position 2, got %d", confirmed.Data[0].Position)
	}

	// Remove track.
	removeReq := httptest.NewRequest(http.MethodDelete, fmt.Sprintf("/api/v1/playlists/%d/tracks/%d", playlistID, track.ID), http.NoBody)
	removeReq.Header.Set("Authorization", authz)
	removeResp := httptest.NewRecorder()
	router.ServeHTTP(removeResp, removeReq)
	if removeResp.Code != http.StatusNoContent {
		t.Fatalf("remove status = %d, body = %s", removeResp.Code, removeResp.Body.String())
	}

	// Delete playlist.
	deleteReq := httptest.NewRequest(http.MethodDelete, fmt.Sprintf("/api/v1/playlists/%d", playlistID), http.NoBody)
	deleteReq.Header.Set("Authorization", authz)
	deleteResp := httptest.NewRecorder()
	router.ServeHTTP(deleteResp, deleteReq)
	if deleteResp.Code != http.StatusNoContent {
		t.Fatalf("delete status = %d, body = %s", deleteResp.Code, deleteResp.Body.String())
	}

	// Confirm delete.
	getDeletedReq := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/api/v1/playlists/%d", playlistID), http.NoBody)
	getDeletedReq.Header.Set("Authorization", authz)
	getDeletedResp := httptest.NewRecorder()
	router.ServeHTTP(getDeletedResp, getDeletedReq)
	if getDeletedResp.Code != http.StatusNotFound {
		t.Fatalf("get deleted status = %d, body = %s", getDeletedResp.Code, getDeletedResp.Body.String())
	}
}

