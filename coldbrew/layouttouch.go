package coldbrew

import "github.com/TheBitDrifter/bappa/blueprint/input"

// TouchLayout maps touch input to game actions
type TouchLayout interface {
	RegisterTouch(input.Input)
}

type touchLayout struct {
	active bool        // indicates if touch input is enabled
	input  input.Input // associated game action
}

// RegisterTouch enables touch input and maps it to a game action
func (layout *touchLayout) RegisterTouch(localInput input.Input) {
	layout.active = true
	layout.input = localInput
}
