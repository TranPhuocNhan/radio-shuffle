package radiobrowser

import (
	"github.com/gin-gonic/gin"
	"github.com/tranphuocnhan/radio-shuffle/internal/platform/httperr"
)

// Module wires radio browser station HTTP routes.
type Module struct {
	h *Handler
}

// NewModule constructs a radio browser Module.
func NewModule(h *Handler) *Module {
	return &Module{h: h}
}

// RegisterRoutes mounts /radio-browser/stations under api.
func (m *Module) RegisterRoutes(api *gin.RouterGroup) {
	g := api.Group("/radio-browser")
	{
		g.GET("/stations", httperr.Wrap(m.h.List, MapError))
	}
}
