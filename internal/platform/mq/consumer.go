package mq

import "github.com/rabbitmq/amqp091-go"

func (c *Client) Consume(consumer string) (<-chan amqp091.Delivery, error) {
	return c.ch.Consume(c.cfg.Queue, consumer, false, false, false, false, nil)
}
