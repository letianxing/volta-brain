package eventbus

import (
	"context"
	"encoding/json"
	"errors"
	"sync"

	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/ThreeDotsLabs/watermill/pubsub/gochannel"
)

// Handler 定义 eventbus 订阅回调。
type Handler func(ctx context.Context, topic string, payload []byte) error

// Bus 是基于 Watermill GoChannel 的进程内事件总线。
type Bus struct {
	pubsub *gochannel.GoChannel
	pub    message.Publisher
	sub    message.Subscriber

	mu     sync.RWMutex
	closed bool
}

// New 创建一个新的内部事件总线。
func New(buffer int64) *Bus {
	logger := watermill.NewStdLogger(false, false)
	pubsub := gochannel.NewGoChannel(gochannel.Config{
		OutputChannelBuffer:            buffer,
		Persistent:                     false,
		BlockPublishUntilSubscriberAck: false,
	}, logger)

	return &Bus{
		pubsub: pubsub,
		pub:    pubsub,
		sub:    pubsub,
	}
}

// PublishJSON 将结构体编码为 JSON 后发布到 topic。
func (b *Bus) PublishJSON(ctx context.Context, topic string, payload any) error {
	b.mu.RLock()
	defer b.mu.RUnlock()

	if b.closed {
		return errors.New("eventbus closed")
	}

	data, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	msg := message.NewMessage(watermill.NewUUID(), data)
	return b.pub.Publish(topic, msg)
}

// Subscribe 注册一个 topic 处理器。
func (b *Bus) Subscribe(ctx context.Context, topic string, handler Handler) error {
	b.mu.RLock()
	defer b.mu.RUnlock()

	if b.closed {
		return errors.New("eventbus closed")
	}

	ch, err := b.sub.Subscribe(ctx, topic)
	if err != nil {
		return err
	}

	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case msg, ok := <-ch:
				if !ok {
					return
				}
				if err := handler(ctx, topic, msg.Payload); err != nil {
					msg.Nack()
					continue
				}
				msg.Ack()
			}
		}
	}()

	return nil
}

// Close 关闭事件总线。
func (b *Bus) Close() error {
	b.mu.Lock()
	defer b.mu.Unlock()

	if b.closed {
		return nil
	}
	b.closed = true
	return b.pubsub.Close()
}
