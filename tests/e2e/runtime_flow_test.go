package e2e

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/letianxing/volta-brain/internal/config"
	"github.com/letianxing/volta-brain/internal/contracts"
	"github.com/letianxing/volta-brain/internal/infra/natsio"
	"github.com/letianxing/volta-brain/internal/runtime"
)

func TestRuntimeFlow_EmitsPipelineActionStateMachineAndState(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	external := natsio.NewMemoryBus()
	rt := runtime.New(config.Default(), external)
	defer func() {
		_ = rt.Close()
	}()

	if err := rt.Start(ctx); err != nil {
		t.Fatalf("start runtime: %v", err)
	}

	frame := contracts.InputFrame{
		ID:      "input-1",
		TraceID: "trace-1",
		Source:  "test",
		Directives: []contracts.PathDirective{
			{
				Path:         contracts.PathExplicit,
				Active:       true,
				Priority:     10,
				Allow:        false,
				RejectReason: "gate.blocked",
			},
			{
				Path:        contracts.PathReflex,
				Active:      true,
				Priority:    30,
				Allow:       true,
				TargetState: "state.alpha",
				Payload: map[string]any{
					"route": "winner",
				},
			},
			{
				Path:        contracts.PathLLM,
				Active:      true,
				Priority:    20,
				Allow:       true,
				TargetState: "state.beta",
				Payload: map[string]any{
					"route": "loser",
				},
			},
		},
	}

	if err := external.PublishJSON(ctx, contracts.SubjectInputFrame, frame); err != nil {
		t.Fatalf("publish input frame: %v", err)
	}

	waitFor(t, 2*time.Second, func() bool {
		return countPublished(external.Published(), contracts.SubjectAction) >= 2 &&
			countPublished(external.Published(), contracts.SubjectStateMachine) >= 1 &&
			countPublished(external.Published(), contracts.SubjectState) >= 1 &&
			countPublished(external.Published(), contracts.SubjectPipeline) >= 1
	})

	actions := decodePublished[contracts.ActionCommand](t, external.Published(), contracts.SubjectAction)
	if len(actions) < 2 {
		t.Fatalf("expected at least 2 action outputs, got %d", len(actions))
	}

	var accepted *contracts.ActionCommand
	var rejected *contracts.ActionCommand
	for idx := range actions {
		action := actions[idx]
		switch action.Status {
		case "accepted":
			accepted = &action
		case "rejected":
			rejected = &action
		}
	}
	if accepted == nil {
		t.Fatal("expected one accepted action command")
	}
	if rejected == nil {
		t.Fatal("expected one rejected action command")
	}
	if accepted.State != "state.alpha" {
		t.Fatalf("expected accepted action state.alpha, got %q", accepted.State)
	}
	if got := accepted.Payload["route"]; got != "winner" {
		t.Fatalf("expected accepted payload route=winner, got %#v", got)
	}

	stateMachines := decodePublished[contracts.StateMachineOutput](t, external.Published(), contracts.SubjectStateMachine)
	if len(stateMachines) == 0 {
		t.Fatal("expected state machine output")
	}
	if current := stateMachines[len(stateMachines)-1].State.Current; current != "state.alpha" {
		t.Fatalf("expected state.alpha as current state, got %q", current)
	}
}

func TestRuntimeFlow_RejectedOnlyStillProducesActionFeedback(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	external := natsio.NewMemoryBus()
	rt := runtime.New(config.Default(), external)
	defer func() {
		_ = rt.Close()
	}()

	if err := rt.Start(ctx); err != nil {
		t.Fatalf("start runtime: %v", err)
	}

	frame := contracts.InputFrame{
		ID:      "input-rejected",
		TraceID: "trace-rejected",
		Source:  "test",
		Directives: []contracts.PathDirective{
			{
				Path:         contracts.PathExplicit,
				Active:       true,
				Priority:     1,
				Allow:        false,
				RejectReason: "blocked",
			},
		},
	}

	if err := external.PublishJSON(ctx, contracts.SubjectInputFrame, frame); err != nil {
		t.Fatalf("publish rejected-only frame: %v", err)
	}

	waitFor(t, 2*time.Second, func() bool {
		return countPublished(external.Published(), contracts.SubjectAction) >= 1
	})

	actions := decodePublished[contracts.ActionCommand](t, external.Published(), contracts.SubjectAction)
	if len(actions) != 1 {
		t.Fatalf("expected exactly 1 action feedback, got %d", len(actions))
	}
	if actions[0].Status != "rejected" {
		t.Fatalf("expected rejected action, got %q", actions[0].Status)
	}
	if countPublished(external.Published(), contracts.SubjectStateMachine) != 0 {
		t.Fatal("did not expect state machine output for rejected-only flow")
	}
}

func waitFor(t *testing.T, timeout time.Duration, predicate func() bool) {
	t.Helper()
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		if predicate() {
			return
		}
		time.Sleep(10 * time.Millisecond)
	}
	t.Fatal("timed out waiting for condition")
}

func countPublished(messages []natsio.PublishedMessage, subject string) int {
	count := 0
	for _, message := range messages {
		if message.Subject == subject {
			count++
		}
	}
	return count
}

func decodePublished[T any](t *testing.T, messages []natsio.PublishedMessage, subject string) []T {
	t.Helper()

	result := make([]T, 0)
	for _, message := range messages {
		if message.Subject != subject {
			continue
		}
		var value T
		if err := json.Unmarshal(message.Payload, &value); err != nil {
			t.Fatalf("decode published payload for %s: %v", subject, err)
		}
		result = append(result, value)
	}
	return result
}
