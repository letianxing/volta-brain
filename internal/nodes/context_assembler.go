package nodes

import (
	"context"

	"github.com/letianxing/volta-brain/internal/contracts"
	"github.com/letianxing/volta-brain/internal/infra/eventbus"
	"github.com/letianxing/volta-brain/internal/state"
)

// ContextAssemblerNode 给输入补齐 personality/memory/internal-state 上下文。
type ContextAssemblerNode struct {
	name     string
	internal *eventbus.Bus
	store    *state.Store
}

// NewContextAssemblerNode 创建上下文组装节点。
func NewContextAssemblerNode(internal *eventbus.Bus, store *state.Store) *ContextAssemblerNode {
	return &ContextAssemblerNode{
		name:     "context_assembler",
		internal: internal,
		store:    store,
	}
}

// Name 返回节点名称。
func (n *ContextAssemblerNode) Name() string { return n.name }

// Start 注册订阅。
func (n *ContextAssemblerNode) Start(ctx context.Context) error {
	return n.internal.Subscribe(ctx, contracts.TopicInputIngress, func(runCtx context.Context, _ string, payload []byte) error {
		frame, err := decode[contracts.InputFrame](payload)
		if err != nil {
			return err
		}

		contextualized := contracts.ContextualFrame{
			Frame:   frame,
			Context: n.store.Snapshot(),
		}
		if err := n.internal.PublishJSON(runCtx, contracts.TopicInputContext, contextualized); err != nil {
			return err
		}
		return publishPipeline(runCtx, n.internal, contracts.PipelineEvent{
			Node:    n.name,
			Phase:   "contextualized",
			TraceID: frame.TraceID,
			InputID: frame.ID,
		})
	})
}
