package nodes

import (
	"context"
	"encoding/json"
	"fmt"
	"sync/atomic"
	"time"

	"github.com/letianxing/volta-brain/internal/contracts"
	"github.com/letianxing/volta-brain/internal/infra/eventbus"
)

// Node 定义 runtime 可启动的结构节点。
type Node interface {
	Name() string
	Start(ctx context.Context) error
}

var idCounter atomic.Uint64

func nextID(prefix string) string {
	return fmt.Sprintf("%s-%06d", prefix, idCounter.Add(1))
}

func decode[T any](payload []byte) (T, error) {
	var target T
	err := json.Unmarshal(payload, &target)
	return target, err
}

func now() time.Time {
	return time.Now().UTC()
}

func publishPipeline(ctx context.Context, bus *eventbus.Bus, event contracts.PipelineEvent) error {
	if event.Timestamp.IsZero() {
		event.Timestamp = now()
	}
	return bus.PublishJSON(ctx, contracts.TopicPipeline, event)
}

func pathTopic(path contracts.Path) string {
	switch path {
	case contracts.PathExplicit:
		return contracts.TopicPathExplicit
	case contracts.PathReflex:
		return contracts.TopicPathReflex
	case contracts.PathLLM:
		return contracts.TopicPathLLM
	default:
		return ""
	}
}

func mergeMetadata(left, right map[string]string) map[string]string {
	size := len(left) + len(right)
	if size == 0 {
		return nil
	}

	merged := make(map[string]string, size)
	for key, value := range left {
		merged[key] = value
	}
	for key, value := range right {
		merged[key] = value
	}
	return merged
}
