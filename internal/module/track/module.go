package track

import (
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	plmw "github.com/tranphuocnhan/radio-shuffle/internal/platform/mw"
)

// Module wires track HTTP routes.
type Module struct {
	h          *Handler
	authMw     gin.HandlerFunc
}

// NewModule constructs a track Module.
// signingKey and issuer are required to build the AuthRequired middleware for
// write endpoints (POST, PATCH, DELETE).
func NewModule(pool *pgxpool.Pool, signingKey []byte, issuer string) *Module {
	repo := NewRepository(pool)
	svc := NewService(repo)
	return &Module{
		h:      NewHandler(svc),
		authMw: plmw.AuthRequired(signingKey, issuer),
	}
}

// RegisterRoutes mounts /stations/:station_id/tracks under api.
// GET endpoints are public; write endpoints require a valid JWT.
func (m *Module) RegisterRoutes(api *gin.RouterGroup) {
	g := api.Group("/stations/:station_id/tracks")
	{
		// Public reads
		g.GET("", m.h.List)
		g.GET("/:id", m.h.GetByID)

		// Authenticated writes
		g.POST("", m.authMw, m.h.Create)
		g.PATCH("/:id", m.authMw, m.h.Update)
		g.DELETE("/:id", m.authMw, m.h.Delete)
	}
}
