package nodes

import (
	"context"
	"fmt"
	"sync"

	"github.com/letianxing/volta-brain/internal/contracts"
	"github.com/letianxing/volta-brain/internal/infra/eventbus"
)

// ParallelDispatcherNode 负责把一个输入拆成三路径并行工作项。
type ParallelDispatcherNode struct {
	name     string
	internal *eventbus.Bus
}

// NewParallelDispatcherNode 创建并行分发节点。
func NewParallelDispatcherNode(internal *eventbus.Bus) *ParallelDispatcherNode {
	return &ParallelDispatcherNode{
		name:     "parallel_dispatcher",
		internal: internal,
	}
}

// Name 返回节点名称。
func (n *ParallelDispatcherNode) Name() string { return n.name }

// Start 注册订阅。
func (n *ParallelDispatcherNode) Start(ctx context.Context) error {
	return n.internal.Subscribe(ctx, contracts.TopicInputContext, func(runCtx context.Context, _ string, payload []byte) error {
		frame, err := decode[contracts.ContextualFrame](payload)
		if err != nil {
			return err
		}

		var wg sync.WaitGroup
		errCh := make(chan error, len(frame.Frame.Directives))
		dispatched := 0

		for _, directive := range frame.Frame.Directives {
			if !directive.Active {
				continue
			}
			topic := pathTopic(directive.Path)
			if topic == "" {
				return fmt.Errorf("unsupported path %q", directive.Path)
			}

			dispatched++
			work := contracts.PathWorkItem{
				Frame:     frame,
				Directive: directive,
			}

			wg.Add(1)
			go func(topic string, work contracts.PathWorkItem) {
				defer wg.Done()
				if err := n.internal.PublishJSON(runCtx, topic, work); err != nil {
					errCh <- err
					return
				}
				errCh <- publishPipeline(runCtx, n.internal, contracts.PipelineEvent{
					Node:    n.name,
					Phase:   "dispatch",
					TraceID: work.Frame.Frame.TraceID,
					InputID: work.Frame.Frame.ID,
					Path:    work.Directive.Path,
				})
			}(topic, work)
		}

		wg.Wait()
		close(errCh)
		for dispatchErr := range errCh {
			if dispatchErr != nil {
				return dispatchErr
			}
		}

		return publishPipeline(runCtx, n.internal, contracts.PipelineEvent{
			Node:    n.name,
			Phase:   "fanout_complete",
			TraceID: frame.Frame.TraceID,
			InputID: frame.Frame.ID,
			Details: map[string]any{
				"active_paths": dispatched,
			},
		})
	})
}
