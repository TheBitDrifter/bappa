package coldbrew

import (
	"github.com/TheBitDrifter/bappa/blueprint/input"
	"github.com/TheBitDrifter/mask"
	"github.com/hajimehoshi/ebiten/v2"
)

// PadLayout manages gamepad input mapping configuration
type PadLayout interface {
	RegisterPad(padID int)
	RegisterGamepadButton(ebiten.GamepadButton, input.Action)

	RegisterGamepadJustPressedButton(ebiten.GamepadButton, input.Action)
	RegisterGamepadReleasedButton(ebiten.GamepadButton, input.Action)

	RegisterGamepadAxes(left bool, input input.Action)
}

type padLayout struct {
	padID int

	mask            mask.Mask
	justPressedMask mask.Mask
	releaseMask     mask.Mask

	buttons        []input.Action
	pressed        []input.Action
	released       []input.Action
	leftAxes       bool
	rightAxes      bool
	leftAxesInput  input.Action
	rightAxesInput input.Action
}

// RegisterPad sets the gamepad identifier
func (layout *padLayout) RegisterPad(padID int) {
	layout.padID = padID
}

// RegisterGamepadButton maps a continuously pressed gamepad button to an input action.
func (layout *padLayout) RegisterGamepadButton(btn ebiten.GamepadButton, localInput input.Action) {
	layout.registerButton(btn, localInput, &layout.buttons, &layout.mask)
}

// RegisterGamepadJustPressedButton maps a gamepad button that was just pressed to an input action.
func (layout *padLayout) RegisterGamepadJustPressedButton(btn ebiten.GamepadButton, localInput input.Action) {
	layout.registerButton(btn, localInput, &layout.pressed, &layout.justPressedMask)
}

// RegisterGamepadReleasedButton maps a gamepad button that was just released to an input action.
func (layout *padLayout) RegisterGamepadReleasedButton(btn ebiten.GamepadButton, localInput input.Action) {
	layout.registerButton(btn, localInput, &layout.released, &layout.releaseMask)
}

func (layout *padLayout) registerButton(btn ebiten.GamepadButton, localInput input.Action, buttons *[]input.Action, m *mask.Mask) {
	// Ensure the slice is large enough to hold the new mapping.
	if len(*buttons) <= int(btn) {
		newBtns := make([]input.Action, btn+1)
		copy(newBtns, *buttons)
		*buttons = newBtns
	}
	(*buttons)[btn] = localInput
	m.Mark(uint32(btn))
}

// RegisterGamepadAxes maps an analog stick to an input action
func (layout *padLayout) RegisterGamepadAxes(left bool, localInput input.Action) {
	if left {
		layout.leftAxes = true
		layout.leftAxesInput = localInput
		return
	}
	layout.rightAxes = true
	layout.rightAxesInput = localInput
}
