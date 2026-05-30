package track

import (
	"github.com/gin-gonic/gin"
	"github.com/tranphuocnhan/radio-shuffle/internal/platform/httperr"
)

// Module wires track HTTP routes.
type Module struct {
	h      *Handler
	authMw gin.HandlerFunc
}

// NewModule constructs a track Module.
func NewModule(h *Handler, authMw gin.HandlerFunc) *Module {
	return &Module{
		h:      h,
		authMw: authMw,
	}
}

// RegisterRoutes mounts /stations/:station_id/tracks under api.
// GET endpoints are public; write endpoints require a valid JWT.
func (m *Module) RegisterRoutes(api *gin.RouterGroup) {
	g := api.Group("/stations/:station_id/tracks")
	{
		// Public reads
		g.GET("", httperr.Wrap(m.h.List, MapError))
		g.GET("/:id", httperr.Wrap(m.h.GetByID, MapError))

		// Authenticated writes
		g.POST("", m.authMw, httperr.Wrap(m.h.Create, MapError))
		g.PATCH("/:id", m.authMw, httperr.Wrap(m.h.Update, MapError))
		g.DELETE("/:id", m.authMw, httperr.Wrap(m.h.Delete, MapError))
	}
}
