package syncer

import (
	"context"
	"encoding/json"
	"log/slog"

	"github.com/rabbitmq/amqp091-go"
	"github.com/tranphuocnhan/radio-shuffle/internal/platform/mq"
)

type publisher struct {
	client   *mq.Client
	routeKey string
}

func NewPublisher(client *mq.Client, routeKey string) JobPublisher {
	return &publisher{client: client, routeKey: routeKey}
}

func (p *publisher) PublishSyncCommand(ctx context.Context, cmd SyncCommand) error {
	msg := toSyncRequestedMessage(cmd)
	body, err := json.Marshal(msg)
	if err != nil {
		return err
	}
	headers := amqp091.Table{
		mq.HeaderRequestID:  cmd.RequestID,
		mq.HeaderRetryCount: int32(0),
	}
	slog.Info(
		"rabbitmq publish sync requested",
		"request_id", cmd.RequestID,
		"requested_by", cmd.RequestedBy,
		"routing_key", p.routeKey,
	)
	if err := p.client.Publish(ctx, p.routeKey, body, headers); err != nil {
		slog.Error(
			"rabbitmq publish sync failed",
			"request_id", cmd.RequestID,
			"requested_by", cmd.RequestedBy,
			"routing_key", p.routeKey,
			"err", err,
		)
		return err
	}
	slog.Info(
		"rabbitmq publish sync completed",
		"request_id", cmd.RequestID,
		"requested_by", cmd.RequestedBy,
		"routing_key", p.routeKey,
	)
	return nil
}
