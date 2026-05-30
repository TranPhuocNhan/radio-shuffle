package playlist

import (
	"github.com/gin-gonic/gin"
	"github.com/tranphuocnhan/radio-shuffle/internal/platform/httperr"
)

// Module wires playlist HTTP routes.
type Module struct {
	h      *Handler
	authMw gin.HandlerFunc
}

// NewModule constructs a playlist Module.
func NewModule(h *Handler, authMw gin.HandlerFunc) *Module {
	return &Module{
		h:      h,
		authMw: authMw,
	}
}

// RegisterRoutes mounts /playlists under api.
func (m *Module) RegisterRoutes(api *gin.RouterGroup) {
	g := api.Group("/playlists", m.authMw)
	{
		g.POST("", httperr.Wrap(m.h.create, MapError))
		g.GET("", httperr.Wrap(m.h.list, MapError))
		g.GET("/:id", httperr.Wrap(m.h.getByID, MapError))
		g.PATCH("/:id", httperr.Wrap(m.h.update, MapError))
		g.DELETE("/:id", httperr.Wrap(m.h.delete, MapError))

		g.POST("/:id/tracks", httperr.Wrap(m.h.addTrack, MapError))
		g.GET("/:id/tracks", httperr.Wrap(m.h.listTracks, MapError))
		g.DELETE("/:id/tracks/:track_id", httperr.Wrap(m.h.removeTrack, MapError))
		g.PATCH("/:id/tracks/reorder", httperr.Wrap(m.h.reorderTracks, MapError))
	}
}
