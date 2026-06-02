package mq

import (
	"context"

	"github.com/rabbitmq/amqp091-go"
)

func (c *Client) Publish(ctx context.Context, routingKey string, body []byte, headers amqp091.Table) error {
	msg := amqp091.Publishing{
		ContentType: "application/json",
		Body:        body,
		Headers:     headers,
	}
	return c.ch.PublishWithContext(ctx, c.cfg.Exchange, routingKey, false, false, msg)
}
