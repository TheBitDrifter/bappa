package client

import (
	"github.com/TheBitDrifter/warehouse"
)

type defaultComponents struct {
	SpriteBundle       warehouse.AccessibleComponent[SpriteBundle]
	SoundBundle        warehouse.AccessibleComponent[SoundBundle]
	CameraIndex        warehouse.AccessibleComponent[CameraIndex]
	ParallaxBackground warehouse.AccessibleComponent[ParallaxBackground]
}

var Components = defaultComponents{
	CameraIndex:        warehouse.FactoryNewComponent[CameraIndex](),
	ParallaxBackground: warehouse.FactoryNewComponent[ParallaxBackground](),
	SpriteBundle:       warehouse.FactoryNewComponent[SpriteBundle](),
	SoundBundle:        warehouse.FactoryNewComponent[SoundBundle](),
}
