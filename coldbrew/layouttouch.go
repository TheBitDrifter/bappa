package coldbrew

import blueprintinput "github.com/TheBitDrifter/blueprint/input"

// TouchLayout maps touch input to game actions
type TouchLayout interface {
	RegisterTouch(blueprintinput.Input)
}

type touchLayout struct {
	active bool                 // indicates if touch input is enabled
	input  blueprintinput.Input // associated game action
}

// RegisterTouch enables touch input and maps it to a game action
func (layout *touchLayout) RegisterTouch(input blueprintinput.Input) {
	layout.active = true
	layout.input = input
}
