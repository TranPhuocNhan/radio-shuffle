package main

import (
	"context"

	"github.com/rabbitmq/amqp091-go"
	"github.com/tranphuocnhan/radio-shuffle/internal/platform/mq"
)

const (
	headerRequestID  = "x-request-id"
	headerRetryCount = "x-retry-count"
)

type deliveryMeta struct {
	RequestID   string
	RetryCount  int
	RoutingKey  string
	DeliveryTag uint64
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
	if v, ok := headers[headerRequestID]; ok {
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
	if v, ok := headers[headerRetryCount]; ok {
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
	return client.Publish(ctx, syncRetryRouteKey, msg.Body, headers)
}

func publishToDLQ(ctx context.Context, client *mq.Client, msg amqp091.Delivery, retryCount int) error {
	headers := headersWithRetryCount(msg.Headers, retryCount)
	return client.Publish(ctx, syncDLQRouteKey, msg.Body, headers)
}

func headersWithRetryCount(headers amqp091.Table, retryCount int) amqp091.Table {
	next := amqp091.Table{}
	for key, value := range headers {
		next[key] = value
	}
	next[headerRetryCount] = int32(retryCount)
	return next
}
