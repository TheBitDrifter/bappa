package coldbrew

import (
	"io/fs"
)

type assetManager struct {
	SpriteLoader
	SoundLoader
}

func newAssetManager(embeddedFS fs.FS) *assetManager {
	return &assetManager{
		SpriteLoader: NewSpriteLoader(embeddedFS),
		SoundLoader:  NewSoundLoader(embeddedFS),
	}
}
