package auth

import (
	"net/http"

	"github.com/tranphuocnhan/radio-shuffle/internal/platform/httperr"
	"github.com/tranphuocnhan/radio-shuffle/internal/platform/response"

	"github.com/gin-gonic/gin"
)

// Handler exposes auth HTTP endpoints.
type Handler struct {
	svc Service
}

// NewHandler constructs a Handler.
func NewHandler(svc Service) *Handler {
	return &Handler{svc: svc}
}

func (h *Handler) Register(c *gin.Context) error {
	var req RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		return httperr.BadRequest(err.Error())
	}
	out, err := h.svc.Register(c.Request.Context(), RegisterInput{Email: req.Email, Password: req.Password})
	if err != nil {
		return err
	}
	response.OK(c, http.StatusCreated, AuthResponse{AccessToken: out.AccessToken, RefreshToken: out.RefreshToken})
	return nil
}

func (h *Handler) Login(c *gin.Context) error {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		return httperr.BadRequest(err.Error())
	}
	out, err := h.svc.Login(c.Request.Context(), LoginInput{Email: req.Email, Password: req.Password})
	if err != nil {
		return err
	}
	response.OK(c, http.StatusOK, AuthResponse{AccessToken: out.AccessToken, RefreshToken: out.RefreshToken})
	return nil
}

func (h *Handler) Refresh(c *gin.Context) error {
	var req RefreshRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		return httperr.BadRequest(err.Error())
	}
	out, err := h.svc.Refresh(c.Request.Context(), RefreshInput{RefreshToken: req.RefreshToken})
	if err != nil {
		return err
	}
	response.OK(c, http.StatusOK, AuthResponse{AccessToken: out.AccessToken, RefreshToken: out.RefreshToken})
	return nil
}

func (h *Handler) Logout(c *gin.Context) error {
	var req LogoutRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		return httperr.BadRequest(err.Error())
	}
	if err := h.svc.Logout(c.Request.Context(), LogoutInput{RefreshToken: req.RefreshToken}); err != nil {
		return err
	}
	c.Status(http.StatusNoContent)
	return nil
}
