package coldbrew

import (
	"errors"
	"fmt"
	"image"
	"io/fs"
	"sync"

	"github.com/TheBitDrifter/bappa/blueprint/client"
	"github.com/TheBitDrifter/bappa/environment"
	"github.com/TheBitDrifter/bappa/warehouse"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
)

type SpriteLoader interface {
	Load(spriteBundle *client.SpriteBundle, cache warehouse.Cache[Sprite]) error
	PreLoad(bundle client.PreLoadAssetBundle, cache warehouse.Cache[Sprite]) error
}

// spriteLoader handles loading and caching of sprite images
type spriteLoader struct {
	mu sync.RWMutex
	fs fs.FS
}

// NewSpriteLoader creates a sprite loader with the provided filesystem
func NewSpriteLoader(embeddedFS fs.FS) *spriteLoader {
	return &spriteLoader{
		fs: embeddedFS,
	}
}

// Load processes sprite bundles and caches their contents
// It uses the provided cache for lookups and registration
// which enables cache busting when a new cache is provided
func (loader *spriteLoader) Load(spriteBundle *client.SpriteBundle, cache warehouse.Cache[Sprite]) error {
	for i := range spriteBundle.Blueprints {

		spriteBlueprint := &spriteBundle.Blueprints[i]
		if spriteBlueprint.Location.Key == "" {
			continue
		}

		spriteIndex, ok := cache.GetIndex(spriteBlueprint.Location.Key)
		if ok {
			if spriteIndex > int(ClientConfig.maxSpritesCached.Load()) {
				return errors.New("max sprites error")
			}
			spriteBlueprint.Location.Index.Store(uint32(spriteIndex))
			continue
		}

		spr, err := loader.loadSpriteFromPath(spriteBlueprint.Location.Key)
		if err != nil {
			return err
		}

		index, err := cache.Register(spriteBlueprint.Location.Key, spr)
		if err != nil {
			return err
		}
		if index > int(ClientConfig.maxSpritesCached.Load()) {
			return errors.New("max sprites error")
		}

		spriteBlueprint.Location.Index.Store(uint32(index))
	}

	return nil
}

// Load processes sprite bundles and caches their contents
// It uses the provided cache for lookups and registration
// which enables cache busting when a new cache is provided
func (loader *spriteLoader) PreLoad(bundle client.PreLoadAssetBundle, cache warehouse.Cache[Sprite]) error {
	for i := range bundle {
		preloadBlueprint := &bundle[i]

		if preloadBlueprint.Path == "" || preloadBlueprint.Type != client.PreloadSprite {
			continue
		}

		spriteIndex, ok := cache.GetIndex(preloadBlueprint.Path)
		if ok {
			if spriteIndex > int(ClientConfig.maxSpritesCached.Load()) {
				return errors.New("max sprites error")
			}
			continue
		}

		spr, err := loader.loadSpriteFromPath(preloadBlueprint.Path)
		if err != nil {
			return err
		}

		index, err := cache.Register(preloadBlueprint.Path, spr)
		if err != nil {
			return err
		}
		if index > int(ClientConfig.maxSpritesCached.Load()) {
			return errors.New("max sprites error")
		}
	}

	return nil
}

// loadSpriteFromPath loads an image file from either filesystem or embedded assets
// In development mode, it loads from the local filesystem
// In production mode, it loads from embedded assets
func (loader *spriteLoader) loadSpriteFromPath(path string) (Sprite, error) {
	if !environment.IsProd() && !environment.IsWASM() {
		// Development mode: load from filesystem
		updatedPath := ClientConfig.localAssetPath + path
		img, _, err := ebitenutil.NewImageFromFile(updatedPath)
		if err != nil {
			return &sprite{}, err
		}
		return &sprite{
			image: img,
			name:  updatedPath,
		}, nil
	}

	// Production mode: load from embedded assets
	imgFile, err := loader.fs.Open(path)
	if err != nil {
		return &sprite{}, fmt.Errorf("failed to open embedded image %s: %w", path, err)
	}
	defer imgFile.Close()

	img, _, err := image.Decode(imgFile)
	if err != nil {
		return &sprite{}, fmt.Errorf("failed to decode embedded image %s: %w", path, err)
	}

	return &sprite{
		image: ebiten.NewImageFromImage(img),
		name:  path,
	}, nil
}
