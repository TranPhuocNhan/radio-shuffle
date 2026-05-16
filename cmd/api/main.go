package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/tranphuocnhan/radio-shuffle/internal/module/auth"
	authadapters "github.com/tranphuocnhan/radio-shuffle/internal/module/auth/adapters"
	"github.com/tranphuocnhan/radio-shuffle/internal/module/playlist"
	"github.com/tranphuocnhan/radio-shuffle/internal/module/radiobrowser"
	"github.com/tranphuocnhan/radio-shuffle/internal/module/station"
	"github.com/tranphuocnhan/radio-shuffle/internal/module/stream"
	"github.com/tranphuocnhan/radio-shuffle/internal/module/track"
	"github.com/tranphuocnhan/radio-shuffle/internal/module/user"
	"github.com/tranphuocnhan/radio-shuffle/internal/platform/config"
	"github.com/tranphuocnhan/radio-shuffle/internal/platform/database"
	plhealth "github.com/tranphuocnhan/radio-shuffle/internal/platform/health"
	plmw "github.com/tranphuocnhan/radio-shuffle/internal/platform/mw"
	"github.com/tranphuocnhan/radio-shuffle/internal/router"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func main() {
	setupLogger()
	rootCtx := context.Background()

	cfg, err := config.Load()
	if err != nil {
		slog.Error("config load failed", "err", err)
		os.Exit(1)
	}

	pool, stopPool, err := database.NewPool(rootCtx, cfg.DatabaseURL)
	if err != nil {
		slog.Error("database pool failed", "err", err)
		os.Exit(1)
	}
	defer stopPool()

	gin.SetMode(cfg.GinMode())
	r := gin.New()
	r.Use(plmw.RequestID(), plmw.Recover(), plmw.LoggerStructured())

	cmw := cors.New(cors.Config{
		AllowOrigins: cfg.CORSAllowedOriginsList(),
		AllowMethods: []string{"GET", "POST", "PATCH", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders: []string{"Origin", "Content-Type", "Authorization", "X-Request-ID"},
	})
	r.Use(cmw)

	healthHandlers := plhealth.New(pool)
	r.GET("/health", healthHandlers.Live)
	r.GET("/ready", healthHandlers.Ready)

	api := r.Group("/api/v1")

	userRepo := user.NewRepository(pool)
	authUsers := authadapters.NewAuthUserAdapter(userRepo)
	authTokens := auth.NewTokenRepository(pool)

	router.Register(
		api,
		router.RouteRegistrarFunc(auth.NewModule(&cfg, authUsers, authUsers, authTokens, nil).RegisterRoutes),
		router.RouteRegistrarFunc(user.RegisterRoutes),
		router.RouteRegistrarFunc(track.NewModule(pool, []byte(cfg.JWTSigningKey), cfg.JWTIssuer).RegisterRoutes),
		router.RouteRegistrarFunc(playlist.RegisterRoutes),
		router.RouteRegistrarFunc(stream.RegisterRoutes),
		router.RouteRegistrarFunc(station.NewModule(pool, []byte(cfg.JWTSigningKey), cfg.JWTIssuer).RegisterRoutes),
		router.RouteRegistrarFunc(radiobrowser.NewModule(pool).RegisterRoutes),
	)

	srv := router.NewServer(&cfg, r)

	go func() {
		if err := srv.Listen(); err != nil && err != http.ErrServerClosed {
			slog.Error("http server stopped", "err", err)
			os.Exit(1)
		}
	}()

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
	<-sig

	ctx, cancel := context.WithTimeout(context.Background(), cfg.ShutdownTimeout())
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		slog.Error("shutdown", "err", err)
	}
}

func setupLogger() {
	h := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo})
	slog.SetDefault(slog.New(h))
}
