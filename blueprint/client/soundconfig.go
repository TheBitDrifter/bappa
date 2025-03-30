package client

// SoundConfig provides configuration settings for sound resources
type SoundConfig struct {
	// Path location of the sound file
	Path string
	// AudioPlayerCount number of simultaneous instances that can play this sound
	AudioPlayerCount int
}

// NewSoundConfig creates a new sound configuration with specified path and player count
// The player count determines how many instances of this sound can play simultaneously
func NewSoundConfig(path string, audioPlayerCount int) SoundConfig {
	return SoundConfig{Path: path, AudioPlayerCount: audioPlayerCount}
}
