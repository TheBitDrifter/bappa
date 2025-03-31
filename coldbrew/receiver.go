package coldbrew

import (
	"github.com/TheBitDrifter/bappa/blueprint/input"
)

// Receiver combines multiple input layouts and manages input state
// It handles keyboard, gamepad, mouse, and touch inputs
type Receiver interface {
	RegisterPad(padID int)
	Active() bool
	PopInputs() []input.StampedInput
	PadLayout
	KeyLayout
	MouseLayout
	TouchLayout
}

type receiver struct {
	active bool
	inputs
	*keyLayout
	*padLayout
	*mouseLayout
	*touchLayout
}

type inputs struct {
	touches []input.StampedInput // Input buffer for touch events
	pad     []input.StampedInput // Input buffer for gamepad events
	mouse   []input.StampedInput // Input buffer for mouse events
	kb      []input.StampedInput // Input buffer for keyboard events
}

// Active returns whether the receiver is accepting input
func (receiver receiver) Active() bool {
	return receiver.active
}

// PopInputs collects all buffered inputs and clears the buffers
func (receiver *receiver) PopInputs() []input.StampedInput {
	removed := []input.StampedInput{}
	for _, input := range receiver.inputs.kb {
		removed = append(removed, input)
	}
	for _, input := range receiver.inputs.mouse {
		removed = append(removed, input)
	}
	for _, input := range receiver.inputs.pad {
		removed = append(removed, input)
	}
	for _, input := range receiver.inputs.touches {
		removed = append(removed, input)
	}
	receiver.inputs = inputs{}

	return removed
}
