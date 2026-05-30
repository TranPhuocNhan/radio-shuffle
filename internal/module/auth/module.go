package auth

import (
	"github.com/gin-gonic/gin"
	"github.com/tranphuocnhan/radio-shuffle/internal/platform/httperr"
)

// Module wires auth HTTP routes.
type Module struct {
	h *Handler
}

// NewModule constructs an auth Module.
func NewModule(h *Handler) *Module {
	return &Module{h: h}
}

// RegisterRoutes mounts /auth under api.
func (m *Module) RegisterRoutes(api *gin.RouterGroup) {
	g := api.Group("/auth")
	{
		g.POST("/register", httperr.Wrap(m.h.Register, MapError))
		g.POST("/login", httperr.Wrap(m.h.Login, MapError))
		g.POST("/refresh", httperr.Wrap(m.h.Refresh, MapError))
		g.POST("/logout", httperr.Wrap(m.h.Logout, MapError))
	}
}
