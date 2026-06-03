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

type deliveryMeta struct {
	RequestID   string
	RetryCount  int
	RoutingKey  string
	DeliveryTag uint64
}

type failedMessageRoute struct {
	RoutingKey string
	RetryCount int
	ToDLQ      bool
}

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

	slog.Error("invalid sync request sent to dlq", meta.logAttrs("target_routing_key", mq.SyncDLQRouteKey)...)
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
	route := failedMessageRouteFor(meta, h.maxRetries)
	if route.ToDLQ {
		slog.Error("sync failed, sending to dlq", meta.logAttrs("err", processErr)...)
		if err := publishToDLQ(ctx, h.mqClient, msg, route.RetryCount); err != nil {
			slog.Error("dlq publish failed", meta.logAttrs("err", err)...)
			nackMessage(msg, cmd.RequestID, true, "dlq publish failed nacked")
			return
		}
		slog.Error("sync message sent to dlq", meta.logAttrs("target_routing_key", route.RoutingKey)...)
		ackMessage(msg, cmd.RequestID, "sync dlq acknowledged")
		return
	}

	if err := publishToRetry(ctx, h.mqClient, msg, route.RetryCount); err != nil {
		slog.Error("retry publish failed", meta.logAttrs("err", err)...)
		nackMessage(msg, cmd.RequestID, true, "retry publish failed nacked")
		return
	}
	slog.Warn("sync failed, retry scheduled", meta.logAttrs("retry", route.RetryCount, "err", processErr)...)
	ackMessage(msg, cmd.RequestID, "sync retry acknowledged")
}

func failedMessageRouteFor(meta deliveryMeta, maxRetries int) failedMessageRoute {
	if meta.RetryCount >= maxRetries {
		return failedMessageRoute{
			RoutingKey: mq.SyncDLQRouteKey,
			RetryCount: meta.RetryCount,
			ToDLQ:      true,
		}
	}
	return failedMessageRoute{
		RoutingKey: mq.SyncRetryRouteKey,
		RetryCount: meta.RetryCount + 1,
	}
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

func deliveryMetaFrom(msg amqp091.Delivery) deliveryMeta {
	return deliveryMeta{
		RequestID:   requestIDFromHeaders(msg.Headers),
		RetryCount:  retryCountFromHeaders(msg.Headers),
		RoutingKey:  msg.RoutingKey,
		DeliveryTag: msg.DeliveryTag,
	}
}

func (m deliveryMeta) logAttrs(extra ...any) []any {
	attrs := []any{
		"request_id", m.RequestID,
		"routing_key", m.RoutingKey,
		"delivery_tag", m.DeliveryTag,
		"retry", m.RetryCount,
	}
	return append(attrs, extra...)
}

func ackMessage(msg amqp091.Delivery, requestID string, successMessage string) {
	meta := deliveryMetaFrom(msg)
	meta.RequestID = requestID

	if err := msg.Ack(false); err != nil {
		logAckError("rabbitmq ack failed", meta, err, nil)
		return
	}
	logDeliverySuccess(successMessage, meta, nil)
}

func nackMessage(msg amqp091.Delivery, requestID string, requeue bool, successMessage string) {
	meta := deliveryMetaFrom(msg)
	meta.RequestID = requestID

	if err := msg.Nack(false, requeue); err != nil {
		logAckError("rabbitmq nack failed", meta, err, []any{"requeue", requeue})
		return
	}
	logDeliverySuccess(successMessage, meta, []any{"requeue", requeue})
}

func requestIDFromHeaders(headers amqp091.Table) string {
	if headers == nil {
		return ""
	}
	if v, ok := headers[mq.HeaderRequestID]; ok {
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
	if v, ok := headers[mq.HeaderRetryCount]; ok {
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
	headers := headersWithRetryCount(msg.Headers, retryCount)
	return client.Publish(ctx, mq.SyncRetryRouteKey, msg.Body, headers)
}

func publishToDLQ(ctx context.Context, client *mq.Client, msg amqp091.Delivery, retryCount int) error {
	headers := headersWithRetryCount(msg.Headers, retryCount)
	return client.Publish(ctx, mq.SyncDLQRouteKey, msg.Body, headers)
}

func headersWithRetryCount(headers amqp091.Table, retryCount int) amqp091.Table {
	next := amqp091.Table{}
	for key, value := range headers {
		next[key] = value
	}
	next[mq.HeaderRetryCount] = int32(retryCount)
	return next
}

func logAckError(message string, meta deliveryMeta, err error, extra []any) {
	attrs := meta.logAttrs(extra...)
	attrs = append(attrs, "err", err)
	slog.Error(message, attrs...)
}

func logDeliverySuccess(message string, meta deliveryMeta, extra []any) {
	slog.Info(message, meta.logAttrs(extra...)...)
}
