package nodes

import (
	"context"
	"encoding/json"

	"github.com/letianxing/volta-brain/internal/contracts"
	"github.com/letianxing/volta-brain/internal/infra/eventbus"
	"github.com/letianxing/volta-brain/internal/infra/natsio"
)

// NATSOutputNode 将内部四类输出转发到外部 NATS subject。
type NATSOutputNode struct {
	name     string
	internal *eventbus.Bus
	external natsio.Bus
	subjects map[string]string
}

// NewNATSOutputNode 创建输出边界节点。
func NewNATSOutputNode(external natsio.Bus, internal *eventbus.Bus, subjects map[string]string) *NATSOutputNode {
	return &NATSOutputNode{
		name:     "nats_output",
		internal: internal,
		external: external,
		subjects: subjects,
	}
}

// Name 返回节点名称。
func (n *NATSOutputNode) Name() string { return n.name }

// Start 注册内部输出订阅。
func (n *NATSOutputNode) Start(ctx context.Context) error {
	for topic, subject := range n.subjects {
		topic := topic
		subject := subject
		if err := n.internal.Subscribe(ctx, topic, func(runCtx context.Context, _ string, payload []byte) error {
			if err := n.external.PublishJSON(runCtx, subject, json.RawMessage(payload)); err != nil {
				return err
			}
			traceID, inputID := traceKeys(topic, payload)
			return publishPipeline(runCtx, n.internal, contracts.PipelineEvent{
				Node:    n.name,
				Phase:   "egress",
				TraceID: traceID,
				InputID: inputID,
				Details: map[string]any{
					"topic":   topic,
					"subject": subject,
				},
			})
		}); err != nil {
			return err
		}
	}
	return nil
}

func traceKeys(topic string, payload []byte) (string, string) {
	switch topic {
	case contracts.TopicBehaviorOutput:
		value, err := decode[contracts.ActionCommand](payload)
		if err != nil {
			return "", ""
		}
		return value.TraceID, value.InputID
	case contracts.TopicStateMachineOutput:
		value, err := decode[contracts.StateMachineOutput](payload)
		if err != nil {
			return "", ""
		}
		return value.TraceID, value.InputID
	case contracts.TopicStateOutput:
		value, err := decode[contracts.StateOutput](payload)
		if err != nil {
			return "", ""
		}
		return value.TraceID, value.InputID
	case contracts.TopicPipeline:
		value, err := decode[contracts.PipelineEvent](payload)
		if err != nil {
			return "", ""
		}
		return value.TraceID, value.InputID
	default:
		return "", ""
	}
}
