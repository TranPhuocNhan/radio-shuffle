package auth

import (
	"github.com/gin-gonic/gin"
	"github.com/tranphuocnhan/radio-shuffle/internal/platform/config"
)

// Module wires auth HTTP routes.
type Module struct {
	h *Handler
}

// NewModule constructs an auth Module.
func NewModule(cfg *config.Config, users UserReader, writers UserWriter, tokens TokenRepository, verifier EmailVerifier) *Module {
	svc := NewService(users, writers, tokens, Config{
		SigningKey: []byte(cfg.JWTSigningKey),
		Issuer:     cfg.JWTIssuer,
		AccessTTL:  cfg.AccessTTL(),
		RefreshTTL: cfg.RefreshTTL(),
	}, verifier)
	return &Module{h: NewHandler(svc)}
}

// RegisterRoutes mounts /auth under api.
func (m *Module) RegisterRoutes(api *gin.RouterGroup) {
	g := api.Group("/auth")
	{
		g.POST("/register", m.h.Register)
		g.POST("/login", m.h.Login)
		g.POST("/refresh", m.h.Refresh)
		g.POST("/logout", m.h.Logout)
	}
}
