package radiobrowser

import (
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Module wires radio browser station HTTP routes.
type Module struct {
	h *Handler
}

// NewModule constructs a radio browser Module.
func NewModule(pool *pgxpool.Pool) *Module {
	repo := NewRepository(pool)
	svc := NewService(repo)
	return &Module{h: NewHandler(svc)}
}

// RegisterRoutes mounts /radio-browser/stations under api.
func (m *Module) RegisterRoutes(api *gin.RouterGroup) {
	g := api.Group("/radio-browser")
	{
		g.GET("/stations", m.h.List)
	}
}
