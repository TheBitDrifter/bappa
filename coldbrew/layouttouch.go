package coldbrew

import "github.com/TheBitDrifter/bappa/blueprint/input"

// TouchLayout maps touch input to game actions
type TouchLayout interface {
	RegisterTouch(input.Action)
}

type touchLayout struct {
	active bool         // indicates if touch input is enabled
	input  input.Action // associated game action
}

// RegisterTouch enables touch input and maps it to a game action
func (layout *touchLayout) RegisterTouch(localInput input.Action) {
	layout.active = true
	layout.input = localInput
}
