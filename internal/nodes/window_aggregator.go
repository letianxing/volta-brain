package nodes

import (
	"context"
	"sync"
	"time"

	"github.com/letianxing/volta-brain/internal/contracts"
	"github.com/letianxing/volta-brain/internal/infra/eventbus"
)

// WindowAggregatorNode 是事件汇聚的唯一窗口节点。
type WindowAggregatorNode struct {
	name     string
	internal *eventbus.Bus
	window   time.Duration

	ctx   context.Context
	mu    sync.Mutex
	timer *time.Timer
	items []contracts.CandidateEvent
}

// NewWindowAggregatorNode 创建窗口聚合节点。
func NewWindowAggregatorNode(internal *eventbus.Bus, window time.Duration) *WindowAggregatorNode {
	return &WindowAggregatorNode{
		name:     "window_aggregator",
		internal: internal,
		window:   window,
	}
}

// Name 返回节点名称。
func (n *WindowAggregatorNode) Name() string { return n.name }

// Start 注册订阅。
func (n *WindowAggregatorNode) Start(ctx context.Context) error {
	n.ctx = ctx
	return n.internal.Subscribe(ctx, contracts.TopicCandidateEvent, func(_ context.Context, _ string, payload []byte) error {
		event, err := decode[contracts.CandidateEvent](payload)
		if err != nil {
			return err
		}
		n.enqueue(event)
		return nil
	})
}

func (n *WindowAggregatorNode) enqueue(event contracts.CandidateEvent) {
	n.mu.Lock()
	defer n.mu.Unlock()

	n.items = append(n.items, event)
	if n.timer != nil {
		return
	}
	n.timer = time.AfterFunc(n.window, n.flush)
}

func (n *WindowAggregatorNode) flush() {
	n.mu.Lock()
	events := append([]contracts.CandidateEvent(nil), n.items...)
	n.items = nil
	n.timer = nil
	n.mu.Unlock()

	if len(events) == 0 || n.ctx == nil {
		return
	}

	envelope := contracts.WindowEnvelope{
		WindowID: nextID("window"),
		Events:   events,
		ClosedAt: now(),
	}
	_ = n.internal.PublishJSON(n.ctx, contracts.TopicWindowReady, envelope)

	first := events[0]
	_ = publishPipeline(n.ctx, n.internal, contracts.PipelineEvent{
		Node:    n.name,
		Phase:   "window_ready",
		TraceID: first.TraceID,
		InputID: first.InputID,
		Details: map[string]any{
			"events": len(events),
		},
	})
}
