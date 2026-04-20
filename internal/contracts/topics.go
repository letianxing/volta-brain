package contracts

const (
	TopicInputIngress       = "input.ingress"
	TopicInputContext       = "input.context"
	TopicPathExplicit       = "path.explicit"
	TopicPathReflex         = "path.reflex"
	TopicPathLLM            = "path.llm"
	TopicCandidateEvent     = "candidate.event"
	TopicWindowReady        = "window.ready"
	TopicGateAccepted       = "gate.accepted"
	TopicGateRejected       = "gate.rejected"
	TopicArbitrationWinner  = "arbitration.winner"
	TopicBehaviorIntent     = "behavior.intent"
	TopicBehaviorReady      = "behavior.ready"
	TopicBehaviorOutput     = "behavior.output"
	TopicStateMachineOutput = "state_machine.output"
	TopicStateOutput        = "state.output"
	TopicPipeline           = "pipeline.output"
)

const (
	SubjectInputFrame   = "volta.input.frame"
	SubjectPipeline     = "volta.output.pipeline"
	SubjectAction       = "volta.output.action"
	SubjectStateMachine = "volta.output.state_machine"
	SubjectState        = "volta.output.state"
)
