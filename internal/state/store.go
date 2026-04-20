package state

import (
	"sync"

	"github.com/letianxing/volta-brain/internal/contracts"
)

// Store 持有 personality/memory/internal-state 的骨架数据。
type Store struct {
	mu          sync.RWMutex
	personality contracts.PersonalityShell
	memory      contracts.MemoryShell
	internal    contracts.InternalStateShell
	state       string
}

// NewStore 创建一份带默认值的状态仓。
func NewStore() *Store {
	return &Store{
		personality: contracts.PersonalityShell{},
		memory:      contracts.MemoryShell{},
		internal:    contracts.InternalStateShell{},
		state:       "",
	}
}

// Snapshot 返回完整上下文快照。
func (s *Store) Snapshot() contracts.BrainContext {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return contracts.BrainContext{
		Personality: s.personality,
		Memory:      s.memory,
		Internal:    s.internal,
	}
}

// CurrentState 返回当前状态机状态。
func (s *Store) CurrentState() string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.state
}

// UpdateState 更新当前状态并返回状态机快照。
func (s *Store) UpdateState(next string) contracts.StateMachineSnapshot {
	s.mu.Lock()
	defer s.mu.Unlock()

	previous := s.state
	if next == "" {
		next = previous
	}
	s.state = next

	return contracts.StateMachineSnapshot{
		Previous:  previous,
		Current:   s.state,
		UpdatedAt: now(),
	}
}

// SetPersonality 覆盖默认 personality 仓。
func (s *Store) SetPersonality(snapshot contracts.PersonalityShell) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.personality = snapshot
}

// SetMemory 覆盖默认 memory 仓。
func (s *Store) SetMemory(snapshot contracts.MemoryShell) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.memory = snapshot
}

// SetInternal 覆盖默认 internal-state 仓。
func (s *Store) SetInternal(snapshot contracts.InternalStateShell) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.internal = snapshot
}
