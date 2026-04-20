package nodes

import (
	"context"

	"github.com/letianxing/volta-brain/internal/contracts"
	"github.com/letianxing/volta-brain/internal/infra/eventbus"
	"github.com/letianxing/volta-brain/internal/state"
)

// PersonalizationNode 保留图中的“个性化/策略”层。
type PersonalizationNode struct {
	name     string
	internal *eventbus.Bus
	store    *state.Store
}

// NewPersonalizationNode 创建个性化节点。
func NewPersonalizationNode(internal *eventbus.Bus, store *state.Store) *PersonalizationNode {
	return &PersonalizationNode{
		name:     "personalization",
		internal: internal,
		store:    store,
	}
}

// Name 返回节点名称。
func (n *PersonalizationNode) Name() string { return n.name }

// Start 注册订阅。
func (n *PersonalizationNode) Start(ctx context.Context) error {
	return n.internal.Subscribe(ctx, contracts.TopicBehaviorIntent, func(runCtx context.Context, _ string, payload []byte) error {
		intent, err := decode[contracts.BehaviorIntent](payload)
		if err != nil {
			return err
		}

		personality := n.store.Snapshot().Personality

		ready := contracts.BehaviorReady{
			Intent:      intent,
			Personality: personality,
		}
		if err := n.internal.PublishJSON(runCtx, contracts.TopicBehaviorReady, ready); err != nil {
			return err
		}
		return publishPipeline(runCtx, n.internal, contracts.PipelineEvent{
			Node:    n.name,
			Phase:   "behavior_prepared",
			TraceID: intent.TraceID,
			InputID: intent.InputID,
			Path:    intent.Decision.Winner.Path,
		})
	})
}
