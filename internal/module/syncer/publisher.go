package syncer

import (
	"context"
	"encoding/json"

	"github.com/rabbitmq/amqp091-go"
	"github.com/tranphuocnhan/radio-shuffle/internal/platform/mq"
)

type publisher struct {
	client    *mq.Client
	routeKey  string
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
		"x-request-id": cmd.RequestID,
		"x-retry-count": int32(0),
	}
	return p.client.Publish(ctx, p.routeKey, body, headers)
}
