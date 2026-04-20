package natsio

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/nats-io/nats.go"
)

// Config 描述真实 NATS 客户端配置。
type Config struct {
	URL  string
	Name string
}

// Client 是 `nats.go` 的轻量封装。
type Client struct {
	conn *nats.Conn
}

// NewClient 创建真实 NATS 客户端。
func NewClient(cfg Config) (*Client, error) {
	conn, err := nats.Connect(cfg.URL, nats.Name(cfg.Name))
	if err != nil {
		return nil, err
	}
	return &Client{conn: conn}, nil
}

// PublishJSON 发布 JSON 消息。
func (c *Client) PublishJSON(_ context.Context, subject string, payload any) error {
	data, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	return c.conn.Publish(subject, data)
}

// Subscribe 订阅 NATS subject。
func (c *Client) Subscribe(ctx context.Context, subject string, handler Handler) error {
	_, err := c.conn.Subscribe(subject, func(msg *nats.Msg) {
		if err := handler(ctx, subject, msg.Data); err == nil {
			return
		}
	})
	return err
}

// Close 关闭客户端。
func (c *Client) Close() error {
	if c.conn == nil {
		return nil
	}
	c.conn.Close()
	if err := c.conn.LastError(); err != nil {
		return fmt.Errorf("close nats client: %w", err)
	}
	return nil
}
