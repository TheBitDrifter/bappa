package coldbrew

import (
	"github.com/TheBitDrifter/bappa/blueprint/input"
	"github.com/TheBitDrifter/mask"
	"github.com/hajimehoshi/ebiten/v2"
)

// KeyLayout maps keyboard keys to game actions
type KeyLayout interface {
	RegisterKey(ebiten.Key, input.Action)
}

type keyLayout struct {
	mask mask.Mask256
	keys []input.Action // indexed by ebiten key
}

// RegisterKey maps a key to an action and marks it in the mask.
func (layout *keyLayout) RegisterKey(key ebiten.Key, localInput input.Action) {
	if len(layout.keys) <= int(key) {
		newKeys := make([]input.Action, key+1)
		copy(newKeys, layout.keys)
		layout.keys = newKeys
	}
	layout.keys[key] = localInput
	layout.mask.Mark(uint32(key))
}
