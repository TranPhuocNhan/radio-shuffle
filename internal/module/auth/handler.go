package auth

import (
	"errors"
	"net/http"

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

func (h *Handler) Register(c *gin.Context) {
	var req RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	out, err := h.svc.Register(c.Request.Context(), RegisterInput{Email: req.Email, Password: req.Password})
	if err != nil {
		switch {
		case errors.Is(err, ErrEmailTaken):
			response.Conflict(c, err.Error())
		case errors.Is(err, ErrWeakPassword):
			response.BadRequest(c, err.Error())
		default:
			response.Internal(c, err)
		}
		return
	}
	response.OK(c, http.StatusCreated, AuthResponse{AccessToken: out.AccessToken, RefreshToken: out.RefreshToken})
}

func (h *Handler) Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	out, err := h.svc.Login(c.Request.Context(), LoginInput{Email: req.Email, Password: req.Password})
	if err != nil {
		if errors.Is(err, ErrInvalidCredentials) {
			response.Unauthorized(c, err.Error())
			return
		}
		response.Internal(c, err)
		return
	}
	response.OK(c, http.StatusOK, AuthResponse{AccessToken: out.AccessToken, RefreshToken: out.RefreshToken})
}

func (h *Handler) Refresh(c *gin.Context) {
	var req RefreshRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	out, err := h.svc.Refresh(c.Request.Context(), RefreshInput{RefreshToken: req.RefreshToken})
	if err != nil {
		if errors.Is(err, ErrInvalidRefreshToken) {
			response.Unauthorized(c, err.Error())
			return
		}
		response.Internal(c, err)
		return
	}
	response.OK(c, http.StatusOK, AuthResponse{AccessToken: out.AccessToken, RefreshToken: out.RefreshToken})
}

func (h *Handler) Logout(c *gin.Context) {
	var req LogoutRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	if err := h.svc.Logout(c.Request.Context(), LogoutInput{RefreshToken: req.RefreshToken}); err != nil {
		response.Internal(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}
