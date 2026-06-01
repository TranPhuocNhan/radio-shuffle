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
	syncerRepo := syncer.NewRepository(pool)
	syncService := syncer.NewService(syncerRepo, rbClient)
	mod := syncer.NewModule(syncService)
	syncSvc := mod.Service()
	processor := syncer.NewJobProcessor(syncerRepo, syncSvc)

	msgs, err := mqClient.Consume("syncer-worker")
	if err != nil {
		slog.Error("rabbitmq consume failed", "err", err)
		os.Exit(1)
	}

	slog.Info("syncer started", "queue", cfg.SyncQueue)

	for {
		select {
		case msg, ok := <-msgs:
			if !ok {
				slog.Error("rabbitmq consumer channel closed", "queue", cfg.SyncQueue)
				return
			}
			handleMessage(ctx, mqClient, processor, cfg, msg)
		case <-ctx.Done():
			slog.Info("syncer shutting down")
			return
		}
	}
}

func handleMessage(ctx context.Context, mqClient *mq.Client, processor syncer.JobProcessor, cfg config.Config, msg amqp091.Delivery) {
	headerRequestID := requestIDFromHeaders(msg.Headers)
	retryCount := retryCountFromHeaders(msg.Headers)
	slog.Info(
		"rabbitmq message received",
		"request_id", headerRequestID,
		"routing_key", msg.RoutingKey,
		"delivery_tag", msg.DeliveryTag,
		"retry", retryCount,
	)

	var payload mq.SyncRequestedMessage
	if err := json.Unmarshal(msg.Body, &payload); err != nil {
		slog.Error(
			"invalid sync request",
			"request_id", headerRequestID,
			"routing_key", msg.RoutingKey,
			"delivery_tag", msg.DeliveryTag,
			"retry", retryCount,
			"err", err,
		)
		if err := publishToDLQ(ctx, mqClient, msg, 0); err != nil {
			slog.Error(
				"invalid sync request dlq publish failed",
				"request_id", headerRequestID,
				"routing_key", msg.RoutingKey,
				"delivery_tag", msg.DeliveryTag,
				"err", err,
			)
			nackMessage(msg, headerRequestID, true, "invalid sync request nacked")
			return
		}
		slog.Error(
			"invalid sync request sent to dlq",
			"request_id", headerRequestID,
			"routing_key", "sync.requested.dlq",
			"delivery_tag", msg.DeliveryTag,
		)
		ackMessage(msg, headerRequestID, "invalid sync request acknowledged")
		return
	}
	cmd := syncer.FromSyncRequestedMessage(payload)

	slog.Info(
		"sync processing started",
		"request_id", cmd.RequestID,
		"requested_by", cmd.RequestedBy,
		"routing_key", msg.RoutingKey,
		"delivery_tag", msg.DeliveryTag,
		"retry", retryCount,
	)
	result, err := processor.Process(ctx, cmd)
	if err == nil {
		slog.Info(
			"sync business result",
			"request_id", cmd.RequestID,
			"requested_by", cmd.RequestedBy,
			"status", "completed",
			"fetched", result.Fetched,
			"upserted", result.Upserted,
			"retry", retryCount,
		)
		ackMessage(msg, cmd.RequestID, "sync completed acknowledged")
		return
	}
	if errors.Is(err, syncer.ErrJobDuplicate) || errors.Is(err, syncer.ErrJobInProgress) {
		slog.Warn(
			"sync business result",
			"request_id", cmd.RequestID,
			"requested_by", cmd.RequestedBy,
			"status", "skipped",
			"fetched", result.Fetched,
			"upserted", result.Upserted,
			"retry", retryCount,
			"err", err,
		)
		ackMessage(msg, cmd.RequestID, "sync skipped acknowledged")
		return
	}
	slog.Error(
		"sync business result",
		"request_id", cmd.RequestID,
		"requested_by", cmd.RequestedBy,
		"status", "failed",
		"fetched", result.Fetched,
		"upserted", result.Upserted,
		"retry", retryCount,
		"err", err,
	)

	if retryCount >= cfg.SyncMaxRetries {
		slog.Error("sync failed, sending to dlq", "request_id", cmd.RequestID, "retry", retryCount, "err", err)
		if err := publishToDLQ(ctx, mqClient, msg, retryCount); err != nil {
			slog.Error("dlq publish failed", "err", err)
			nackMessage(msg, cmd.RequestID, true, "dlq publish failed nacked")
			return
		}
		slog.Error(
			"sync message sent to dlq",
			"request_id", cmd.RequestID,
			"routing_key", "sync.requested.dlq",
			"delivery_tag", msg.DeliveryTag,
			"retry", retryCount,
		)
		ackMessage(msg, cmd.RequestID, "sync dlq acknowledged")
		return
	}

	if err := publishToRetry(ctx, mqClient, msg, retryCount+1); err != nil {
		slog.Error("retry publish failed", "err", err)
		nackMessage(msg, cmd.RequestID, true, "retry publish failed nacked")
		return
	}
	slog.Warn("sync failed, retry scheduled", "request_id", cmd.RequestID, "retry", retryCount+1, "err", err)
	ackMessage(msg, cmd.RequestID, "sync retry acknowledged")
}

func ackMessage(msg amqp091.Delivery, requestID string, successMessage string) {
	if err := msg.Ack(false); err != nil {
		slog.Error(
			"rabbitmq ack failed",
			"request_id", requestID,
			"routing_key", msg.RoutingKey,
			"delivery_tag", msg.DeliveryTag,
			"err", err,
		)
		return
	}
	slog.Info(
		successMessage,
		"request_id", requestID,
		"routing_key", msg.RoutingKey,
		"delivery_tag", msg.DeliveryTag,
	)
}

func nackMessage(msg amqp091.Delivery, requestID string, requeue bool, successMessage string) {
	if err := msg.Nack(false, requeue); err != nil {
		slog.Error(
			"rabbitmq nack failed",
			"request_id", requestID,
			"routing_key", msg.RoutingKey,
			"delivery_tag", msg.DeliveryTag,
			"requeue", requeue,
			"err", err,
		)
		return
	}
	slog.Warn(
		successMessage,
		"request_id", requestID,
		"routing_key", msg.RoutingKey,
		"delivery_tag", msg.DeliveryTag,
		"requeue", requeue,
	)
}

func requestIDFromHeaders(headers amqp091.Table) string {
	if headers == nil {
		return ""
	}
	if v, ok := headers["x-request-id"]; ok {
		if requestID, ok := v.(string); ok {
			return requestID
		}
	}
	return ""
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
