package client

import "github.com/TheBitDrifter/blueprint/vector"

// ParallaxBackground contains configuration for translating background layers based on camera positions
// Each RelativeTranslation maps to a coldbrew.Camera
type ParallaxBackground struct {
	// RelativeTranslations stores position offsets for each camera view
	RelativeTranslations [MaxSplit]vector.Two
	// SpeedX horizontal parallax movement multiplier
	SpeedX float64
	// SpeedY vertical parallax movement multiplier
	SpeedY float64

	DisableLooping bool
}
