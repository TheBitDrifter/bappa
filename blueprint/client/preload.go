package client

const (
	PreloadSprite PreloadAssetType = iota
	PreloadSound
)

type (
	PreloadAssetType int

	PreLoadAssetBlueprint struct {
		Path string
		Type PreloadAssetType

		// Only for sounds
		AudioPlayerCount int
	}

	PreLoadAssetBundle []PreLoadAssetBlueprint
)

func NewPreLoadBlueprint() *PreLoadAssetBundle {
	return &PreLoadAssetBundle{}
}

func (p *PreLoadAssetBundle) AddSprite(path string) *PreLoadAssetBundle {
	*p = append(*p, PreLoadAssetBlueprint{
		Path: path,
		Type: PreloadSprite,
	})
	return p
}

func (p *PreLoadAssetBundle) AddSoundFromPath(path string, playerCount int) *PreLoadAssetBundle {
	*p = append(*p, PreLoadAssetBlueprint{
		Path:             path,
		Type:             PreloadSound,
		AudioPlayerCount: playerCount,
	})
	return p
}

func (p *PreLoadAssetBundle) AddSound(sound SoundConfig) *PreLoadAssetBundle {
	*p = append(*p, PreLoadAssetBlueprint{
		Path:             sound.Path,
		Type:             PreloadSound,
		AudioPlayerCount: sound.AudioPlayerCount,
	})
	return p
}
