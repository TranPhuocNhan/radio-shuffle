package station

import (
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	plmw "github.com/tranphuocnhan/radio-shuffle/internal/platform/mw"
)

// Module wires station HTTP routes.
type Module struct {
	h      *Handler
	authMw gin.HandlerFunc
}

// NewModule constructs a station Module.
// signingKey and issuer are required to build the AuthRequired middleware for
// follow endpoints.
func NewModule(pool *pgxpool.Pool, signingKey []byte, issuer string) *Module {
	repo := NewRepository(pool)
	svc := NewService(repo)
	return &Module{
		h:      NewHandler(svc),
		authMw: plmw.AuthRequired(signingKey, issuer),
	}
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

		g.POST("/:station_id/follow", m.authMw, m.h.Follow)
		g.DELETE("/:station_id/follow", m.authMw, m.h.Unfollow)
		g.GET("/:station_id/following", m.authMw, m.h.IsFollowing)
		g.GET("/:station_id/followers/count", m.authMw, m.h.CountFollowers)
		g.GET("/followed", m.authMw, m.h.ListFollowed)
	}
}
