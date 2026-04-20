package nodes

import (
	"context"
	"sort"

	"github.com/letianxing/volta-brain/internal/contracts"
	"github.com/letianxing/volta-brain/internal/infra/eventbus"
)

// ArbiterNode 在 accepted 候选事件中选出唯一 winner。
type ArbiterNode struct {
	name     string
	internal *eventbus.Bus
}

// NewArbiterNode 创建仲裁器节点。
func NewArbiterNode(internal *eventbus.Bus) *ArbiterNode {
	return &ArbiterNode{
		name:     "arbiter",
		internal: internal,
	}
}

// Name 返回节点名称。
func (n *ArbiterNode) Name() string { return n.name }

// Start 注册订阅。
func (n *ArbiterNode) Start(ctx context.Context) error {
	return n.internal.Subscribe(ctx, contracts.TopicGateAccepted, func(runCtx context.Context, _ string, payload []byte) error {
		envelope, err := decode[contracts.GateAcceptedEnvelope](payload)
		if err != nil {
			return err
		}
		if len(envelope.Accepted) == 0 {
			return nil
		}

		candidates := append([]contracts.CandidateEvent(nil), envelope.Accepted...)
		sort.SliceStable(candidates, func(i, j int) bool {
			if candidates[i].Priority != candidates[j].Priority {
				return candidates[i].Priority > candidates[j].Priority
			}
			return candidates[i].Timestamp.Before(candidates[j].Timestamp)
		})

		decision := contracts.ArbitrationDecision{
			Winner:    candidates[0],
			Rejected:  candidates[1:],
			Reason:    "priority_then_timestamp",
			Timestamp: now(),
		}
		if err := n.internal.PublishJSON(runCtx, contracts.TopicArbitrationWinner, decision); err != nil {
			return err
		}
		return publishPipeline(runCtx, n.internal, contracts.PipelineEvent{
			Node:    n.name,
			Phase:   "winner_selected",
			TraceID: decision.Winner.TraceID,
			InputID: decision.Winner.InputID,
			Path:    decision.Winner.Path,
			Details: map[string]any{
				"rejected": len(decision.Rejected),
			},
		})
	})
}
