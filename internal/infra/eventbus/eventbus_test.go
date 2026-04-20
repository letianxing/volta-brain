package eventbus

import (
	"context"
	"testing"
	"time"
)

func TestBusPublishAndSubscribe(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	bus := New(16)
	defer func() {
		_ = bus.Close()
	}()

	received := make(chan string, 1)
	if err := bus.Subscribe(ctx, "topic.test", func(_ context.Context, _ string, payload []byte) error {
		received <- string(payload)
		return nil
	}); err != nil {
		t.Fatalf("subscribe: %v", err)
	}

	if err := bus.PublishJSON(ctx, "topic.test", map[string]string{"ok": "yes"}); err != nil {
		t.Fatalf("publish: %v", err)
	}

	select {
	case value := <-received:
		if value == "" {
			t.Fatal("expected payload")
		}
	case <-time.After(time.Second):
		t.Fatal("timed out waiting for eventbus delivery")
	}
}
