package input

import (
	"fmt"
	"sort"
)

// ActionBuffer represents a buffer of timestamped inputs with automatic deduplication.
type ActionBuffer struct {
	Values        []StampedAction
	ReceiverIndex int
}

// Add appends a new stamped action to the buffer, automatically deduplicating
// by keeping only the most recent action of each type
func (buffer *ActionBuffer) Add(action StampedAction) {
	// Find the most recent action among existing inputs of the same type and the new input
	var mostRecent StampedAction = action

	// Create a new slice to hold non-matching actions
	newActions := make([]StampedAction, 0, len(buffer.Values))

	// Check all existing action
	for _, existing := range buffer.Values {
		if existing.Val == action.Val {
			// If this is more recent than our current "most recent", update it
			if existing.Tick > mostRecent.Tick {
				mostRecent = existing
			}
		} else {
			newActions = append(newActions, existing)
		}
	}

	newActions = append(newActions, mostRecent)
	buffer.Values = newActions
}

// ForceAdd appends a new stamped action to the buffer, without deduplicating
func (buffer *ActionBuffer) ForceAdd(action StampedAction) {
	buffer.Values = append(buffer.Values, action)
}

// AddBatch adds multiple stamped actions to the buffer with automatic deduplication
func (buffer *ActionBuffer) AddBatch(actions []StampedAction) {
	// Group actions by value, keeping only the most recent
	latestByValue := make(map[Action]StampedAction)

	// First process existing buffer
	for _, stamped := range buffer.Values {
		existing, exists := latestByValue[stamped.Val]
		if !exists || stamped.Tick > existing.Tick {
			latestByValue[stamped.Val] = stamped
		}
	}

	// Then process new actions
	for _, input := range actions {
		existing, exists := latestByValue[input.Val]
		if !exists || input.Tick > existing.Tick {
			latestByValue[input.Val] = input
		}
	}

	// Convert map back to slice
	newActions := make([]StampedAction, 0, len(latestByValue))
	for _, stamped := range latestByValue {
		newActions = append(newActions, stamped)
	}

	buffer.Values = newActions
}

// ConsumeAction finds and removes the most recent occurrence of the target input.
// Returns the consumed input and whether it was found.
func (buffer *ActionBuffer) ConsumeAction(target Action) (StampedAction, bool) {
	var mostRecent StampedAction
	var found bool

	// Find the most recent one
	for _, stamped := range buffer.Values {
		if stamped.Val == target && (!found || stamped.Tick > mostRecent.Tick) {
			mostRecent = stamped
			found = true
		}
	}

	if found {
		// Remove the consumed action
		newActions := make([]StampedAction, 0, len(buffer.Values))
		for _, stamped := range buffer.Values {
			if stamped.Val != target {
				newActions = append(newActions, stamped)
			}
		}
		buffer.Values = newActions
	}

	return mostRecent, found
}

// Clear removes all actions from the buffer
func (buffer *ActionBuffer) Clear() {
	buffer.Values = make([]StampedAction, 0)
}

// SetActions replaces all actions in the buffer with the provided actions,
// automatically deduplicating them
func (buffer *ActionBuffer) SetActions(actions []StampedAction) {
	buffer.Clear()
	buffer.AddBatch(actions)
}

// Size returns the current number of actions in the buffer
func (buffer *ActionBuffer) Size() int {
	return len(buffer.Values)
}

// IsEmpty returns true if the buffer contains no actions
func (buffer *ActionBuffer) IsEmpty() bool {
	return len(buffer.Values) == 0
}

// PeekLatest returns the most recent action in the buffer without removing it.
// Returns false if the buffer is empty.
func (buffer *ActionBuffer) PeekLatest() (StampedAction, bool) {
	if len(buffer.Values) == 0 {
		return StampedAction{}, false
	}

	latest := buffer.Values[0]
	for _, stamped := range buffer.Values {
		if stamped.Tick > latest.Tick {
			latest = stamped
		}
	}
	return latest, true
}

// PeekLatestOfType returns the most recent action of a specific type without removing it.
// Returns false if no action of that type exists.
func (buffer *ActionBuffer) PeekLatestOfType(target Action) (StampedAction, bool) {
	for _, stamped := range buffer.Values {
		if stamped.Val == target {
			return stamped, true
		}
	}
	return StampedAction{}, false
}

// HasAction returns true if the buffer contains the specified input type
func (buffer *ActionBuffer) HasAction(target Action) bool {
	for _, stamped := range buffer.Values {
		if stamped.Val == target {
			return true
		}
	}
	return false
}

// GetTimeRange returns the earliest and latest ticks in the buffer.
// Returns (0, 0) if the buffer is empty.
func (buffer *ActionBuffer) GetTimeRange() (earliest int, latest int) {
	if len(buffer.Values) == 0 {
		return 0, 0
	}

	earliest = buffer.Values[0].Tick
	latest = earliest

	for _, stamped := range buffer.Values {
		if stamped.Tick < earliest {
			earliest = stamped.Tick
		}
		if stamped.Tick > latest {
			latest = stamped.Tick
		}
	}
	return earliest, latest
}

// Clone returns a new ActionBuffer with a copy of all current inputs
func (buffer *ActionBuffer) Clone() ActionBuffer {
	clone := ActionBuffer{
		Values: make([]StampedAction, len(buffer.Values)),
	}
	copy(clone.Values, buffer.Values)
	return clone
}

// GetActionsInTimeRange returns all actions between startTick and endTick (inclusive)
func (buffer *ActionBuffer) GetActionsInTimeRange(startTick, endTick int) []StampedAction {
	result := make([]StampedAction, 0)
	for _, stamped := range buffer.Values {
		if stamped.Tick >= startTick && stamped.Tick <= endTick {
			result = append(result, stamped)
		}
	}
	return result
}

// GetSortedByTime returns all actions sorted by their tick values
func (buffer *ActionBuffer) GetSortedByTime() []StampedAction {
	sorted := make([]StampedAction, len(buffer.Values))
	copy(sorted, buffer.Values)
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].Tick < sorted[j].Tick
	})
	return sorted
}

// String returns a human-readable representation of the buffer
func (buffer *ActionBuffer) String() string {
	if len(buffer.Values) == 0 {
		return "ActionBuffer{empty}"
	}

	sorted := buffer.GetSortedByTime()
	result := "ActionBuffer{\n"
	for _, action := range sorted {
		result += fmt.Sprintf("  {Val: %v, Tick: %d}\n", action.Val, action.Tick)
	}
	result += "}"
	return result
}
