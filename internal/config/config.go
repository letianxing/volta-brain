package config

import (
	"os"
	"strconv"
	"time"

	"github.com/letianxing/volta-brain/internal/contracts"
)

// Config 定义 volta-brain 的运行时配置。
type Config struct {
	ServiceName         string
	NATSURL             string
	EventBusBuffer      int64
	WindowDuration      time.Duration
	InputSubject        string
	PipelineSubject     string
	ActionSubject       string
	StateMachineSubject string
	StateSubject        string
}

// Default 返回一组适合本地开发的默认配置。
func Default() Config {
	return Config{
		ServiceName:         "volta-brain",
		NATSURL:             "nats://127.0.0.1:4222",
		EventBusBuffer:      256,
		WindowDuration:      100 * time.Millisecond,
		InputSubject:        contracts.SubjectInputFrame,
		PipelineSubject:     contracts.SubjectPipeline,
		ActionSubject:       contracts.SubjectAction,
		StateMachineSubject: contracts.SubjectStateMachine,
		StateSubject:        contracts.SubjectState,
	}
}

// Load 从环境变量加载配置，缺失项回落到默认值。
func Load() Config {
	cfg := Default()

	if value := os.Getenv("VOLTA_SERVICE_NAME"); value != "" {
		cfg.ServiceName = value
	}
	if value := os.Getenv("VOLTA_NATS_URL"); value != "" {
		cfg.NATSURL = value
	}
	if value := os.Getenv("VOLTA_EVENTBUS_BUFFER"); value != "" {
		if parsed, err := strconv.ParseInt(value, 10, 64); err == nil && parsed > 0 {
			cfg.EventBusBuffer = parsed
		}
	}
	if value := os.Getenv("VOLTA_WINDOW_MS"); value != "" {
		if parsed, err := strconv.Atoi(value); err == nil && parsed > 0 {
			cfg.WindowDuration = time.Duration(parsed) * time.Millisecond
		}
	}
	if value := os.Getenv("VOLTA_INPUT_SUBJECT"); value != "" {
		cfg.InputSubject = value
	}
	if value := os.Getenv("VOLTA_PIPELINE_SUBJECT"); value != "" {
		cfg.PipelineSubject = value
	}
	if value := os.Getenv("VOLTA_ACTION_SUBJECT"); value != "" {
		cfg.ActionSubject = value
	}
	if value := os.Getenv("VOLTA_STATE_MACHINE_SUBJECT"); value != "" {
		cfg.StateMachineSubject = value
	}
	if value := os.Getenv("VOLTA_STATE_SUBJECT"); value != "" {
		cfg.StateSubject = value
	}

	return cfg
}
