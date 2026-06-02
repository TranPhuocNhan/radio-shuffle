package main

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/tranphuocnhan/radio-shuffle/internal/module/syncer"
	"github.com/tranphuocnhan/radio-shuffle/internal/platform/config"
	"github.com/tranphuocnhan/radio-shuffle/internal/platform/mq"
)

const consumerName = "syncer-worker"

type app struct {
	cfg      config.Config
	mqClient *mq.Client
	handler  *messageHandler
}

func newSyncerApp(cfg config.Config, mqClient *mq.Client, processor syncer.JobProcessor) *app {
	return &app{
		cfg:      cfg,
		mqClient: mqClient,
		handler:  newMessageHandler(mqClient, processor, cfg.SyncMaxRetries),
	}
}

func (a *app) run(ctx context.Context) error {
	msgs, err := a.mqClient.Consume(consumerName)
	if err != nil {
		return fmt.Errorf("consume rabbitmq messages: %w", err)
	}

	slog.Info("syncer started", "queue", a.cfg.SyncQueue)
	for {
		select {
		case msg, ok := <-msgs:
			if !ok {
				return fmt.Errorf("rabbitmq consumer channel closed: queue=%s", a.cfg.SyncQueue)
			}
			a.handler.handle(ctx, msg)
		case <-ctx.Done():
			slog.Info("syncer shutting down")
			return nil
		}
	}
}
