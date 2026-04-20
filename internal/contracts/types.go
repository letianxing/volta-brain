package contracts

import "time"

// Path 标识三路径中的某一路。
type Path string

const (
	PathExplicit Path = "explicit"
	PathReflex   Path = "reflex"
	PathLLM      Path = "llm"
)

// InputFrame 是从 NATS 进入 brain 的统一输入包。
//
// 这里不承载任何业务 topic 或 sensor schema，只携带结构化链路控制信息。
type InputFrame struct {
	ID         string            `json:"id"`
	TraceID    string            `json:"trace_id"`
	RobotID    string            `json:"robot_id"`
	Source     string            `json:"source"`
	Timestamp  time.Time         `json:"timestamp"`
	Payload    map[string]any    `json:"payload,omitempty"`
	Directives []PathDirective   `json:"directives"`
	Metadata   map[string]string `json:"metadata,omitempty"`
}

// PathDirective 用于告诉骨架某条路径应该产出什么候选事件。
//
// 这使得项目可以完全去业务化，只验证链路和节点职责。
type PathDirective struct {
	Path         Path              `json:"path"`
	Active       bool              `json:"active"`
	Priority     int               `json:"priority"`
	Allow        bool              `json:"allow"`
	TargetState  string            `json:"target_state,omitempty"`
	RejectReason string            `json:"reject_reason,omitempty"`
	Payload      map[string]any    `json:"payload,omitempty"`
	Metadata     map[string]string `json:"metadata,omitempty"`
}

// PersonalityShell 对应图中的“性格”上下文仓空壳。
type PersonalityShell struct{}

// MemoryShell 对应图中的“记忆”上下文仓空壳。
type MemoryShell struct{}

// InternalStateShell 对应图中的“内部状态”空壳。
type InternalStateShell struct{}

// BrainContext 汇总 personality/memory/internal-state。
type BrainContext struct {
	Personality PersonalityShell   `json:"personality"`
	Memory      MemoryShell        `json:"memory"`
	Internal    InternalStateShell `json:"internal"`
}

// ContextualFrame 是完成上下文注入后的输入载体。
type ContextualFrame struct {
	Frame   InputFrame   `json:"frame"`
	Context BrainContext `json:"context"`
}

// PathWorkItem 是派发给具体路径节点的工作项。
type PathWorkItem struct {
	Frame     ContextualFrame `json:"frame"`
	Directive PathDirective   `json:"directive"`
}

// CandidateEvent 是三路径并行输出后的标准化候选事件。
type CandidateEvent struct {
	ID           string            `json:"id"`
	TraceID      string            `json:"trace_id"`
	InputID      string            `json:"input_id"`
	Path         Path              `json:"path"`
	Priority     int               `json:"priority"`
	Allow        bool              `json:"allow"`
	TargetState  string            `json:"target_state,omitempty"`
	RejectReason string            `json:"reject_reason,omitempty"`
	Context      BrainContext      `json:"context"`
	Payload      map[string]any    `json:"payload,omitempty"`
	Metadata     map[string]string `json:"metadata,omitempty"`
	Timestamp    time.Time         `json:"timestamp"`
}

// WindowEnvelope 是窗口聚合的输出。
type WindowEnvelope struct {
	WindowID string           `json:"window_id"`
	Events   []CandidateEvent `json:"events"`
	ClosedAt time.Time        `json:"closed_at"`
}

// GateAcceptedEnvelope 表示资格门放行后的候选事件集合。
type GateAcceptedEnvelope struct {
	WindowID  string           `json:"window_id"`
	Accepted  []CandidateEvent `json:"accepted"`
	Timestamp time.Time        `json:"timestamp"`
}

// RejectedEvent 表示被资格门挡下的事件。
type RejectedEvent struct {
	Event  CandidateEvent `json:"event"`
	Reason string         `json:"reason"`
}

// GateRejectedEnvelope 表示资格门拒绝结果集合。
type GateRejectedEnvelope struct {
	WindowID  string          `json:"window_id"`
	Rejected  []RejectedEvent `json:"rejected"`
	Timestamp time.Time       `json:"timestamp"`
}

// ArbitrationDecision 是仲裁器输出。
type ArbitrationDecision struct {
	Winner    CandidateEvent   `json:"winner"`
	Rejected  []CandidateEvent `json:"rejected"`
	Reason    string           `json:"reason"`
	Timestamp time.Time        `json:"timestamp"`
}

// StateMachineSnapshot 是状态机输出的通用状态快照。
type StateMachineSnapshot struct {
	Previous  string    `json:"previous"`
	Current   string    `json:"current"`
	UpdatedAt time.Time `json:"updated_at"`
}

// StateMachineOutput 是对外广播的状态机变化输出。
type StateMachineOutput struct {
	TraceID   string               `json:"trace_id"`
	InputID   string               `json:"input_id"`
	State     StateMachineSnapshot `json:"state"`
	Timestamp time.Time            `json:"timestamp"`
}

// BehaviorIntent 是状态机传给个性化/行为库的中间表示。
type BehaviorIntent struct {
	Decision ArbitrationDecision  `json:"decision"`
	State    StateMachineSnapshot `json:"state"`
	TraceID  string               `json:"trace_id"`
	InputID  string               `json:"input_id"`
}

// BehaviorReady 是个性化层传给行为库的结果。
type BehaviorReady struct {
	Intent      BehaviorIntent   `json:"intent"`
	Personality PersonalityShell `json:"personality"`
}

// ActionCommand 是行为库最终输出的纯结构命令。
type ActionCommand struct {
	ID                string            `json:"id"`
	TraceID           string            `json:"trace_id"`
	InputID           string            `json:"input_id"`
	SourceCandidateID string            `json:"source_candidate_id"`
	Status            string            `json:"status"`
	Path              Path              `json:"path"`
	State             string            `json:"state"`
	Payload           map[string]any    `json:"payload,omitempty"`
	GeneratedAt       time.Time         `json:"generated_at"`
	Metadata          map[string]string `json:"metadata,omitempty"`
}

// StateOutput 是输出给外部订阅者的状态快照。
type StateOutput struct {
	TraceID   string               `json:"trace_id"`
	InputID   string               `json:"input_id"`
	State     StateMachineSnapshot `json:"state"`
	Context   InternalStateShell   `json:"context"`
	Timestamp time.Time            `json:"timestamp"`
}

// PipelineEvent 用于输出每个结构节点的经过记录。
type PipelineEvent struct {
	Node      string         `json:"node"`
	Phase     string         `json:"phase"`
	TraceID   string         `json:"trace_id"`
	InputID   string         `json:"input_id"`
	Path      Path           `json:"path,omitempty"`
	Timestamp time.Time      `json:"timestamp"`
	Details   map[string]any `json:"details,omitempty"`
}
