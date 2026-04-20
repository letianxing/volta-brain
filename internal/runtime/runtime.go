package runtime

import (
	"context"

	"github.com/letianxing/volta-brain/internal/config"
	"github.com/letianxing/volta-brain/internal/contracts"
	"github.com/letianxing/volta-brain/internal/infra/eventbus"
	"github.com/letianxing/volta-brain/internal/infra/natsio"
	"github.com/letianxing/volta-brain/internal/nodes"
	"github.com/letianxing/volta-brain/internal/state"
)

// Runtime 负责装配并启动全部结构节点。
type Runtime struct {
	cfg      config.Config
	internal *eventbus.Bus
	external natsio.Bus
	store    *state.Store
	nodes    []nodes.Node
}

// New 创建一个新的运行时实例。
func New(cfg config.Config, external natsio.Bus) *Runtime {
	internal := eventbus.New(cfg.EventBusBuffer)
	store := state.NewStore()

	nodeList := []nodes.Node{
		nodes.NewNATSInputNode(cfg.InputSubject, external, internal),
		nodes.NewContextAssemblerNode(internal, store),
		nodes.NewParallelDispatcherNode(internal),
		nodes.NewPathNode("path_explicit", contracts.PathExplicit, contracts.TopicPathExplicit, internal),
		nodes.NewPathNode("path_reflex", contracts.PathReflex, contracts.TopicPathReflex, internal),
		nodes.NewPathNode("path_llm", contracts.PathLLM, contracts.TopicPathLLM, internal),
		nodes.NewWindowAggregatorNode(internal, cfg.WindowDuration),
		nodes.NewGateNode(internal),
		nodes.NewArbiterNode(internal),
		nodes.NewStateMachineNode(internal, store),
		nodes.NewPersonalizationNode(internal, store),
		nodes.NewBehaviorLibraryNode(internal),
		nodes.NewNATSOutputNode(external, internal, map[string]string{
			contracts.TopicPipeline:           cfg.PipelineSubject,
			contracts.TopicBehaviorOutput:     cfg.ActionSubject,
			contracts.TopicStateMachineOutput: cfg.StateMachineSubject,
			contracts.TopicStateOutput:        cfg.StateSubject,
		}),
	}

	return &Runtime{
		cfg:      cfg,
		internal: internal,
		external: external,
		store:    store,
		nodes:    nodeList,
	}
}

// Start 启动全部节点。
func (r *Runtime) Start(ctx context.Context) error {
	for _, node := range r.nodes {
		if err := node.Start(ctx); err != nil {
			return err
		}
	}
	return nil
}

// Close 关闭 runtime。
func (r *Runtime) Close() error {
	if err := r.internal.Close(); err != nil {
		_ = r.external.Close()
		return err
	}
	return r.external.Close()
}

// Store 暴露内存状态仓，便于测试注入 personality/memory/internal-state。
func (r *Runtime) Store() *state.Store {
	return r.store
}
