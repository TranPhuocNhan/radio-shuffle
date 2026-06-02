package main

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"

	"github.com/rabbitmq/amqp091-go"
	"github.com/tranphuocnhan/radio-shuffle/internal/module/syncer"
	"github.com/tranphuocnhan/radio-shuffle/internal/platform/mq"
)

type messageHandler struct {
	mqClient   *mq.Client
	processor  syncer.JobProcessor
	maxRetries int
}

func newMessageHandler(mqClient *mq.Client, processor syncer.JobProcessor, maxRetries int) *messageHandler {
	return &messageHandler{
		mqClient:   mqClient,
		processor:  processor,
		maxRetries: maxRetries,
	}
}

func (h *messageHandler) handle(ctx context.Context, msg amqp091.Delivery) {
	meta := deliveryMetaFrom(msg)
	slog.Info("rabbitmq message received", meta.logAttrs()...)

	cmd, err := decodeSyncCommand(msg.Body)
	if err != nil {
		h.handleInvalidMessage(ctx, msg, meta, err)
		return
	}

	h.processCommand(ctx, msg, meta, cmd)
}

func decodeSyncCommand(body []byte) (syncer.SyncCommand, error) {
	var payload mq.SyncRequestedMessage
	if err := json.Unmarshal(body, &payload); err != nil {
		return syncer.SyncCommand{}, err
	}
	return syncer.FromSyncRequestedMessage(payload), nil
}

func (h *messageHandler) handleInvalidMessage(ctx context.Context, msg amqp091.Delivery, meta deliveryMeta, decodeErr error) {
	slog.Error("invalid sync request", meta.logAttrs("err", decodeErr)...)

	if err := publishToDLQ(ctx, h.mqClient, msg, 0); err != nil {
		slog.Error("invalid sync request dlq publish failed", meta.logAttrs("err", err)...)
		nackMessage(msg, meta.RequestID, true, "invalid sync request nacked")
		return
	}

	slog.Error("invalid sync request sent to dlq", meta.logAttrs("target_routing_key", syncDLQRouteKey)...)
	ackMessage(msg, meta.RequestID, "invalid sync request acknowledged")
}

func (h *messageHandler) processCommand(ctx context.Context, msg amqp091.Delivery, meta deliveryMeta, cmd syncer.SyncCommand) {
	meta.RequestID = cmd.RequestID
	slog.Info(
		"sync processing started",
		meta.logAttrs("requested_by", cmd.RequestedBy)...,
	)

	result, err := h.processor.Process(ctx, cmd)
	if err == nil {
		logSyncResult(slog.LevelInfo, "completed", cmd, result, meta, nil)
		ackMessage(msg, cmd.RequestID, "sync completed acknowledged")
		return
	}
	if isSkippableJobError(err) {
		logSyncResult(slog.LevelWarn, "skipped", cmd, result, meta, err)
		ackMessage(msg, cmd.RequestID, "sync skipped acknowledged")
		return
	}

	logSyncResult(slog.LevelError, "failed", cmd, result, meta, err)
	h.routeFailedMessage(ctx, msg, meta, cmd, err)
}

func (h *messageHandler) routeFailedMessage(
	ctx context.Context,
	msg amqp091.Delivery,
	meta deliveryMeta,
	cmd syncer.SyncCommand,
	processErr error,
) {
	if meta.RetryCount >= h.maxRetries {
		slog.Error("sync failed, sending to dlq", meta.logAttrs("err", processErr)...)
		if err := publishToDLQ(ctx, h.mqClient, msg, meta.RetryCount); err != nil {
			slog.Error("dlq publish failed", meta.logAttrs("err", err)...)
			nackMessage(msg, cmd.RequestID, true, "dlq publish failed nacked")
			return
		}
		slog.Error("sync message sent to dlq", meta.logAttrs("target_routing_key", syncDLQRouteKey)...)
		ackMessage(msg, cmd.RequestID, "sync dlq acknowledged")
		return
	}

	nextRetry := meta.RetryCount + 1
	if err := publishToRetry(ctx, h.mqClient, msg, nextRetry); err != nil {
		slog.Error("retry publish failed", meta.logAttrs("err", err)...)
		nackMessage(msg, cmd.RequestID, true, "retry publish failed nacked")
		return
	}
	slog.Warn("sync failed, retry scheduled", meta.logAttrs("retry", nextRetry, "err", processErr)...)
	ackMessage(msg, cmd.RequestID, "sync retry acknowledged")
}

func isSkippableJobError(err error) bool {
	return errors.Is(err, syncer.ErrJobDuplicate) || errors.Is(err, syncer.ErrJobInProgress)
}

func logSyncResult(level slog.Level, status string, cmd syncer.SyncCommand, result syncer.SyncResult, meta deliveryMeta, err error) {
	attrs := meta.logAttrs(
		"requested_by", cmd.RequestedBy,
		"status", status,
		"fetched", result.Fetched,
		"upserted", result.Upserted,
	)
	if err != nil {
		attrs = append(attrs, "err", err)
	}

	switch level {
	case slog.LevelError:
		slog.Error("sync business result", attrs...)
	case slog.LevelWarn:
		slog.Warn("sync business result", attrs...)
	default:
		slog.Info("sync business result", attrs...)
	}
}
