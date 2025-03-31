package coldbrew

import (
	"github.com/TheBitDrifter/bappa/blueprint/input"
	"github.com/TheBitDrifter/mask"
	"github.com/hajimehoshi/ebiten/v2"
)

// MouseLayout maps mouse buttons to game inputs
type MouseLayout interface {
	RegisterMouseButton(ebiten.MouseButton, input.Input)
}

type mouseLayout struct {
	mask            mask.Mask
	mouseButtonsRaw []ebiten.MouseButton // stores original button mappings
	mouseButtons    []input.Input        // indexed by button ID
}

// RegisterMouseButton maps a mouse button to an input. Duplicate registrations are ignored
func (layout *mouseLayout) RegisterMouseButton(button ebiten.MouseButton, localInput input.Input) {
	btnU32 := uint32(button)
	if layout.mask.Contains(btnU32) {
		return
	}
	layout.mouseButtonsRaw = append(layout.mouseButtonsRaw, button)
	if len(layout.mouseButtons) <= int(btnU32) {
		newMouseBtns := make([]input.Input, button+1)
		copy(newMouseBtns, layout.mouseButtons)
		layout.mouseButtons = newMouseBtns
	}
	layout.mouseButtons[button] = localInput
	layout.mask.Mark(btnU32)
}
