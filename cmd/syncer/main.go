package main

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/rabbitmq/amqp091-go"
	"github.com/tranphuocnhan/radio-shuffle/internal/module/syncer"
	"github.com/tranphuocnhan/radio-shuffle/internal/platform/config"
	"github.com/tranphuocnhan/radio-shuffle/internal/platform/database"
	"github.com/tranphuocnhan/radio-shuffle/internal/platform/mq"
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
		PrefetchCount: 1,
	})
	if err != nil {
		slog.Error("rabbitmq connect failed", "err", err)
		os.Exit(1)
	}
	defer func() {
		_ = mqClient.Close()
	}()

	rbClient := radiobrowser.NewClient(cfg.RadioBrowserBaseURL)
	mod := syncer.NewModule(pool, rbClient)
	syncSvc := mod.Service()
	processor := syncer.NewJobProcessor(syncer.NewRepository(pool), syncSvc)

	msgs, err := mqClient.Consume("syncer-worker")
	if err != nil {
		slog.Error("rabbitmq consume failed", "err", err)
		os.Exit(1)
	}

	slog.Info("syncer started", "queue", cfg.SyncQueue)

	for {
		select {
		case msg := <-msgs:
			handleMessage(ctx, mqClient, processor, cfg, msg)
		case <-ctx.Done():
			slog.Info("syncer shutting down")
			return
		}
	}
}

func handleMessage(ctx context.Context, mqClient *mq.Client, processor syncer.JobProcessor, cfg config.Config, msg amqp091.Delivery) {
	var req syncer.SyncRequest
	if err := json.Unmarshal(msg.Body, &req); err != nil {
		slog.Error("invalid sync request", "err", err)
		_ = publishToDLQ(ctx, mqClient, msg, 0)
		_ = msg.Ack(false)
		return
	}

	retryCount := retryCountFromHeaders(msg.Headers)

	result, err := processor.Process(ctx, req)
	if err == nil {
		slog.Info("sync completed", "request_id", req.RequestID, "fetched", result.Fetched, "upserted", result.Upserted)
		_ = msg.Ack(false)
		return
	}
	if errors.Is(err, syncer.ErrJobDuplicate) || errors.Is(err, syncer.ErrJobInProgress) {
		slog.Warn("sync skipped", "request_id", req.RequestID, "err", err)
		_ = msg.Ack(false)
		return
	}

	if retryCount >= cfg.SyncMaxRetries {
		slog.Error("sync failed, sending to dlq", "request_id", req.RequestID, "retry", retryCount, "err", err)
		if err := publishToDLQ(ctx, mqClient, msg, retryCount); err != nil {
			slog.Error("dlq publish failed", "err", err)
			_ = msg.Nack(false, true)
			return
		}
		_ = msg.Ack(false)
		return
	}

	if err := publishToRetry(ctx, mqClient, msg, retryCount+1); err != nil {
		slog.Error("retry publish failed", "err", err)
		_ = msg.Nack(false, true)
		return
	}
	slog.Warn("sync failed, retry scheduled", "request_id", req.RequestID, "retry", retryCount+1, "err", err)
	_ = msg.Ack(false)
}

func retryCountFromHeaders(headers amqp091.Table) int {
	if headers == nil {
		return 0
	}
	if v, ok := headers["x-retry-count"]; ok {
		switch typed := v.(type) {
		case int32:
			return int(typed)
		case int64:
			return int(typed)
		case int:
			return typed
		case float32:
			return int(typed)
		case float64:
			return int(typed)
		}
	}
	return 0
}

func publishToRetry(ctx context.Context, client *mq.Client, msg amqp091.Delivery, retryCount int) error {
	headers := msg.Headers
	if headers == nil {
		headers = amqp091.Table{}
	}
	headers["x-retry-count"] = int32(retryCount)
	return client.Publish(ctx, "sync.requested.retry", msg.Body, headers)
}

func publishToDLQ(ctx context.Context, client *mq.Client, msg amqp091.Delivery, retryCount int) error {
	headers := msg.Headers
	if headers == nil {
		headers = amqp091.Table{}
	}
	headers["x-retry-count"] = int32(retryCount)
	return client.Publish(ctx, "sync.requested.dlq", msg.Body, headers)
}

func setupLogger() {
	h := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo})
	slog.SetDefault(slog.New(h))
}
