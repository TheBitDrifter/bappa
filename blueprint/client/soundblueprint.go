package client

import "github.com/TheBitDrifter/warehouse"

// SoundBlueprint defines sound resources that can be loaded conditionally
// It allows for entity creation without direct sound dependencies, enabling
// the same entities to function in both client contexts (where sounds play)
// and server/headless contexts (where sound systems aren't active)
type SoundBlueprint struct {
	// Location reference to the sound resource in the cache
	Location warehouse.CacheLocation
	// AudioPlayerCount number of concurrent audio players to allocate
	AudioPlayerCount int
}
