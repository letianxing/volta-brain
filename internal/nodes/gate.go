package nodes

import (
	"context"

	"github.com/letianxing/volta-brain/internal/contracts"
	"github.com/letianxing/volta-brain/internal/infra/eventbus"
)

// GateNode 负责资格门拆分 accepted/rejected。
type GateNode struct {
	name     string
	internal *eventbus.Bus
}

// NewGateNode 创建资格门节点。
func NewGateNode(internal *eventbus.Bus) *GateNode {
	return &GateNode{
		name:     "gate",
		internal: internal,
	}
}

// Name 返回节点名称。
func (n *GateNode) Name() string { return n.name }

// Start 注册订阅。
func (n *GateNode) Start(ctx context.Context) error {
	return n.internal.Subscribe(ctx, contracts.TopicWindowReady, func(runCtx context.Context, _ string, payload []byte) error {
		window, err := decode[contracts.WindowEnvelope](payload)
		if err != nil {
			return err
		}

		accepted := make([]contracts.CandidateEvent, 0, len(window.Events))
		rejected := make([]contracts.RejectedEvent, 0)
		for _, event := range window.Events {
			if event.Allow {
				accepted = append(accepted, event)
				continue
			}
			rejected = append(rejected, contracts.RejectedEvent{
				Event:  event,
				Reason: event.RejectReason,
			})
		}

		if len(accepted) > 0 {
			if err := n.internal.PublishJSON(runCtx, contracts.TopicGateAccepted, contracts.GateAcceptedEnvelope{
				WindowID:  window.WindowID,
				Accepted:  accepted,
				Timestamp: now(),
			}); err != nil {
				return err
			}
		}

		if len(rejected) > 0 {
			if err := n.internal.PublishJSON(runCtx, contracts.TopicGateRejected, contracts.GateRejectedEnvelope{
				WindowID:  window.WindowID,
				Rejected:  rejected,
				Timestamp: now(),
			}); err != nil {
				return err
			}
		}

		traceID := ""
		inputID := ""
		if len(window.Events) > 0 {
			traceID = window.Events[0].TraceID
			inputID = window.Events[0].InputID
		}
		return publishPipeline(runCtx, n.internal, contracts.PipelineEvent{
			Node:    n.name,
			Phase:   "qualified",
			TraceID: traceID,
			InputID: inputID,
			Details: map[string]any{
				"accepted": len(accepted),
				"rejected": len(rejected),
			},
		})
	})
}
