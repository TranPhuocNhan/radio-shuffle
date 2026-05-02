package router

import (
	"context"
	"net/http"
	"time"

	"github.com/tranphuocnhan/radio-shuffle/internal/module/auth"
	"github.com/tranphuocnhan/radio-shuffle/internal/module/playlist"
	"github.com/tranphuocnhan/radio-shuffle/internal/module/station"
	"github.com/tranphuocnhan/radio-shuffle/internal/module/stream"
	"github.com/tranphuocnhan/radio-shuffle/internal/module/track"
	"github.com/tranphuocnhan/radio-shuffle/internal/module/user"
	"github.com/tranphuocnhan/radio-shuffle/internal/platform/config"
	"github.com/tranphuocnhan/radio-shuffle/internal/platform/response"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
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

func Register(api *gin.RouterGroup, cfg *config.Config, pool *pgxpool.Pool) {
	attachAPIRoutes(api)
	registerModules(api, cfg, pool)
}

func registerModules(api *gin.RouterGroup, cfg *config.Config, pool *pgxpool.Pool) {
	userRepo := user.NewRepository(pool)
	authUsers := authUserAdapter{repo: userRepo}
	authTokens := auth.NewTokenRepository(pool)
	auth.NewModule(cfg, authUsers, authUsers, authTokens, nil).RegisterRoutes(api)

	user.RegisterRoutes(api)
	track.RegisterRoutes(api)
	playlist.RegisterRoutes(api)
	stream.RegisterRoutes(api)
	station.NewModule(pool).RegisterRoutes(api)
}

func attachAPIRoutes(api *gin.RouterGroup) {
	api.GET("/ping", func(c *gin.Context) {
		response.OK(c, http.StatusOK, map[string]any{"service": "radio-shuffle"})
	})
}

// authUserAdapter maps user.Repository to auth.UserReader/UserWriter.
type authUserAdapter struct {
	repo user.Repository
}

func (a authUserAdapter) GetByEmail(ctx context.Context, email string) (auth.User, error) {
	row, err := a.repo.GetByEmail(ctx, email)
	if err != nil {
		return auth.User{}, err
	}
	return auth.User{ID: row.ID, Email: row.Email, PasswordHash: row.PasswordHash, Role: row.Role}, nil
}

func (a authUserAdapter) GetByID(ctx context.Context, id int64) (auth.User, error) {
	row, err := a.repo.GetByID(ctx, id)
	if err != nil {
		return auth.User{}, err
	}
	return auth.User{ID: row.ID, Email: row.Email, PasswordHash: row.PasswordHash, Role: row.Role}, nil
}

func (a authUserAdapter) Create(ctx context.Context, in auth.CreateUserInput) (auth.User, error) {
	row, err := a.repo.Create(ctx, user.CreateUserInput{
		Email:        in.Email,
		PasswordHash: in.PasswordHash,
		Role:         in.Role,
	})
	if err != nil {
		return auth.User{}, err
	}
	return auth.User{ID: row.ID, Email: row.Email, PasswordHash: row.PasswordHash, Role: row.Role}, nil
}
