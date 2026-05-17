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

func (p *publisher) PublishSyncRequest(ctx context.Context, req SyncRequest) error {
	body, err := json.Marshal(req)
	if err != nil {
		return err
	}
	headers := amqp091.Table{
		"x-request-id": req.RequestID,
		"x-retry-count": int32(0),
	}
	return p.client.Publish(ctx, p.routeKey, body, headers)
}

