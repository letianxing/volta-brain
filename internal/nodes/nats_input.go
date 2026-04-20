package nodes

import (
	"context"

	"github.com/letianxing/volta-brain/internal/contracts"
	"github.com/letianxing/volta-brain/internal/infra/eventbus"
	"github.com/letianxing/volta-brain/internal/infra/natsio"
)

// NATSInputNode 将外部 NATS 输入转换成内部 input.ingress 事件。
type NATSInputNode struct {
	name     string
	subject  string
	internal *eventbus.Bus
	external natsio.Bus
}

// NewNATSInputNode 创建输入边界节点。
func NewNATSInputNode(subject string, external natsio.Bus, internal *eventbus.Bus) *NATSInputNode {
	return &NATSInputNode{
		name:     "nats_input",
		subject:  subject,
		internal: internal,
		external: external,
	}
}

// Name 返回节点名称。
func (n *NATSInputNode) Name() string { return n.name }

// Start 启动 NATS 订阅。
func (n *NATSInputNode) Start(ctx context.Context) error {
	return n.external.Subscribe(ctx, n.subject, func(runCtx context.Context, _ string, payload []byte) error {
		frame, err := decode[contracts.InputFrame](payload)
		if err != nil {
			return err
		}
		if frame.ID == "" {
			frame.ID = nextID("input")
		}
		if frame.TraceID == "" {
			frame.TraceID = frame.ID
		}
		if frame.Timestamp.IsZero() {
			frame.Timestamp = now()
		}
		if err := n.internal.PublishJSON(runCtx, contracts.TopicInputIngress, frame); err != nil {
			return err
		}
		return publishPipeline(runCtx, n.internal, contracts.PipelineEvent{
			Node:    n.name,
			Phase:   "ingress",
			TraceID: frame.TraceID,
			InputID: frame.ID,
			Details: map[string]any{
				"directives": len(frame.Directives),
			},
		})
	})
}
