package nodes

import (
	"context"
	"testing"
	"time"

	"github.com/letianxing/volta-brain/internal/contracts"
	"github.com/letianxing/volta-brain/internal/infra/eventbus"
)

func TestArbiterNodeSelectsHighestPriorityThenEarliest(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	bus := eventbus.New(32)
	defer func() {
		_ = bus.Close()
	}()

	node := NewArbiterNode(bus)
	if err := node.Start(ctx); err != nil {
		t.Fatalf("start node: %v", err)
	}

	winnerCh := make(chan contracts.ArbitrationDecision, 1)
	if err := bus.Subscribe(ctx, contracts.TopicArbitrationWinner, func(_ context.Context, _ string, payload []byte) error {
		decision, err := decode[contracts.ArbitrationDecision](payload)
		if err != nil {
			return err
		}
		winnerCh <- decision
		return nil
	}); err != nil {
		t.Fatalf("subscribe winner: %v", err)
	}

	base := now()
	envelope := contracts.GateAcceptedEnvelope{
		WindowID: "window-1",
		Accepted: []contracts.CandidateEvent{
			{ID: "low", TraceID: "trace-1", InputID: "input-1", Priority: 10, Timestamp: base.Add(2 * time.Millisecond)},
			{ID: "high-late", TraceID: "trace-1", InputID: "input-1", Priority: 20, Timestamp: base.Add(3 * time.Millisecond)},
			{ID: "high-early", TraceID: "trace-1", InputID: "input-1", Priority: 20, Timestamp: base},
		},
		Timestamp: now(),
	}

	if err := bus.PublishJSON(ctx, contracts.TopicGateAccepted, envelope); err != nil {
		t.Fatalf("publish accepted: %v", err)
	}

	select {
	case decision := <-winnerCh:
		if decision.Winner.ID != "high-early" {
			t.Fatalf("expected high-early to win, got %s", decision.Winner.ID)
		}
		if got := len(decision.Rejected); got != 2 {
			t.Fatalf("expected 2 rejected events, got %d", got)
		}
	case <-time.After(time.Second):
		t.Fatal("timed out waiting for arbitration decision")
	}
}
