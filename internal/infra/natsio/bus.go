package natsio

import (
	"context"
	"encoding/json"
	"sync"
	"time"
)

// Handler 定义 NATS subject 处理回调。
type Handler func(ctx context.Context, subject string, payload []byte) error

// Bus 抽象了 volta-brain 对外消息边界。
type Bus interface {
	PublishJSON(ctx context.Context, subject string, payload any) error
	Subscribe(ctx context.Context, subject string, handler Handler) error
	Close() error
}

// PublishedMessage 用于测试场景下观测输出。
type PublishedMessage struct {
	Subject   string
	Payload   []byte
	Timestamp time.Time
}

// MemoryBus 是测试用的内存版 NATS 实现。
type MemoryBus struct {
	mu        sync.RWMutex
	handlers  map[string][]Handler
	published []PublishedMessage
	closed    bool
}

// NewMemoryBus 创建一个用于测试的内存消息总线。
func NewMemoryBus() *MemoryBus {
	return &MemoryBus{
		handlers: make(map[string][]Handler),
	}
}

// PublishJSON 将 JSON 消息投递给订阅者。
func (b *MemoryBus) PublishJSON(ctx context.Context, subject string, payload any) error {
	data, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	b.mu.Lock()
	if b.closed {
		b.mu.Unlock()
		return nil
	}
	handlers := append([]Handler(nil), b.handlers[subject]...)
	b.published = append(b.published, PublishedMessage{
		Subject:   subject,
		Payload:   append([]byte(nil), data...),
		Timestamp: time.Now(),
	})
	b.mu.Unlock()

	for _, handler := range handlers {
		if err := handler(ctx, subject, data); err != nil {
			return err
		}
	}
	return nil
}

// Subscribe 注册 subject 订阅回调。
func (b *MemoryBus) Subscribe(_ context.Context, subject string, handler Handler) error {
	b.mu.Lock()
	defer b.mu.Unlock()
	if b.closed {
		return nil
	}
	b.handlers[subject] = append(b.handlers[subject], handler)
	return nil
}

// Close 关闭总线。
func (b *MemoryBus) Close() error {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.closed = true
	return nil
}

// Published 返回所有已发布消息的副本。
func (b *MemoryBus) Published() []PublishedMessage {
	b.mu.RLock()
	defer b.mu.RUnlock()

	copied := make([]PublishedMessage, len(b.published))
	copy(copied, b.published)
	return copied
}
