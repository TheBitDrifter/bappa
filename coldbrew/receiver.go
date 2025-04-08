package coldbrew

import (
	"github.com/TheBitDrifter/bappa/blueprint/input"
)

// Receiver combines multiple input layouts and manages input state
// It handles keyboard, gamepad, mouse, and touch inputs
type Receiver interface {
	RegisterPad(padID int)
	Active() bool
	PopActions() []input.StampedAction
	PadLayout
	KeyLayout
	MouseLayout
	TouchLayout
}

type receiver struct {
	active bool
	actions
	*keyLayout
	*padLayout
	*mouseLayout
	*touchLayout
}

type actions struct {
	touches []input.StampedAction
	pad     []input.StampedAction
	mouse   []input.StampedAction
	kb      []input.StampedAction
}

// Active returns whether the receiver is accepting input
func (receiver receiver) Active() bool {
	return receiver.active
}

// PopActions collects all buffered actions and clears the buffers
func (receiver *receiver) PopActions() []input.StampedAction {
	removed := []input.StampedAction{}
	for _, input := range receiver.actions.kb {
		removed = append(removed, input)
	}
	for _, input := range receiver.actions.mouse {
		removed = append(removed, input)
	}
	for _, input := range receiver.actions.pad {
		removed = append(removed, input)
	}
	for _, input := range receiver.actions.touches {
		removed = append(removed, input)
	}
	receiver.actions = actions{}

	return removed
}
