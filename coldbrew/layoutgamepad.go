package coldbrew

import (
	"github.com/TheBitDrifter/bappa/blueprint/input"
	"github.com/TheBitDrifter/mask"
	"github.com/hajimehoshi/ebiten/v2"
)

// PadLayout manages gamepad input mapping configuration
type PadLayout interface {
	RegisterPad(padID int)
	RegisterGamepadButton(ebiten.GamepadButton, input.Input)
	RegisterGamepadAxes(left bool, input input.Input)
}

type padLayout struct {
	padID          int
	mask           mask.Mask
	buttons        []input.Input
	leftAxes       bool
	rightAxes      bool
	leftAxesInput  input.Input
	rightAxesInput input.Input
}

// RegisterPad sets the gamepad identifier
func (layout *padLayout) RegisterPad(padID int) {
	layout.padID = padID
}

// RegisterGamepadButton maps a gamepad button to an input action
func (layout *padLayout) RegisterGamepadButton(btn ebiten.GamepadButton, localInput input.Input) {
	if len(layout.buttons) <= int(btn) {
		newBtns := make([]input.Input, btn+1)
		copy(newBtns, layout.buttons)
		layout.buttons = newBtns
	}
	layout.buttons[btn] = localInput
	layout.mask.Mark(uint32(btn))
}

// RegisterGamepadAxes maps an analog stick to an input action
func (layout *padLayout) RegisterGamepadAxes(left bool, localInput input.Input) {
	if left {
		layout.leftAxes = true
		layout.leftAxesInput = localInput
		return
	}
	layout.rightAxes = true
	layout.rightAxesInput = localInput
}
