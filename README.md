# volta-brain

`volta-brain` 是一个去业务逻辑化的认知链路骨架项目。

它保留了 `robo-brain` 的核心节点形态：

- 输入接入
- 三路径并行处理
- 内部事件总线
- 资格门
- 仲裁器
- 状态机
- 个性化
- 行为库
- 外部输出

与 `robo-brain` 不同的是，`volta-brain` 不直接承载任何 `ROS` 或 `rosbridge` 逻辑，外部边界统一使用 `NATS`。

## 架构摘要

```text
NATS Input
  -> ContextAssembler
  -> ParallelDispatcher
  -> explicit/reflex/llm path nodes
  -> WindowAggregator
  -> Gate
  -> Arbiter
  -> StateMachine
  -> Personalization
  -> BehaviorLibrary
  -> NATS Output
```

内部节点间通信使用 `Watermill + GoChannel`，外部输入输出使用 `nats.go`。

## 目录

- `cmd/volta-brain`: 启动入口
- `docs`: 技术架构文档
- `internal/contracts`: 核心消息契约
- `internal/infra/eventbus`: 内部事件总线
- `internal/infra/natsio`: NATS 抽象与实现
- `internal/nodes`: 各个结构化节点
- `internal/runtime`: 节点装配与生命周期
- `internal/state`: personality/memory/internal-state 仓
- `tests/e2e`: 端到端链路测试

## 运行

```bash
go run ./cmd/volta-brain
```

默认 NATS 输入输出 subject：

- `volta.input.frame`
- `volta.output.pipeline`
- `volta.output.action`
- `volta.output.state_machine`
- `volta.output.state`
