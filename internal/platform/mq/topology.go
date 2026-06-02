package mq

import (
	"fmt"
	"log/slog"

	"github.com/rabbitmq/amqp091-go"
)

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
