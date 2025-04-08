package coldbrew

import (
	"github.com/TheBitDrifter/bappa/blueprint/input"
	"github.com/TheBitDrifter/mask"
	"github.com/hajimehoshi/ebiten/v2"
)

// MouseLayout maps mouse buttons to game actions
type MouseLayout interface {
	RegisterMouseButton(ebiten.MouseButton, input.Action)
}

type mouseLayout struct {
	mask            mask.Mask
	mouseButtonsRaw []ebiten.MouseButton // stores original button mappings
	mouseButtons    []input.Action       // indexed by button ID
}

// RegisterMouseButton maps a mouse button to an input. Duplicate registrations are ignored
func (layout *mouseLayout) RegisterMouseButton(button ebiten.MouseButton, localInput input.Action) {
	btnU32 := uint32(button)
	if layout.mask.Contains(btnU32) {
		return
	}
	layout.mouseButtonsRaw = append(layout.mouseButtonsRaw, button)
	if len(layout.mouseButtons) <= int(btnU32) {
		newMouseBtns := make([]input.Action, button+1)
		copy(newMouseBtns, layout.mouseButtons)
		layout.mouseButtons = newMouseBtns
	}
	layout.mouseButtons[button] = localInput
	layout.mask.Mark(btnU32)
}
