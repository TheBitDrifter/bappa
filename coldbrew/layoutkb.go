package coldbrew

import (
	blueprint_input "github.com/TheBitDrifter/blueprint/input"
	"github.com/TheBitDrifter/mask"
	"github.com/hajimehoshi/ebiten/v2"
)

// KeyLayout maps keyboard keys to game inputs
type KeyLayout interface {
	RegisterKey(ebiten.Key, blueprint_input.Input)
}

type keyLayout struct {
	mask mask.Mask256
	keys []blueprint_input.Input // indexed by ebiten key
}

// RegisterKey maps a key to an input and marks it in the mask.
func (layout *keyLayout) RegisterKey(key ebiten.Key, input blueprint_input.Input) {
	if len(layout.keys) <= int(key) {
		newKeys := make([]blueprint_input.Input, key+1)
		copy(newKeys, layout.keys)
		layout.keys = newKeys
	}
	layout.keys[key] = input
	layout.mask.Mark(uint32(key))
}
