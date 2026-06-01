package mq

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/rabbitmq/amqp091-go"
)

type Config struct {
	URL           string
	Exchange      string
	Queue         string
	RetryQueue    string
	DLQ           string
	RetryTTLMS    int
	MaxRetries    int
	RetryRouteKey string
	MainRouteKey  string
	DLQRouteKey   string
	PrefetchCount int
}

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

func (c *Client) Publish(ctx context.Context, routingKey string, body []byte, headers amqp091.Table) error {
	msg := amqp091.Publishing{
		ContentType: "application/json",
		Body:        body,
		Headers:     headers,
	}
	return c.ch.PublishWithContext(ctx, c.cfg.Exchange, routingKey, false, false, msg)
}

func (c *Client) Consume(consumer string) (<-chan amqp091.Delivery, error) {
	return c.ch.Consume(c.cfg.Queue, consumer, false, false, false, false, nil)
}

func (c *Client) declareTopology() error {
	cfg := c.cfg
	if err := c.ch.ExchangeDeclare(cfg.Exchange, "direct", true, false, false, false, nil); err != nil {
		return fmt.Errorf("declare exchange: %w", err)
	}
	if _, err := c.ch.QueueDeclare(cfg.Queue, true, false, false, false, nil); err != nil {
		return fmt.Errorf("declare queue: %w", err)
	}
	if _, err := c.ch.QueueDeclare(cfg.DLQ, true, false, false, false, nil); err != nil {
		return fmt.Errorf("declare dlq: %w", err)
	}
	retryArgs := amqp091.Table{
		"x-message-ttl":             cfg.RetryTTLMS,
		"x-dead-letter-exchange":    cfg.Exchange,
		"x-dead-letter-routing-key": cfg.MainRouteKey,
	}
	if _, err := c.ch.QueueDeclare(cfg.RetryQueue, true, false, false, false, retryArgs); err != nil {
		return fmt.Errorf("declare retry queue: %w", err)
	}
	if err := c.ch.QueueBind(cfg.Queue, cfg.MainRouteKey, cfg.Exchange, false, nil); err != nil {
		return fmt.Errorf("bind main queue: %w", err)
	}
	if err := c.ch.QueueBind(cfg.RetryQueue, cfg.RetryRouteKey, cfg.Exchange, false, nil); err != nil {
		return fmt.Errorf("bind retry queue: %w", err)
	}
	if err := c.ch.QueueBind(cfg.DLQ, cfg.DLQRouteKey, cfg.Exchange, false, nil); err != nil {
		return fmt.Errorf("bind dlq: %w", err)
	}
	slog.Info(
		"rabbitmq topology declared",
		"exchange", cfg.Exchange,
		"queue", cfg.Queue,
		"retry_queue", cfg.RetryQueue,
		"dlq", cfg.DLQ,
		"retry_ttl_ms", cfg.RetryTTLMS,
	)
	return nil
}
