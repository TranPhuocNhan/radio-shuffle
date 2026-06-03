package main

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/rabbitmq/amqp091-go"
	"github.com/tranphuocnhan/radio-shuffle/internal/platform/mq"
)

func TestDecodeSyncCommand(t *testing.T) {
	requestedAt := time.Date(2026, 6, 3, 12, 0, 0, 0, time.UTC)
	body, err := json.Marshal(mq.SyncRequestedMessage{
		RequestID:   "req-1",
		RequestedBy: "admin-1",
		Scope:       json.RawMessage(`{"mode":"full"}`),
		RequestedAt: requestedAt,
	})
	if err != nil {
		t.Fatalf("marshal payload: %v", err)
	}

	cmd, err := decodeSyncCommand(body)
	if err != nil {
		t.Fatalf("decode command: %v", err)
	}
	if cmd.RequestID != "req-1" {
		t.Fatalf("RequestID = %q, want req-1", cmd.RequestID)
	}
	if cmd.RequestedBy != "admin-1" {
		t.Fatalf("RequestedBy = %q, want admin-1", cmd.RequestedBy)
	}
	if string(cmd.Scope) != `{"mode":"full"}` {
		t.Fatalf("Scope = %s, want full mode", cmd.Scope)
	}
	if !cmd.RequestedAt.Equal(requestedAt) {
		t.Fatalf("RequestedAt = %s, want %s", cmd.RequestedAt, requestedAt)
	}
}

func TestDecodeSyncCommandRejectsInvalidJSON(t *testing.T) {
	if _, err := decodeSyncCommand([]byte(`{"request_id"`)); err == nil {
		t.Fatal("expected invalid JSON error")
	}
}

func TestDeliveryMetaFromHeaders(t *testing.T) {
	msg := amqp091.Delivery{
		Headers: amqp091.Table{
			mq.HeaderRequestID:  "req-1",
			mq.HeaderRetryCount: int32(2),
		},
		RoutingKey:  "sync.requested",
		DeliveryTag: 42,
	}

	meta := deliveryMetaFrom(msg)
	if meta.RequestID != "req-1" {
		t.Fatalf("RequestID = %q, want req-1", meta.RequestID)
	}
	if meta.RetryCount != 2 {
		t.Fatalf("RetryCount = %d, want 2", meta.RetryCount)
	}
	if meta.RoutingKey != "sync.requested" {
		t.Fatalf("RoutingKey = %q, want sync.requested", meta.RoutingKey)
	}
	if meta.DeliveryTag != 42 {
		t.Fatalf("DeliveryTag = %d, want 42", meta.DeliveryTag)
	}
}

func TestRetryCountFromHeadersAcceptsAMQPNumberTypes(t *testing.T) {
	cases := []struct {
		name  string
		value any
		want  int
	}{
		{name: "int32", value: int32(1), want: 1},
		{name: "int64", value: int64(2), want: 2},
		{name: "int", value: 3, want: 3},
		{name: "float32", value: float32(4), want: 4},
		{name: "float64", value: float64(5), want: 5},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := retryCountFromHeaders(amqp091.Table{mq.HeaderRetryCount: tc.value})
			if got != tc.want {
				t.Fatalf("retryCountFromHeaders() = %d, want %d", got, tc.want)
			}
		})
	}
}

func TestHeadersWithRetryCountCopiesHeaders(t *testing.T) {
	original := amqp091.Table{
		mq.HeaderRequestID:  "req-1",
		mq.HeaderRetryCount: int32(1),
		"other":             "keep",
	}

	copied := headersWithRetryCount(original, 2)
	if copied[mq.HeaderRequestID] != "req-1" {
		t.Fatalf("request header = %v, want req-1", copied[mq.HeaderRequestID])
	}
	if copied["other"] != "keep" {
		t.Fatalf("other header = %v, want keep", copied["other"])
	}
	if copied[mq.HeaderRetryCount] != int32(2) {
		t.Fatalf("retry header = %v, want int32(2)", copied[mq.HeaderRetryCount])
	}
	if original[mq.HeaderRetryCount] != int32(1) {
		t.Fatalf("original retry header mutated to %v", original[mq.HeaderRetryCount])
	}
}

func TestFailedMessageRouteFor(t *testing.T) {
	cases := []struct {
		name       string
		retryCount int
		maxRetries int
		want       failedMessageRoute
	}{
		{
			name:       "schedules retry below max",
			retryCount: 1,
			maxRetries: 3,
			want: failedMessageRoute{
				RoutingKey: mq.SyncRetryRouteKey,
				RetryCount: 2,
			},
		},
		{
			name:       "sends to dlq at max",
			retryCount: 3,
			maxRetries: 3,
			want: failedMessageRoute{
				RoutingKey: mq.SyncDLQRouteKey,
				RetryCount: 3,
				ToDLQ:      true,
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := failedMessageRouteFor(deliveryMeta{RetryCount: tc.retryCount}, tc.maxRetries)
			if got != tc.want {
				t.Fatalf("failedMessageRouteFor() = %#v, want %#v", got, tc.want)
			}
		})
	}
}
