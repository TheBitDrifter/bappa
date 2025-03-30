package coldbrew

import (
	"errors"

	blueprintclient "github.com/TheBitDrifter/blueprint/client"
	"github.com/TheBitDrifter/warehouse"
	"github.com/hajimehoshi/ebiten/v2"
)

// MockSpriteLoader implements the sprite loading functionality for tests
type MockSpriteLoader struct{}

// NewMockSpriteLoader creates a new MockSpriteLoader
func NewMockSpriteLoader() *MockSpriteLoader {
	return &MockSpriteLoader{}
}

// Load implements the sprite loading for tests
func (m *MockSpriteLoader) Load(spriteBundle *blueprintclient.SpriteBundle, cache warehouse.Cache[Sprite]) error {
	for i := range spriteBundle.Blueprints {
		spriteBlueprint := &spriteBundle.Blueprints[i]
		if spriteBlueprint.Location.Key == "" {
			continue
		}

		// First check if already in cache
		spriteIndex, ok := cache.GetIndex(spriteBlueprint.Location.Key)

		if ok {
			spriteBlueprint.Location.Index.Store(uint32(spriteIndex))
			continue
		}

		// Check if we have a mock for this sprite
		// Create a new mock sprite on demand
		spr := &sprite{
			name:  spriteBlueprint.Location.Key,
			image: ebiten.NewImage(10, 10), // Small placeholder image

		}

		// Register in cache
		index, err := cache.Register(spriteBlueprint.Location.Key, spr)
		if err != nil {
			return err
		}
		if index > int(ClientConfig.maxSpritesCached.Load()) {
			return errors.New("max sprite error")
		}
		spriteBlueprint.Location.Index.Store(uint32(index))
	}
	return nil
}

// MockSoundLoader implements the sound loading functionality for tests
type MockSoundLoader struct {
	// mu sync.RWMutex
}

// NewMockSoundLoader creates a new MockSoundLoader
func NewMockSoundLoader() *MockSoundLoader {
	return &MockSoundLoader{}
}

// Load implements the sound loading for tests
func (m *MockSoundLoader) Load(soundBundle *blueprintclient.SoundBundle, cache warehouse.Cache[Sound]) error {
	for i := range soundBundle.Blueprints {
		soundBlueprint := &soundBundle.Blueprints[i]
		if soundBlueprint.Location.Key == "" {
			continue
		}

		// First check if already in cache
		soundIndex, ok := cache.GetIndex(soundBlueprint.Location.Key)
		if ok {
			soundBlueprint.Location.Index.Store(uint32(soundIndex))
			continue
		}

		// Check if we have a mock for this sound
		// Create a new mock sound on demand
		snd := Sound{
			name:     soundBlueprint.Location.Key,
			rawData:  []byte{1, 2, 3, 4}, // Dummy data
			audioCtx: nil,
			players:  nil,
		}

		// Register in cache
		index, err := cache.Register(soundBlueprint.Location.Key, snd)
		if err != nil {
			return err
		}
		soundBlueprint.Location.Index.Store(uint32(index))
		if index > int(ClientConfig.maxSpritesCached.Load()) {
			return errors.New("max sounds error")
		}
	}
	return nil
}

// NewTestClient creates a client specifically for testing purposes
// with mock asset loaders that don't require real files
func NewTestClient(baseResX, baseResY, maxSpritesCached, maxSoundsCached, maxScenesCached int) Client {
	cli := &client{
		tickManager:   newTickManager(),
		cameraUtility: newCameraUtility(),
		systemManager: &systemManager{},
		configManager: newConfigManager(),
		sceneManager:  newSceneManager(maxScenesCached),
		assetManager:  newAssetManager(nil),
	}
	cli.assetManager.SpriteLoader = NewMockSpriteLoader()
	cli.assetManager.SoundLoader = NewMockSoundLoader()
	cli.inputManager = newInputManager(cli)
	ClientConfig.maxSoundsCached.Store(uint32(maxSoundsCached))
	ClientConfig.maxSpritesCached.Store(uint32(maxSpritesCached))
	ClientConfig.baseResolution.x = baseResX
	ClientConfig.baseResolution.y = baseResY
	ClientConfig.resolution.x = baseResX
	ClientConfig.resolution.y = baseResY
	ClientConfig.windowSize.x = baseResX
	ClientConfig.windowSize.y = baseResY
	ebiten.SetWindowSize(baseResX, baseResY)
	return cli
}
