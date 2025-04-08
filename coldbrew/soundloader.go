package coldbrew

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"sync"

	client "github.com/TheBitDrifter/bappa/blueprint/client"
	"github.com/TheBitDrifter/bappa/environment"
	"github.com/TheBitDrifter/bappa/warehouse"
	"github.com/hajimehoshi/ebiten/v2/audio"
)

var defaultAudioCtx = audio.NewContext(44100)

type SoundLoader interface {
	Load(bundle *client.SoundBundle, cache warehouse.Cache[Sound]) error

	PreLoad(bundle client.PreLoadAssetBundle, cache warehouse.Cache[Sound]) error
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

// Load processes a batch of sound locations and caches them
// It uses the provided cache for lookups and registration
// which enables cache busting when a new cache is provided
func (loader *soundLoader) Load(bundle *client.SoundBundle, cache warehouse.Cache[Sound]) error {
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

		if environment.IsWASM() || environment.IsProd() {
			audioData, err = fs.ReadFile(loader.fs, soundBlueprint.Location.Key)
			if err != nil {
				return fmt.Errorf("failed to read embedded sound %s: %w", soundBlueprint.Location.Key, err)
			}
		} else {
			path := ClientConfig.localAssetPath + soundBlueprint.Location.Key
			audioData, err = os.ReadFile(path)
			if err != nil {
				return fmt.Errorf("failed to read sound file %s: %w", soundBlueprint.Location.Key, err)
			}
		}

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

func (loader *soundLoader) PreLoad(bundle client.PreLoadAssetBundle, cache warehouse.Cache[Sound]) error {
	for i := range bundle {
		preLoadAssetBp := &bundle[i]
		if preLoadAssetBp.Path == "" || preLoadAssetBp.Type != client.PreloadSound {
			continue
		}

		soundIndex, ok := cache.GetIndex(preLoadAssetBp.Path)

		if ok {
			if soundIndex > int(ClientConfig.maxSpritesCached.Load()) {
				return errors.New("max sprites error")
			}
			continue
		}

		// Load sound data
		var audioData []byte
		var err error

		if environment.IsWASM() || environment.IsProd() {
			audioData, err = fs.ReadFile(loader.fs, preLoadAssetBp.Path)
			if err != nil {
				return fmt.Errorf("failed to read embedded sound %s: %w", preLoadAssetBp.Path, err)
			}
		} else {
			path := ClientConfig.localAssetPath + preLoadAssetBp.Path
			audioData, err = os.ReadFile(path)
			if err != nil {
				return fmt.Errorf("failed to read sound file %s: %w", preLoadAssetBp.Path, err)
			}
		}

		snd, err := newSound(preLoadAssetBp.Path, audioData, loader.audioCtx, preLoadAssetBp.AudioPlayerCount)
		if err != nil {
			return fmt.Errorf("failed to create sound %s: %w", preLoadAssetBp.Path, err)
		}

		index, err := cache.Register(preLoadAssetBp.Path, snd)
		if err != nil {
			return err
		}

		if index > int(ClientConfig.maxSoundsCached.Load()) {
			return errors.New("max sounds error")
		}

	}
	return nil
}
