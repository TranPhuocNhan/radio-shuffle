package auth

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
	"github.com/tranphuocnhan/radio-shuffle/internal/platform/httperr"
)

type stubService struct {
	registerOut AuthOutput
	registerErr error
	loginOut    AuthOutput
	loginErr    error
	refreshOut  AuthOutput
	refreshErr  error
	logoutErr   error
}

func (s *stubService) Register(_ context.Context, _ RegisterInput) (AuthOutput, error) {
	return s.registerOut, s.registerErr
}

func (s *stubService) Login(_ context.Context, _ LoginInput) (AuthOutput, error) {
	return s.loginOut, s.loginErr
}

func (s *stubService) Refresh(_ context.Context, _ RefreshInput) (AuthOutput, error) {
	return s.refreshOut, s.refreshErr
}

func (s *stubService) Logout(_ context.Context, _ LogoutInput) error {
	return s.logoutErr
}

func TestRegisterHandler(t *testing.T) {
	gin.SetMode(gin.TestMode)
	stub := &stubService{registerOut: AuthOutput{AccessToken: "access", RefreshToken: "refresh"}}
	h := NewHandler(stub)
	router := gin.New()
	router.POST("/auth/register", httperr.Wrap(h.Register, MapError))

	payload := `{"email":"test@example.com","password":"password123"}`
	req := httptest.NewRequest(http.MethodPost, "/auth/register", bytes.NewBufferString(payload))
	req.Header.Set("Content-Type", "application/json")
	res := httptest.NewRecorder()

	router.ServeHTTP(res, req)
	require.Equal(t, http.StatusCreated, res.Code)

	var body map[string]any
	require.NoError(t, json.Unmarshal(res.Body.Bytes(), &body))
	require.Equal(t, true, body["success"])
}

func TestLoginHandlerInvalidCredentials(t *testing.T) {
	gin.SetMode(gin.TestMode)
	stub := &stubService{loginErr: ErrInvalidCredentials}
	h := NewHandler(stub)
	router := gin.New()
	router.POST("/auth/login", httperr.Wrap(h.Login, MapError))

	payload := `{"email":"test@example.com","password":"wrong"}`
	req := httptest.NewRequest(http.MethodPost, "/auth/login", bytes.NewBufferString(payload))
	req.Header.Set("Content-Type", "application/json")
	res := httptest.NewRecorder()

	router.ServeHTTP(res, req)
	require.Equal(t, http.StatusUnauthorized, res.Code)
}

func TestRefreshHandler(t *testing.T) {
	gin.SetMode(gin.TestMode)
	stub := &stubService{refreshOut: AuthOutput{AccessToken: "access", RefreshToken: "refresh"}}
	h := NewHandler(stub)
	router := gin.New()
	router.POST("/auth/refresh", httperr.Wrap(h.Refresh, MapError))

	payload := `{"refresh_token":"token"}`
	req := httptest.NewRequest(http.MethodPost, "/auth/refresh", bytes.NewBufferString(payload))
	req.Header.Set("Content-Type", "application/json")
	res := httptest.NewRecorder()

	router.ServeHTTP(res, req)
	require.Equal(t, http.StatusOK, res.Code)
}

func TestLogoutHandler(t *testing.T) {
	gin.SetMode(gin.TestMode)
	stub := &stubService{}
	h := NewHandler(stub)
	router := gin.New()
	router.POST("/auth/logout", httperr.Wrap(h.Logout, MapError))

	payload := `{"refresh_token":"token"}`
	req := httptest.NewRequest(http.MethodPost, "/auth/logout", bytes.NewBufferString(payload))
	req.Header.Set("Content-Type", "application/json")
	res := httptest.NewRecorder()

	router.ServeHTTP(res, req)
	require.Equal(t, http.StatusNoContent, res.Code)
}
