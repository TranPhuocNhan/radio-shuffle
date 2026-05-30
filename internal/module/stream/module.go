package stream

import (
	"github.com/gin-gonic/gin"
	"github.com/tranphuocnhan/radio-shuffle/internal/platform/httperr"
)

// Module wires stream HTTP routes.
type Module struct {
	h      *Handler
	authMw gin.HandlerFunc
}

// NewModule constructs a stream Module.
func NewModule(h *Handler, authMw gin.HandlerFunc) *Module {
	return &Module{
		h:      h,
		authMw: authMw,
	}
}

// RegisterRoutes mounts /streams under api. All routes require auth.
func (m *Module) RegisterRoutes(api *gin.RouterGroup) {
	g := api.Group("/streams", m.authMw)
	{
		g.POST("", httperr.Wrap(m.h.Start, MapError))
		g.GET("", httperr.Wrap(m.h.List, MapError))
		g.GET("/:id", httperr.Wrap(m.h.GetByID, MapError))
		g.PATCH("/:id/end", httperr.Wrap(m.h.End, MapError))
	}
}
