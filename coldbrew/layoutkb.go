package coldbrew

import (
	"github.com/TheBitDrifter/bappa/blueprint/input"
	"github.com/TheBitDrifter/mask"
	"github.com/hajimehoshi/ebiten/v2"
)

// KeyLayout maps keyboard keys to game actions
type KeyLayout interface {
	RegisterKey(ebiten.Key, input.Action)
	RegisterReleasedKey(ebiten.Key, input.Action)
	RegisterJustPressedKey(ebiten.Key, input.Action)
}

type keyLayout struct {
	mask            mask.Mask256
	releasedMask    mask.Mask256
	justPressedMask mask.Mask256

	keys            []input.Action // indexed by ebiten key
	releasedKeys    []input.Action // indexed by ebiten key
	justPressedKeys []input.Action // indexed by ebiten key
}

func (layout *keyLayout) RegisterKey(key ebiten.Key, localInput input.Action) {
	layout.register(key, localInput, &layout.keys, &layout.mask)
}

func (layout *keyLayout) RegisterReleasedKey(key ebiten.Key, action input.Action) {
	layout.register(key, action, &layout.releasedKeys, &layout.releasedMask)
}

func (layout *keyLayout) RegisterJustPressedKey(key ebiten.Key, action input.Action) {
	layout.register(key, action, &layout.justPressedKeys, &layout.justPressedMask)
}

func (layout *keyLayout) register(key ebiten.Key, action input.Action, selectedKeySlice *[]input.Action, selectedKeyMask *mask.Mask256) {
	if len(*selectedKeySlice) <= int(key) {
		newKeys := make([]input.Action, key+1)
		copy(newKeys, *selectedKeySlice)
		*selectedKeySlice = newKeys
	}

	(*selectedKeySlice)[key] = action
	selectedKeyMask.Mark(uint32(key))
}
