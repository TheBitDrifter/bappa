package coldbrew

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"runtime"
	"sync"

	blueprintclient "github.com/TheBitDrifter/blueprint/client"
	"github.com/TheBitDrifter/warehouse"
	"github.com/hajimehoshi/ebiten/v2/audio"
)

var defaultAudioCtx = audio.NewContext(44100)

type SoundLoader interface {
	Load(bundle *blueprintclient.SoundBundle, cache warehouse.Cache[Sound]) error
}

// soundLoader handles loading and caching of audio files
type soundLoader struct {
	mu       sync.RWMutex
	fs       fs.FS
	audioCtx *audio.Context
}

// NewSoundLoader creates a sound loader with 44.1kHz sample rate
func NewSoundLoader(embeddedFS fs.FS) *soundLoader {
	return &soundLoader{
		fs:       embeddedFS,
		audioCtx: defaultAudioCtx,
	}
}

// isWASM returns true if running in WebAssembly environment
func isWASM() bool {
	return runtime.GOOS == "js" && runtime.GOARCH == "wasm"
}

// Load processes a batch of sound locations and caches them
// It uses the provided cache for lookups and registration
// which enables cache busting when a new cache is provided
func (loader *soundLoader) Load(bundle *blueprintclient.SoundBundle, cache warehouse.Cache[Sound]) error {
	for i := range bundle.Blueprints {
		soundBlueprint := &bundle.Blueprints[i]
		if soundBlueprint.Location.Key == "" {
			continue
		}

		soundIndex, ok := cache.GetIndex(soundBlueprint.Location.Key)

		if ok {
			if soundIndex > int(ClientConfig.maxSpritesCached.Load()) {
				return errors.New("max sprites error")
			}
			soundBlueprint.Location.Index.Store(uint32(soundIndex))
			continue
		}

		// Load sound data
		var audioData []byte
		var err error

		// Always use embedded assets when in WASM or production mode
		if isWASM() || isProd {
			// Load from embedded assets
			audioData, err = fs.ReadFile(loader.fs, filepath.Join("assets/sounds", soundBlueprint.Location.Key))
			if err != nil {
				return fmt.Errorf("failed to read embedded sound %s: %w", soundBlueprint.Location.Key, err)
			}
		} else {
			// Development mode (non-WASM): load from filesystem
			audioData, err = os.ReadFile(fmt.Sprintf("assets/sounds/%s", soundBlueprint.Location.Key))
			if err != nil {
				return fmt.Errorf("failed to read sound file %s: %w", soundBlueprint.Location.Key, err)
			}
		}

		// Create a new sound (always with pooling enabled)
		snd, err := newSound(soundBlueprint.Location.Key, audioData, loader.audioCtx, soundBlueprint.AudioPlayerCount)
		if err != nil {
			return fmt.Errorf("failed to create sound %s: %w", soundBlueprint.Location.Key, err)
		}

		index, err := cache.Register(soundBlueprint.Location.Key, snd)
		if err != nil {
			return err
		}

		if index > int(ClientConfig.maxSoundsCached.Load()) {
			return errors.New("max sounds error")
		}

		soundBlueprint.Location.Index.Store(uint32(index))
	}
	return nil
}
