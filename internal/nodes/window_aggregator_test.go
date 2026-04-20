package nodes

import (
	"context"
	"testing"
	"time"

	"github.com/letianxing/volta-brain/internal/contracts"
	"github.com/letianxing/volta-brain/internal/infra/eventbus"
)

func TestWindowAggregatorNodeAggregatesWithinWindow(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	bus := eventbus.New(32)
	defer func() {
		_ = bus.Close()
	}()

	node := NewWindowAggregatorNode(bus, 30*time.Millisecond)
	if err := node.Start(ctx); err != nil {
		t.Fatalf("start node: %v", err)
	}

	ready := make(chan contracts.WindowEnvelope, 1)
	if err := bus.Subscribe(ctx, contracts.TopicWindowReady, func(_ context.Context, _ string, payload []byte) error {
		envelope, err := decode[contracts.WindowEnvelope](payload)
		if err != nil {
			return err
		}
		ready <- envelope
		return nil
	}); err != nil {
		t.Fatalf("subscribe window ready: %v", err)
	}

	first := contracts.CandidateEvent{ID: "a", TraceID: "trace-1", InputID: "input-1", Priority: 10, Timestamp: now()}
	second := contracts.CandidateEvent{ID: "b", TraceID: "trace-1", InputID: "input-1", Priority: 20, Timestamp: now()}

	if err := bus.PublishJSON(ctx, contracts.TopicCandidateEvent, first); err != nil {
		t.Fatalf("publish first: %v", err)
	}
	if err := bus.PublishJSON(ctx, contracts.TopicCandidateEvent, second); err != nil {
		t.Fatalf("publish second: %v", err)
	}

	select {
	case envelope := <-ready:
		if got := len(envelope.Events); got != 2 {
			t.Fatalf("expected 2 events in window, got %d", got)
		}
	case <-time.After(time.Second):
		t.Fatal("timed out waiting for aggregated window")
	}
}
