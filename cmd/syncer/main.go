package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/tranphuocnhan/radio-shuffle/internal/platform/config"
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

	app, cleanup, err := newApp(ctx, cfg)
	if err != nil {
		slog.Error("syncer bootstrap failed", "err", err)
		os.Exit(1)
	}
	defer cleanup()

	if err := app.run(ctx); err != nil {
		slog.Error("syncer stopped", "err", err)
		os.Exit(1)
	}
}
