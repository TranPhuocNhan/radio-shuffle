package station

import (
	"github.com/gin-gonic/gin"
	"github.com/tranphuocnhan/radio-shuffle/internal/platform/httperr"
)

// Module wires station HTTP routes.
type Module struct {
	h      *Handler
	authMw gin.HandlerFunc
}

// NewModule constructs a station Module.
func NewModule(h *Handler, authMw gin.HandlerFunc) *Module {
	return &Module{
		h:      h,
		authMw: authMw,
	}
}

// RegisterRoutes mounts /stations under api.
func (m *Module) RegisterRoutes(api *gin.RouterGroup) {
	g := api.Group("/stations")
	{
		g.POST("", httperr.Wrap(m.h.Create, MapError))
		g.GET("", httperr.Wrap(m.h.List, MapError))
		g.GET("/:station_id", httperr.Wrap(m.h.GetByID, MapError))
		g.PATCH("/:station_id", httperr.Wrap(m.h.Update, MapError))
		g.DELETE("/:station_id", httperr.Wrap(m.h.Delete, MapError))

		g.POST("/:station_id/follow", m.authMw, httperr.Wrap(m.h.Follow, MapError))
		g.DELETE("/:station_id/follow", m.authMw, httperr.Wrap(m.h.Unfollow, MapError))
		g.GET("/:station_id/following", m.authMw, httperr.Wrap(m.h.IsFollowing, MapError))
		g.GET("/:station_id/followers/count", m.authMw, httperr.Wrap(m.h.CountFollowers, MapError))
		g.GET("/followed", m.authMw, httperr.Wrap(m.h.ListFollowed, MapError))
	}
}
