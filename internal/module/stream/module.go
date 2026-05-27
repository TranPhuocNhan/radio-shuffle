package stream

import (
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	plmw "github.com/tranphuocnhan/radio-shuffle/internal/platform/mw"
)

// Module wires stream HTTP routes.
type Module struct {
	h      *Handler
	authMw gin.HandlerFunc
}

// NewModule constructs a stream Module.
func NewModule(pool *pgxpool.Pool, signingKey []byte, issuer string) *Module {
	repo := NewRepository(pool)
	svc := NewService(repo)
	return &Module{
		h:      NewHandler(svc),
		authMw: plmw.AuthRequired(signingKey, issuer),
	}
}

// RegisterRoutes mounts /streams under api. All routes require auth.
func (m *Module) RegisterRoutes(api *gin.RouterGroup) {
	g := api.Group("/streams", m.authMw)
	{
		g.POST("", m.h.Start)
		g.GET("", m.h.List)
		g.GET("/:id", m.h.GetByID)
		g.PATCH("/:id/end", m.h.End)
	}
}
