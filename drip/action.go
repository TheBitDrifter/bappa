package drip

import (
	"sync"

	"github.com/TheBitDrifter/bappa/blueprint/input"
	"github.com/TheBitDrifter/bappa/table"
)

// bufferedServerActions holds actions received from a connection, tagged with the
// server-determined target entity ID and its recycled count for validation.
type bufferedServerActions struct {
	TargetEntityID table.EntryID
	Recycled       int
	ReceiverIndex  int
	Actions        []input.StampedAction
}

type actionQueue struct {
	mu      sync.RWMutex
	actions []bufferedServerActions
}

func (q *actionQueue) ConsumeAll() []bufferedServerActions {
	q.mu.Lock()
	defer q.mu.Unlock()

	consumed := make([]bufferedServerActions, len(q.actions))
	copy(consumed, q.actions)
	q.actions = q.actions[:0]

	return consumed
}
