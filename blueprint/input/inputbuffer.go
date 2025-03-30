package input

import (
	"fmt"
	"sort"
)

// InputBuffer represents a buffer of timestamped inputs with automatic deduplication.
// It maintains only the most recent input of each type.
type InputBuffer struct {
	Values        []StampedInput
	ReceiverIndex int
}

// Add appends a new stamped input to the buffer, automatically deduplicating
// by keeping only the most recent input of each type
func (buffer *InputBuffer) Add(input StampedInput) {
	// Find the most recent input among existing inputs of the same type and the new input
	var mostRecent StampedInput = input

	// Create a new slice to hold non-matching inputs
	newInputs := make([]StampedInput, 0, len(buffer.Values))

	// Check all existing inputs
	for _, existing := range buffer.Values {
		if existing.Val == input.Val {
			// If this is more recent than our current "most recent", update it
			if existing.Tick > mostRecent.Tick {
				mostRecent = existing
			}
			// Skip adding this to newInputs (we'll add the most recent one later)
		} else {
			// Keep all inputs of different types
			newInputs = append(newInputs, existing)
		}
	}

	// Add the most recent input of the matching type
	newInputs = append(newInputs, mostRecent)

	// Update the buffer
	buffer.Values = newInputs
}

// ForceAdd appends a new stamped input to the buffer, without deduplicating
func (buffer *InputBuffer) ForceAdd(input StampedInput) {
	buffer.Values = append(buffer.Values, input)
}

// AddBatch adds multiple stamped inputs to the buffer with automatic deduplication
func (buffer *InputBuffer) AddBatch(inputs []StampedInput) {
	// Group inputs by value, keeping only the most recent
	latestByValue := make(map[Input]StampedInput)

	// First process existing buffer
	for _, stamped := range buffer.Values {
		existing, exists := latestByValue[stamped.Val]
		if !exists || stamped.Tick > existing.Tick {
			latestByValue[stamped.Val] = stamped
		}
	}

	// Then process new inputs
	for _, input := range inputs {
		existing, exists := latestByValue[input.Val]
		if !exists || input.Tick > existing.Tick {
			latestByValue[input.Val] = input
		}
	}

	// Convert map back to slice
	newInputs := make([]StampedInput, 0, len(latestByValue))
	for _, stamped := range latestByValue {
		newInputs = append(newInputs, stamped)
	}

	buffer.Values = newInputs
}

// ConsumeInput finds and removes the most recent occurrence of the target input.
// Returns the consumed input and whether it was found.
func (buffer *InputBuffer) ConsumeInput(target Input) (StampedInput, bool) {
	var mostRecent StampedInput
	var found bool

	// Find the most recent one
	for _, stamped := range buffer.Values {
		if stamped.Val == target && (!found || stamped.Tick > mostRecent.Tick) {
			mostRecent = stamped
			found = true
		}
	}

	if found {
		// Remove the consumed input
		newInputs := make([]StampedInput, 0, len(buffer.Values))
		for _, stamped := range buffer.Values {
			if stamped.Val != target {
				newInputs = append(newInputs, stamped)
			}
		}
		buffer.Values = newInputs
	}

	return mostRecent, found
}

// Clear removes all inputs from the buffer
func (buffer *InputBuffer) Clear() {
	buffer.Values = make([]StampedInput, 0)
}

// SetInputs replaces all inputs in the buffer with the provided inputs,
// automatically deduplicating them
func (buffer *InputBuffer) SetInputs(inputs []StampedInput) {
	buffer.Clear()
	buffer.AddBatch(inputs)
}

// Size returns the current number of inputs in the buffer
func (buffer *InputBuffer) Size() int {
	return len(buffer.Values)
}

// IsEmpty returns true if the buffer contains no inputs
func (buffer *InputBuffer) IsEmpty() bool {
	return len(buffer.Values) == 0
}

// PeekLatest returns the most recent input in the buffer without removing it.
// Returns false if the buffer is empty.
func (buffer *InputBuffer) PeekLatest() (StampedInput, bool) {
	if len(buffer.Values) == 0 {
		return StampedInput{}, false
	}

	latest := buffer.Values[0]
	for _, stamped := range buffer.Values {
		if stamped.Tick > latest.Tick {
			latest = stamped
		}
	}
	return latest, true
}

// PeekLatestOfType returns the most recent input of a specific type without removing it.
// Returns false if no input of that type exists.
func (buffer *InputBuffer) PeekLatestOfType(target Input) (StampedInput, bool) {
	for _, stamped := range buffer.Values {
		if stamped.Val == target {
			return stamped, true
		}
	}
	return StampedInput{}, false
}

// HasInput returns true if the buffer contains the specified input type
func (buffer *InputBuffer) HasInput(target Input) bool {
	for _, stamped := range buffer.Values {
		if stamped.Val == target {
			return true
		}
	}
	return false
}

// GetTimeRange returns the earliest and latest ticks in the buffer.
// Returns (0, 0) if the buffer is empty.
func (buffer *InputBuffer) GetTimeRange() (earliest int, latest int) {
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

// Clone returns a new InputBuffer with a copy of all current inputs
func (buffer *InputBuffer) Clone() InputBuffer {
	clone := InputBuffer{
		Values: make([]StampedInput, len(buffer.Values)),
	}
	copy(clone.Values, buffer.Values)
	return clone
}

// GetInputsInTimeRange returns all inputs between startTick and endTick (inclusive)
func (buffer *InputBuffer) GetInputsInTimeRange(startTick, endTick int) []StampedInput {
	result := make([]StampedInput, 0)
	for _, stamped := range buffer.Values {
		if stamped.Tick >= startTick && stamped.Tick <= endTick {
			result = append(result, stamped)
		}
	}
	return result
}

// GetSortedByTime returns all inputs sorted by their tick values
func (buffer *InputBuffer) GetSortedByTime() []StampedInput {
	sorted := make([]StampedInput, len(buffer.Values))
	copy(sorted, buffer.Values)
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].Tick < sorted[j].Tick
	})
	return sorted
}

// String returns a human-readable representation of the buffer
func (buffer *InputBuffer) String() string {
	if len(buffer.Values) == 0 {
		return "InputBuffer{empty}"
	}

	sorted := buffer.GetSortedByTime()
	result := "InputBuffer{\n"
	for _, input := range sorted {
		result += fmt.Sprintf("  {Val: %v, Tick: %d}\n", input.Val, input.Tick)
	}
	result += "}"
	return result
}
