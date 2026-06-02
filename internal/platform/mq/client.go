package mq

import (
	"fmt"
	"log/slog"

	"github.com/rabbitmq/amqp091-go"
)

type Client struct {
	conn *amqp091.Connection
	ch   *amqp091.Channel
	cfg  Config
}

func New(cfg Config) (*Client, error) {
	conn, err := amqp091.Dial(cfg.URL)
	if err != nil {
		return nil, fmt.Errorf("dial rabbitmq: %w", err)
	}
	slog.Info("rabbitmq connected")

	ch, err := conn.Channel()
	if err != nil {
		_ = conn.Close()
		return nil, fmt.Errorf("open channel: %w", err)
	}
	client := &Client{conn: conn, ch: ch, cfg: cfg}
	if err := client.declareTopology(); err != nil {
		_ = ch.Close()
		_ = conn.Close()
		return nil, err
	}
	if cfg.PrefetchCount > 0 {
		if err := ch.Qos(cfg.PrefetchCount, 0, false); err != nil {
			_ = ch.Close()
			_ = conn.Close()
			return nil, fmt.Errorf("set qos: %w", err)
		}
	}
	slog.Info(
		"rabbitmq ready",
		"exchange", cfg.Exchange,
		"queue", cfg.Queue,
		"retry_queue", cfg.RetryQueue,
		"dlq", cfg.DLQ,
		"main_routing_key", cfg.MainRouteKey,
		"retry_routing_key", cfg.RetryRouteKey,
		"dlq_routing_key", cfg.DLQRouteKey,
		"prefetch_count", cfg.PrefetchCount,
	)
	return client, nil
}

func (c *Client) Close() error {
	if c.ch != nil {
		_ = c.ch.Close()
	}
	if c.conn != nil {
		return c.conn.Close()
	}
	return nil
}
