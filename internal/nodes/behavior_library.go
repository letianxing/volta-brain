package nodes

import (
	"context"

	"github.com/letianxing/volta-brain/internal/contracts"
	"github.com/letianxing/volta-brain/internal/infra/eventbus"
)

// BehaviorLibraryNode 输出对外可执行的行为命令。
type BehaviorLibraryNode struct {
	name     string
	internal *eventbus.Bus
}

// NewBehaviorLibraryNode 创建行为库节点。
func NewBehaviorLibraryNode(internal *eventbus.Bus) *BehaviorLibraryNode {
	return &BehaviorLibraryNode{
		name:     "behavior_library",
		internal: internal,
	}
}

// Name 返回节点名称。
func (n *BehaviorLibraryNode) Name() string { return n.name }

// Start 注册 accepted/rejected 两类输入。
func (n *BehaviorLibraryNode) Start(ctx context.Context) error {
	if err := n.internal.Subscribe(ctx, contracts.TopicBehaviorReady, n.handleBehaviorReady); err != nil {
		return err
	}
	return n.internal.Subscribe(ctx, contracts.TopicGateRejected, n.handleRejected)
}

func (n *BehaviorLibraryNode) handleBehaviorReady(ctx context.Context, _ string, payload []byte) error {
	ready, err := decode[contracts.BehaviorReady](payload)
	if err != nil {
		return err
	}

	command := contracts.ActionCommand{
		ID:                nextID("cmd"),
		TraceID:           ready.Intent.TraceID,
		InputID:           ready.Intent.InputID,
		SourceCandidateID: ready.Intent.Decision.Winner.ID,
		Status:            "accepted",
		Path:              ready.Intent.Decision.Winner.Path,
		State:             ready.Intent.State.Current,
		Payload:           ready.Intent.Decision.Winner.Payload,
		GeneratedAt:       now(),
	}
	if err := n.internal.PublishJSON(ctx, contracts.TopicBehaviorOutput, command); err != nil {
		return err
	}
	return publishPipeline(ctx, n.internal, contracts.PipelineEvent{
		Node:    n.name,
		Phase:   "behavior_output",
		TraceID: command.TraceID,
		InputID: command.InputID,
		Path:    command.Path,
	})
}

func (n *BehaviorLibraryNode) handleRejected(ctx context.Context, _ string, payload []byte) error {
	envelope, err := decode[contracts.GateRejectedEnvelope](payload)
	if err != nil {
		return err
	}

	for _, rejected := range envelope.Rejected {
		command := contracts.ActionCommand{
			ID:                nextID("cmd"),
			TraceID:           rejected.Event.TraceID,
			InputID:           rejected.Event.InputID,
			SourceCandidateID: rejected.Event.ID,
			Status:            "rejected",
			Path:              rejected.Event.Path,
			State:             "",
			GeneratedAt:       now(),
			Metadata: map[string]string{
				"reject_reason": rejected.Reason,
			},
		}
		if err := n.internal.PublishJSON(ctx, contracts.TopicBehaviorOutput, command); err != nil {
			return err
		}
		if err := publishPipeline(ctx, n.internal, contracts.PipelineEvent{
			Node:    n.name,
			Phase:   "rejection_feedback",
			TraceID: command.TraceID,
			InputID: command.InputID,
			Path:    command.Path,
		}); err != nil {
			return err
		}
	}

	return nil
}
