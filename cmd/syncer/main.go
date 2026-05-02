package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/tranphuocnhan/radio-shuffle/internal/module/syncer"
	"github.com/tranphuocnhan/radio-shuffle/internal/platform/config"
	"github.com/tranphuocnhan/radio-shuffle/internal/platform/database"
	"github.com/tranphuocnhan/radio-shuffle/internal/platform/radiobrowser"
)

func main() {
	setupLogger()

	cfg, err := config.LoadSyncer()
	if err != nil {
		slog.Error("config load failed", "err", err)
		os.Exit(1)
	}

	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	pool, stopPool, err := database.NewPool(ctx, cfg.DatabaseURL)
	if err != nil {
		slog.Error("database pool failed", "err", err)
		os.Exit(1)
	}
	defer stopPool()

	rbClient := radiobrowser.NewClient(cfg.RadioBrowserBaseURL)
	mod := syncer.NewModule(pool, rbClient)
	svc := mod.Service()

	interval := cfg.SyncInterval()
	slog.Info("syncer started", "interval", interval.String())

	runSync(ctx, svc)

	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			runSync(ctx, svc)
		case <-ctx.Done():
			slog.Info("syncer shutting down")
			return
		}
	}
}

func runSync(ctx context.Context, svc syncer.Service) {
	slog.Info("sync started")
	result, err := svc.Sync(ctx)
	if err != nil {
		slog.Error("sync failed", "err", err)
		return
	}
	slog.Info("sync completed", "fetched", result.Fetched, "upserted", result.Upserted)
}

func setupLogger() {
	h := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo})
	slog.SetDefault(slog.New(h))
}
