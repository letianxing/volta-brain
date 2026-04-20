package nodes

import (
	"context"
	"testing"
	"time"

	"github.com/letianxing/volta-brain/internal/contracts"
	"github.com/letianxing/volta-brain/internal/infra/eventbus"
	"github.com/letianxing/volta-brain/internal/state"
)

func TestStateMachineNodeAppliesGenericTargetState(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	bus := eventbus.New(32)
	defer func() {
		_ = bus.Close()
	}()

	store := state.NewStore()
	node := NewStateMachineNode(bus, store)
	if err := node.Start(ctx); err != nil {
		t.Fatalf("start node: %v", err)
	}

	outputCh := make(chan contracts.StateMachineOutput, 1)
	if err := bus.Subscribe(ctx, contracts.TopicStateMachineOutput, func(_ context.Context, _ string, payload []byte) error {
		output, err := decode[contracts.StateMachineOutput](payload)
		if err != nil {
			return err
		}
		outputCh <- output
		return nil
	}); err != nil {
		t.Fatalf("subscribe state machine output: %v", err)
	}

	decision := contracts.ArbitrationDecision{
		Winner: contracts.CandidateEvent{
			ID:          "winner-1",
			TraceID:     "trace-1",
			InputID:     "input-1",
			Priority:    100,
			TargetState: "state.alpha",
			Timestamp:   now(),
		},
		Timestamp: now(),
	}
	if err := bus.PublishJSON(ctx, contracts.TopicArbitrationWinner, decision); err != nil {
		t.Fatalf("publish decision: %v", err)
	}

	select {
	case output := <-outputCh:
		if output.State.Current != "state.alpha" {
			t.Fatalf("expected state.alpha, got %q", output.State.Current)
		}
	case <-time.After(time.Second):
		t.Fatal("timed out waiting for state machine output")
	}
}
