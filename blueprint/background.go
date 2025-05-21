package blueprint

import (
	"github.com/TheBitDrifter/bappa/blueprint/client"
	"github.com/TheBitDrifter/bappa/blueprint/vector"
	"github.com/TheBitDrifter/bappa/tteokbokki/spatial"
	"github.com/TheBitDrifter/bappa/warehouse"
)

// ParallaxLayer defines a single layer in a parallax background
type ParallaxLayer struct {
	// SpritePath is the path to the sprite resource
	SpritePath string
	// SpeedX is the horizontal parallax speed multiplier
	SpeedX float64
	// SpeedY is the vertical parallax speed multiplier
	SpeedY float64
}

// ParallaxBackgroundBuilder provides a fluent API for creating parallax backgrounds
type ParallaxBackgroundBuilder struct {
	storage warehouse.Storage
	layers  []ParallaxLayer
	// Optional position offset for the entire background
	offset vector.Two
	// DisableLooping controls whether the background should loop
	disableLooping bool
}

// NewParallaxBackgroundBuilder creates a new builder for parallax backgrounds
func NewParallaxBackgroundBuilder(sto warehouse.Storage) *ParallaxBackgroundBuilder {
	return &ParallaxBackgroundBuilder{
		storage: sto,
		layers:  []ParallaxLayer{},
	}
}

// WithOffset sets an optional position offset for the entire background
func (b *ParallaxBackgroundBuilder) WithOffset(offset vector.Two) *ParallaxBackgroundBuilder {
	b.offset = offset
	return b
}

// WithDisableLooping sets whether background looping should be disabled
func (b *ParallaxBackgroundBuilder) WithDisableLooping(disable bool) *ParallaxBackgroundBuilder {
	b.disableLooping = disable
	return b
}

// AddLayer adds a new layer to the parallax background
func (b *ParallaxBackgroundBuilder) AddLayer(spritePath string, speedX, speedY float64) *ParallaxBackgroundBuilder {
	b.layers = append(b.layers, ParallaxLayer{
		SpritePath: spritePath,
		SpeedX:     speedX,
		SpeedY:     speedY,
	})
	return b
}

// Build generates all layers and creates the parallax background
func (b *ParallaxBackgroundBuilder) Build() error {
	// Create the backgroundArchetype
	backgroundArchetype, err := b.storage.NewOrExistingArchetype(
		client.Components.SpriteBundle,
		client.Components.ParallaxBackground,
		spatial.Components.Position,
	)
	if err != nil {
		return err
	}
	// Handle empty layer list
	if len(b.layers) == 0 {
		return nil
	}
	// Generate each layer from the provided slice
	for _, layer := range b.layers {

		pos := spatial.NewPosition(0, 0)
		sprite := client.NewSpriteBundle().AddSprite(layer.SpritePath, true)

		if b.offset.X != 0 || b.offset.Y != 0 {
			pos.X = b.offset.X
			pos.Y = b.offset.Y
		}
		err = backgroundArchetype.Generate(
			1,
			sprite,
			client.ParallaxBackground{
				SpeedX:         layer.SpeedX,
				SpeedY:         layer.SpeedY,
				DisableLooping: b.disableLooping,
			},
			pos,
		)
		if err != nil {
			return err
		}
	}
	return nil
}

// CreateStillBackground is a utility function for creating a non-parallax (static) background
// Optional position parameters can be provided to offset the background
func CreateStillBackground(sto warehouse.Storage, spritePath string, pos ...vector.Two) error {
	backgroundArchetype, err := sto.NewOrExistingArchetype(
		client.Components.SpriteBundle,
		client.Components.ParallaxBackground,
		spatial.Components.Position,
	)
	if err != nil {
		return err
	}

	spriteBundle := client.NewSpriteBundle().AddSprite(spritePath, true)

	// Apply position offset if provided
	setPos := vector.Two{}
	if len(pos) > 0 {
		setPos.X = pos[0].X
		setPos.Y = pos[0].Y
	}

	return backgroundArchetype.Generate(
		1,
		spriteBundle,
		client.ParallaxBackground{
			SpeedX: 0,
			SpeedY: 0,
			// Static backgrounds typically should not loop
			DisableLooping: true,
		},
		spatial.NewPosition(setPos.X, setPos.Y),
	)
}
