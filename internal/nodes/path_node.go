package nodes

import (
	"context"

	"github.com/letianxing/volta-brain/internal/contracts"
	"github.com/letianxing/volta-brain/internal/infra/eventbus"
)

// PathNode 是 explicit/reflex/llm 三路径的统一骨架实现。
type PathNode struct {
	name       string
	path       contracts.Path
	inputTopic string
	internal   *eventbus.Bus
}

// NewPathNode 创建路径节点。
func NewPathNode(name string, path contracts.Path, inputTopic string, internal *eventbus.Bus) *PathNode {
	return &PathNode{
		name:       name,
		path:       path,
		inputTopic: inputTopic,
		internal:   internal,
	}
}

// Name 返回节点名称。
func (n *PathNode) Name() string { return n.name }

// Start 注册订阅。
func (n *PathNode) Start(ctx context.Context) error {
	return n.internal.Subscribe(ctx, n.inputTopic, func(runCtx context.Context, _ string, payload []byte) error {
		work, err := decode[contracts.PathWorkItem](payload)
		if err != nil {
			return err
		}

		candidate := contracts.CandidateEvent{
			ID:           nextID("event"),
			TraceID:      work.Frame.Frame.TraceID,
			InputID:      work.Frame.Frame.ID,
			Path:         n.path,
			Priority:     work.Directive.Priority,
			Allow:        work.Directive.Allow,
			TargetState:  work.Directive.TargetState,
			RejectReason: work.Directive.RejectReason,
			Context:      work.Frame.Context,
			Payload:      work.Directive.Payload,
			Metadata:     mergeMetadata(work.Frame.Frame.Metadata, work.Directive.Metadata),
			Timestamp:    now(),
		}

		if err := n.internal.PublishJSON(runCtx, contracts.TopicCandidateEvent, candidate); err != nil {
			return err
		}
		return publishPipeline(runCtx, n.internal, contracts.PipelineEvent{
			Node:    n.name,
			Phase:   "candidate_emitted",
			TraceID: candidate.TraceID,
			InputID: candidate.InputID,
			Path:    candidate.Path,
		})
	})
}
