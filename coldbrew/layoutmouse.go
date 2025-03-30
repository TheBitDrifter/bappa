package coldbrew

import (
	blueprintinput "github.com/TheBitDrifter/blueprint/input"
	"github.com/TheBitDrifter/mask"
	"github.com/hajimehoshi/ebiten/v2"
)

// MouseLayout maps mouse buttons to game inputs
type MouseLayout interface {
	RegisterMouseButton(ebiten.MouseButton, blueprintinput.Input)
}

type mouseLayout struct {
	mask            mask.Mask
	mouseButtonsRaw []ebiten.MouseButton   // stores original button mappings
	mouseButtons    []blueprintinput.Input // indexed by button ID
}

// RegisterMouseButton maps a mouse button to an input. Duplicate registrations are ignored
func (layout *mouseLayout) RegisterMouseButton(button ebiten.MouseButton, input blueprintinput.Input) {
	btnU32 := uint32(button)
	if layout.mask.Contains(btnU32) {
		return
	}
	layout.mouseButtonsRaw = append(layout.mouseButtonsRaw, button)
	if len(layout.mouseButtons) <= int(btnU32) {
		newMouseBtns := make([]blueprintinput.Input, button+1)
		copy(newMouseBtns, layout.mouseButtons)
		layout.mouseButtons = newMouseBtns
	}
	layout.mouseButtons[button] = input
	layout.mask.Mark(btnU32)
}
