package coldbrew

import (
	blueprintinput "github.com/TheBitDrifter/blueprint/input"
	"github.com/TheBitDrifter/mask"
	"github.com/hajimehoshi/ebiten/v2"
)

// PadLayout manages gamepad input mapping configuration
type PadLayout interface {
	RegisterPad(padID int)
	RegisterGamepadButton(ebiten.GamepadButton, blueprintinput.Input)
	RegisterGamepadAxes(left bool, input blueprintinput.Input)
}

type padLayout struct {
	padID          int
	mask           mask.Mask
	buttons        []blueprintinput.Input
	leftAxes       bool
	rightAxes      bool
	leftAxesInput  blueprintinput.Input
	rightAxesInput blueprintinput.Input
}

// RegisterPad sets the gamepad identifier
func (layout *padLayout) RegisterPad(padID int) {
	layout.padID = padID
}

// RegisterGamepadButton maps a gamepad button to an input action
func (layout *padLayout) RegisterGamepadButton(btn ebiten.GamepadButton, input blueprintinput.Input) {
	if len(layout.buttons) <= int(btn) {
		newBtns := make([]blueprintinput.Input, btn+1)
		copy(newBtns, layout.buttons)
		layout.buttons = newBtns
	}
	layout.buttons[btn] = input
	layout.mask.Mark(uint32(btn))
}

// RegisterGamepadAxes maps an analog stick to an input action
func (layout *padLayout) RegisterGamepadAxes(left bool, input blueprintinput.Input) {
	if left {
		layout.leftAxes = true
		layout.leftAxesInput = input
		return
	}
	layout.rightAxes = true
	layout.rightAxesInput = input
}
