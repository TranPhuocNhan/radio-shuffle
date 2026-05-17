package syncer

import (
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/tranphuocnhan/radio-shuffle/internal/platform/mw"
)

// APIModule wires syncer admin HTTP routes.
type APIModule struct {
	h       *Handler
	authMw  gin.HandlerFunc
	adminMw gin.HandlerFunc
}

// NewAPIModule constructs a syncer APIModule.
func NewAPIModule(pool *pgxpool.Pool, publisher JobPublisher, signingKey []byte, issuer string) *APIModule {
	repo := NewRepository(pool)
	svc := NewJobService(repo, publisher)
	return &APIModule{
		h:       NewHandler(svc),
		authMw:  mw.AuthRequired(signingKey, issuer),
		adminMw: mw.AdminRequired(),
	}
}

// RegisterRoutes mounts /syncer admin endpoints under api.
func (m *APIModule) RegisterRoutes(api *gin.RouterGroup) {
	g := api.Group("/syncer", m.authMw, m.adminMw)
	{
		g.POST("/trigger", m.h.Trigger)
		g.GET("/status/:request_id", m.h.Status)
	}
}

