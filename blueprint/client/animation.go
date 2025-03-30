package client

import "github.com/TheBitDrifter/blueprint/vector"

// AnimationData contains configuration for sprite-based animations
type AnimationData struct {
	// Name is the animation name
	Name string

	// PositionOffset represents the sprite's displacement from the entity's position
	PositionOffset vector.Two

	// RowIndex indicates which row in the sprite sheet to use
	RowIndex int

	// FrameWidth specifies the width of each animation frame in pixels
	FrameWidth int

	// FrameHeight specifies the height of each animation frame in pixels
	FrameHeight int

	// FrameCount defines how many frames are in this animation
	FrameCount int

	// Speed controls how quickly the animation plays (how many ticks per frame)
	Speed int

	// StartTick defines when the animation begins
	StartTick int

	// Freeze indicates whether the animation should stay on the last frame once finished
	Freeze bool
}
