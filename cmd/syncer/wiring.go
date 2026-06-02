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
	syncRequestedRouteKey = "sync.requested"
	syncRetryRouteKey     = "sync.requested.retry"
	syncDLQRouteKey       = "sync.requested.dlq"
	syncerPrefetchCount   = 1
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
	syncSvc := syncer.NewModule(syncer.NewService(repo, rbClient)).Service()
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
		MainRouteKey:  syncRequestedRouteKey,
		RetryRouteKey: syncRetryRouteKey,
		DLQRouteKey:   syncDLQRouteKey,
		PrefetchCount: syncerPrefetchCount,
	}
}
