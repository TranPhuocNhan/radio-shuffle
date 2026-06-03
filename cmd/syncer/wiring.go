package main

import (
	"context"

	"github.com/tranphuocnhan/radio-shuffle/internal/module/syncer"
	"github.com/tranphuocnhan/radio-shuffle/internal/platform/config"
	"github.com/tranphuocnhan/radio-shuffle/internal/platform/database"
	"github.com/tranphuocnhan/radio-shuffle/internal/platform/mq"
	"github.com/tranphuocnhan/radio-shuffle/internal/platform/radiobrowser"
)

const (
	syncerPrefetchCount = 1
)

func newApp(ctx context.Context, cfg config.Config) (*app, func(), error) {
	pool, stopPool, err := database.NewPool(ctx, cfg.DatabaseURL)
	if err != nil {
		return nil, nil, err
	}

	mqClient, err := mq.New(syncerMQConfig(cfg))
	if err != nil {
		stopPool()
		return nil, nil, err
	}

	repo := syncer.NewRepository(pool)
	rbClient := radiobrowser.NewClient(cfg.RadioBrowserBaseURL)
	syncSvc := syncer.NewService(repo, rbClient)
	processor := syncer.NewJobProcessor(repo, syncSvc)

	cleanup := func() {
		_ = mqClient.Close()
		stopPool()
	}
	return newSyncerApp(cfg, mqClient, processor), cleanup, nil
}

func syncerMQConfig(cfg config.Config) mq.Config {
	return mq.Config{
		URL:           cfg.RabbitMQURL,
		Exchange:      cfg.SyncEventsExchange,
		Queue:         cfg.SyncQueue,
		RetryQueue:    cfg.SyncRetryQueue,
		DLQ:           cfg.SyncDLQ,
		RetryTTLMS:    cfg.SyncRetryTTLMS,
		MaxRetries:    cfg.SyncMaxRetries,
		MainRouteKey:  mq.SyncRequestedRouteKey,
		RetryRouteKey: mq.SyncRetryRouteKey,
		DLQRouteKey:   mq.SyncDLQRouteKey,
		PrefetchCount: syncerPrefetchCount,
	}
}
