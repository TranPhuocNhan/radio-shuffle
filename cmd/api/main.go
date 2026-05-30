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
	"github.com/tranphuocnhan/radio-shuffle/internal/module/syncer"
	"github.com/tranphuocnhan/radio-shuffle/internal/module/track"
	"github.com/tranphuocnhan/radio-shuffle/internal/module/user"
	"github.com/tranphuocnhan/radio-shuffle/internal/platform/config"
	"github.com/tranphuocnhan/radio-shuffle/internal/platform/database"
	plhealth "github.com/tranphuocnhan/radio-shuffle/internal/platform/health"
	"github.com/tranphuocnhan/radio-shuffle/internal/platform/mq"
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

	mqClient, err := mq.New(mq.Config{
		URL:           cfg.RabbitMQURL,
		Exchange:      cfg.SyncEventsExchange,
		Queue:         cfg.SyncQueue,
		RetryQueue:    cfg.SyncRetryQueue,
		DLQ:           cfg.SyncDLQ,
		RetryTTLMS:    cfg.SyncRetryTTLMS,
		MaxRetries:    cfg.SyncMaxRetries,
		MainRouteKey:  "sync.requested",
		RetryRouteKey: "sync.requested.retry",
		DLQRouteKey:   "sync.requested.dlq",
	})
	if err != nil {
		slog.Error("rabbitmq connect failed", "err", err)
		os.Exit(1)
	}
	defer func() {
		_ = mqClient.Close()
	}()

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
	authSvc := auth.NewService(authUsers, authUsers, authTokens, auth.Config{
		SigningKey: []byte(cfg.JWTSigningKey),
		Issuer:     cfg.JWTIssuer,
		AccessTTL:  cfg.AccessTTL(),
		RefreshTTL: cfg.RefreshTTL(),
	}, nil)
	authHandler := auth.NewHandler(authSvc)
	authModule := auth.NewModule(authHandler)
	playlistRepo := playlist.NewRepository(pool)
	playlistSvc := playlist.NewService(playlistRepo)
	playlistHandler := playlist.NewHandler(playlistSvc)
	playlistAuthMw := plmw.AuthRequired([]byte(cfg.JWTSigningKey), cfg.JWTIssuer)
	playlistModule := playlist.NewModule(playlistHandler, playlistAuthMw)
	stationRepo := station.NewRepository(pool)
	stationSvc := station.NewService(stationRepo)
	stationHandler := station.NewHandler(stationSvc)
	stationAuthMw := plmw.AuthRequired([]byte(cfg.JWTSigningKey), cfg.JWTIssuer)
	stationModule := station.NewModule(stationHandler, stationAuthMw)
	streamRepo := stream.NewRepository(pool)
	streamSvc := stream.NewService(streamRepo)
	streamHandler := stream.NewHandler(streamSvc)
	streamAuthMw := plmw.AuthRequired([]byte(cfg.JWTSigningKey), cfg.JWTIssuer)
	streamModule := stream.NewModule(streamHandler, streamAuthMw)
	radioBrowserRepo := radiobrowser.NewRepository(pool)
	radioBrowserSvc := radiobrowser.NewService(radioBrowserRepo)
	radioBrowserHandler := radiobrowser.NewHandler(radioBrowserSvc)
	radioBrowserModule := radiobrowser.NewModule(radioBrowserHandler)
	trackRepo := track.NewRepository(pool)
	trackSvc := track.NewService(trackRepo)
	trackHandler := track.NewHandler(trackSvc)
	trackAuthMw := plmw.AuthRequired([]byte(cfg.JWTSigningKey), cfg.JWTIssuer)
	trackModule := track.NewModule(trackHandler, trackAuthMw)
	syncerRepo := syncer.NewRepository(pool)
	syncerPublisher := syncer.NewPublisher(mqClient, "sync.requested")
	syncerJobSvc := syncer.NewJobService(syncerRepo, syncerPublisher)
	syncerHandler := syncer.NewHandler(syncerJobSvc)
	syncerAuthMw := plmw.AuthRequired([]byte(cfg.JWTSigningKey), cfg.JWTIssuer)
	syncerAdminMw := plmw.AdminRequired()
	syncerAPIModule := syncer.NewAPIModule(syncerHandler, syncerAuthMw, syncerAdminMw)

	router.Register(
		api,
		router.RouteRegistrarFunc(authModule.RegisterRoutes),
		router.RouteRegistrarFunc(user.RegisterRoutes),
		router.RouteRegistrarFunc(trackModule.RegisterRoutes),
		router.RouteRegistrarFunc(playlistModule.RegisterRoutes),
		router.RouteRegistrarFunc(streamModule.RegisterRoutes),
		router.RouteRegistrarFunc(stationModule.RegisterRoutes),
		router.RouteRegistrarFunc(syncerAPIModule.RegisterRoutes),
		router.RouteRegistrarFunc(radioBrowserModule.RegisterRoutes),
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
