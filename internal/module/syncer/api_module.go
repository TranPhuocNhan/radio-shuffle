package syncer

import (
	"github.com/gin-gonic/gin"
	"github.com/tranphuocnhan/radio-shuffle/internal/platform/httperr"
)

// APIModule wires syncer admin HTTP routes.
type APIModule struct {
	h       *Handler
	authMw  gin.HandlerFunc
	adminMw gin.HandlerFunc
}

// NewAPIModule constructs a syncer APIModule.
func NewAPIModule(h *Handler, authMw gin.HandlerFunc, adminMw gin.HandlerFunc) *APIModule {
	return &APIModule{
		h:       h,
		authMw:  authMw,
		adminMw: adminMw,
	}
}

// RegisterRoutes mounts /syncer admin endpoints under api.
func (m *APIModule) RegisterRoutes(api *gin.RouterGroup) {
	g := api.Group("/syncer", m.authMw, m.adminMw)
	{
		g.POST("/trigger", httperr.Wrap(m.h.Trigger, MapError))
		g.GET("/status/:request_id", httperr.Wrap(m.h.Status, MapError))
	}
}
