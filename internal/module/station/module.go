package station

import (
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Module wires station HTTP routes.
type Module struct {
	h *Handler
}

// NewModule constructs a station Module.
func NewModule(pool *pgxpool.Pool) *Module {
	repo := NewRepository(pool)
	svc := NewService(repo)
	return &Module{h: NewHandler(svc)}
}

// RegisterRoutes mounts /stations under api.
func (m *Module) RegisterRoutes(api *gin.RouterGroup) {
	g := api.Group("/stations")
	{
		g.POST("", m.h.Create)
		g.GET("", m.h.List)
		g.GET("/:station_id", m.h.GetByID)
		g.PATCH("/:station_id", m.h.Update)
		g.DELETE("/:station_id", m.h.Delete)
	}
}
