package coldbrew

import (
	blueprint_input "github.com/TheBitDrifter/blueprint/input"
)

// Receiver combines multiple input layouts and manages input state
// It handles keyboard, gamepad, mouse, and touch inputs
type Receiver interface {
	RegisterPad(padID int)
	Active() bool
	PopInputs() []blueprint_input.StampedInput
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
	touches []blueprint_input.StampedInput // Input buffer for touch events
	pad     []blueprint_input.StampedInput // Input buffer for gamepad events
	mouse   []blueprint_input.StampedInput // Input buffer for mouse events
	kb      []blueprint_input.StampedInput // Input buffer for keyboard events
}

// Active returns whether the receiver is accepting input
func (receiver receiver) Active() bool {
	return receiver.active
}

// PopInputs collects all buffered inputs and clears the buffers
func (receiver *receiver) PopInputs() []blueprint_input.StampedInput {
	removed := []blueprint_input.StampedInput{}
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
