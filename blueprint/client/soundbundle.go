package client

import "github.com/TheBitDrifter/warehouse"

// SoundBundle stores a collection of sound blueprints up to a predefined limit
type SoundBundle struct {
	index int
	// Blueprints collection of sound resources that can be referenced by entities
	Blueprints [SOUND_LIMIT]SoundBlueprint
}

// NewSoundBundle creates an empty sound bundle
func NewSoundBundle() SoundBundle {
	return SoundBundle{}
}

// AddSoundFromConfig adds a new sound blueprint to the bundle based on configuration
// Returns a new bundle with the added sound
// Used to define the number of audio players, which determines how many instances
// of a sound can play simultaneously
func (sb SoundBundle) AddSoundFromConfig(soundConfig SoundConfig) SoundBundle {
	if sb.index >= SOUND_LIMIT {
		panic("Sound limit exceeded")
	}
	sb.Blueprints[sb.index] = SoundBlueprint{
		AudioPlayerCount: soundConfig.AudioPlayerCount,
		Location: warehouse.CacheLocation{
			Key: soundConfig.Path,
		},
	}
	sb.index++
	return sb
}

// AddSoundFromPath adds a new sound blueprint to the bundle using a path string
// Creates a blueprint with default settings (single audio player)
// Returns a new bundle with the added sound
func (sb SoundBundle) AddSoundFromPath(path string) SoundBundle {
	if sb.index >= SOUND_LIMIT {
		panic("Sound limit exceeded")
	}
	sb.Blueprints[sb.index] = SoundBlueprint{
		AudioPlayerCount: 1,
		Location: warehouse.CacheLocation{
			Key: path,
		},
	}
	sb.index++
	return sb
}
