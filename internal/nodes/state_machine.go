package nodes

import (
	"context"

	"github.com/letianxing/volta-brain/internal/contracts"
	"github.com/letianxing/volta-brain/internal/infra/eventbus"
	"github.com/letianxing/volta-brain/internal/state"
)

// StateMachineNode 根据 winner 更新当前 mode 并发出行为意图。
type StateMachineNode struct {
	name     string
	internal *eventbus.Bus
	store    *state.Store
}

// NewStateMachineNode 创建状态机节点。
func NewStateMachineNode(internal *eventbus.Bus, store *state.Store) *StateMachineNode {
	return &StateMachineNode{
		name:     "state_machine",
		internal: internal,
		store:    store,
	}
}

// Name 返回节点名称。
func (n *StateMachineNode) Name() string { return n.name }

// Start 注册订阅。
func (n *StateMachineNode) Start(ctx context.Context) error {
	return n.internal.Subscribe(ctx, contracts.TopicArbitrationWinner, func(runCtx context.Context, _ string, payload []byte) error {
		decision, err := decode[contracts.ArbitrationDecision](payload)
		if err != nil {
			return err
		}

		nextState := decision.Winner.TargetState
		if nextState == "" {
			nextState = n.store.CurrentState()
		}
		stateSnapshot := n.store.UpdateState(nextState)
		snapshot := n.store.Snapshot()

		stateMachineOutput := contracts.StateMachineOutput{
			TraceID:   decision.Winner.TraceID,
			InputID:   decision.Winner.InputID,
			State:     stateSnapshot,
			Timestamp: now(),
		}
		if err := n.internal.PublishJSON(runCtx, contracts.TopicStateMachineOutput, stateMachineOutput); err != nil {
			return err
		}

		stateOutput := contracts.StateOutput{
			TraceID:   decision.Winner.TraceID,
			InputID:   decision.Winner.InputID,
			State:     stateSnapshot,
			Context:   snapshot.Internal,
			Timestamp: now(),
		}
		if err := n.internal.PublishJSON(runCtx, contracts.TopicStateOutput, stateOutput); err != nil {
			return err
		}

		intent := contracts.BehaviorIntent{
			Decision: decision,
			State:    stateSnapshot,
			TraceID:  decision.Winner.TraceID,
			InputID:  decision.Winner.InputID,
		}
		if err := n.internal.PublishJSON(runCtx, contracts.TopicBehaviorIntent, intent); err != nil {
			return err
		}

		return publishPipeline(runCtx, n.internal, contracts.PipelineEvent{
			Node:    n.name,
			Phase:   "mode_applied",
			TraceID: decision.Winner.TraceID,
			InputID: decision.Winner.InputID,
			Path:    decision.Winner.Path,
			Details: map[string]any{
				"state": stateSnapshot.Current,
			},
		})
	})
}
