package rabbitmq

import (
	"context"
	"fmt"
	"sync"

	amqp "github.com/rabbitmq/amqp091-go"
)

type Client struct {
	url string
	mu  sync.Mutex
	conn *amqp.Connection
	ch   *amqp.Channel
}

func Open(url string) (*Client, error) {
	if url == "" {
		return nil, fmt.Errorf("rabbitmq url is empty")
	}
	c := &Client{url: url}
	if err := c.connect(); err != nil {
		return nil, err
	}
	return c, nil
}

func (c *Client) connect() error {
	conn, err := amqp.Dial(c.url)
	if err != nil {
		return fmt.Errorf("amqp dial: %w", err)
	}
	ch, err := conn.Channel()
	if err != nil {
		_ = conn.Close()
		return fmt.Errorf("amqp channel: %w", err)
	}
	c.conn = conn
	c.ch = ch
	return nil
}

func (c *Client) Ping(ctx context.Context) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.conn == nil || c.conn.IsClosed() {
		if err := c.connect(); err != nil {
			return err
		}
	}
	return nil
}

func (c *Client) DeclareQueues(queues []string) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	for _, q := range queues {
		if _, err := c.ch.QueueDeclare(q, true, false, false, false, nil); err != nil {
			return fmt.Errorf("declare queue %s: %w", q, err)
		}
		dlq := q + ".dlq"
		if _, err := c.ch.QueueDeclare(dlq, true, false, false, false, nil); err != nil {
			return fmt.Errorf("declare dlq %s: %w", dlq, err)
		}
	}
	return nil
}

type JobMessage struct {
	CreativeRunID string `json:"creativeRunId"`
	StepID        string `json:"stepId"`
	StepType      string `json:"stepType"`
	Attempt       int    `json:"attempt"`
}

func (c *Client) Publish(ctx context.Context, queue string, body []byte) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.conn == nil || c.conn.IsClosed() {
		if err := c.connect(); err != nil {
			return err
		}
	}
	return c.ch.PublishWithContext(ctx, "", queue, false, false, amqp.Publishing{
		ContentType:  "application/json",
		DeliveryMode: amqp.Persistent,
		Body:         body,
	})
}

func (c *Client) Consume(queue string, handler func(amqp.Delivery) error) error {
	c.mu.Lock()
	if c.conn == nil || c.conn.IsClosed() {
		if err := c.connect(); err != nil {
			c.mu.Unlock()
			return err
		}
	}
	deliveries, err := c.ch.Consume(queue, "", false, false, false, false, nil)
	c.mu.Unlock()
	if err != nil {
		return fmt.Errorf("consume %s: %w", queue, err)
	}
	go func() {
		for d := range deliveries {
			if err := handler(d); err != nil {
				_ = d.Nack(false, true)
				continue
			}
			_ = d.Ack(false)
		}
	}()
	return nil
}

func (c *Client) Close() error {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.ch != nil {
		_ = c.ch.Close()
	}
	if c.conn != nil {
		return c.conn.Close()
	}
	return nil
}
