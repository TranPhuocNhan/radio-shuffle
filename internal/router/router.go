package router

import (
	"context"
	"net/http"
	"time"

	"github.com/tranphuocnhan/radio-shuffle/internal/platform/config"
	"github.com/tranphuocnhan/radio-shuffle/internal/platform/response"

	"github.com/gin-gonic/gin"
)

// Server is a thin wrapper so main can call Shutdown with a context.
type Server struct {
	inner *http.Server
}

func NewServer(cfg *config.Config, handler http.Handler) *Server {
	return &Server{
		inner: &http.Server{
			Addr:              cfg.HTTPAddr,
			Handler:           handler,
			ReadTimeout:       cfg.ReadTimeout(),
			ReadHeaderTimeout: 10 * time.Second,
			WriteTimeout:      cfg.WriteTimeout(),
			IdleTimeout:       90 * time.Second,
		},
	}
}

func (s *Server) Listen() error {
	return s.inner.ListenAndServe()
}

func (s *Server) Shutdown(ctx context.Context) error {
	shutdownDeadline, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()
	return s.inner.Shutdown(shutdownDeadline)
}

// RouteRegistrar mounts routes under the api group.
type RouteRegistrar interface {
	RegisterRoutes(*gin.RouterGroup)
}

// RouteRegistrarFunc adapts a function to a RouteRegistrar.
type RouteRegistrarFunc func(*gin.RouterGroup)

func (f RouteRegistrarFunc) RegisterRoutes(api *gin.RouterGroup) {
	f(api)
}

func Register(api *gin.RouterGroup, registrars ...RouteRegistrar) {
	attachAPIRoutes(api)
	for _, registrar := range registrars {
		registrar.RegisterRoutes(api)
	}
}

func attachAPIRoutes(api *gin.RouterGroup) {
	api.GET("/ping", func(c *gin.Context) {
		response.OK(c, http.StatusOK, map[string]any{"service": "radio-shuffle"})
	})
}
