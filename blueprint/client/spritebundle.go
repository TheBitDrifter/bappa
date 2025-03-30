package client

import (
	"fmt"

	"github.com/TheBitDrifter/blueprint/vector"
	"github.com/TheBitDrifter/warehouse"
)

// SpriteBundle stores a collection of sprite blueprints up to a predefined limit
type SpriteBundle struct {
	index int
	// Blueprints collection of sprite resources that can be referenced by entities
	Blueprints [SPRITE_LIMIT]SpriteBlueprint
}

// NewSpriteBundle creates an empty sprite bundle
func NewSpriteBundle() SpriteBundle {
	return SpriteBundle{}
}

// AddSprite adds a new sprite blueprint to the bundle
// Returns a new bundle with the added sprite
func (sb SpriteBundle) AddSprite(path string, active bool) SpriteBundle {
	if sb.index >= SPRITE_LIMIT {
		panic("Sprite limit exceeded")
	}
	sb.Blueprints[sb.index] = SpriteBlueprint{
		Location: warehouse.CacheLocation{
			Key: path,
		},
	}
	sb.Blueprints[sb.index].Config.Active = active
	sb.index++
	return sb
}

// WithAnimations adds animations to the most recently added sprite
// Returns the updated bundle
func (sb SpriteBundle) WithAnimations(anims ...AnimationData) SpriteBundle {
	if sb.index == 0 {
		panic("No sprite to add animations to")
	}
	blueprint := &sb.Blueprints[sb.index-1]
	for i, anim := range anims {
		if i >= ANIM_LIMIT {
			panic("Animation limit exceeded")
		}
		blueprint.Animations[i] = anim
		blueprint.Config.hasAnim = true
	}
	return sb
}

// WithOffset adds a position offset to the most recently added sprite
// Returns the updated bundle
func (sb SpriteBundle) WithOffset(offset vector.Two) SpriteBundle {
	if sb.index == 0 {
		panic("No sprite to add offset to")
	}
	blueprint := &sb.Blueprints[sb.index-1]
	blueprint.Config.Offset = offset
	return sb
}

// WithPriority sets the rendering priority of the most recently added sprite
// Returns the updated bundle
func (sb SpriteBundle) WithPriority(prio int) SpriteBundle {
	if sb.index == 0 {
		panic("No sprite to prioritize")
	}
	blueprint := &sb.Blueprints[sb.index-1]
	blueprint.Config.Priority = prio
	return sb
}

// WithStatic marks the most recently added sprite as static (not affected by camera)
// Returns the updated bundle
func (sb SpriteBundle) WithStatic(static bool) SpriteBundle {
	if sb.index == 0 {
		panic("No sprite to add animations to")
	}
	blueprint := &sb.Blueprints[sb.index-1]
	blueprint.Config.Static = true
	return sb
}

// WithCustomRenderer marks the most recently added sprite to ignore default rendering
// Returns the updated bundle
func (sb SpriteBundle) WithCustomRenderer() SpriteBundle {
	if sb.index == 0 {
		panic("no sprite to ignore")
	}
	blueprint := &sb.Blueprints[sb.index-1]
	blueprint.Config.IgnoreDefaultRenderer = true
	return sb
}

// Count returns the number of sprites in the bundle
func (sb SpriteBundle) Count() int {
	return sb.index
}

func (sb SpriteBundle) SetActiveAnimation(anim AnimationData) SpriteBundle {
	if sb.index == 0 {
		panic("No sprite to add animations to")
	}
	blueprint := &sb.Blueprints[sb.index-1]
	if !blueprint.Config.hasAnim {
		panic("sprite has no animations")
	}
	match := false
	for i, bpAnim := range blueprint.Animations {
		if anim.Name == bpAnim.Name {
			blueprint.Config.ActiveAnimIndex = i
			match = true
		}
	}
	if !match {
		panic(fmt.Errorf("animation not found: %s", anim.Name))
	}
	return sb
}

func (sb SpriteBundle) SetActiveAnimationFromIndex(index int) SpriteBundle {
	if sb.index == 0 {
		panic("No sprite to add animations to")
	}
	blueprint := &sb.Blueprints[sb.index-1]
	if !blueprint.Config.hasAnim {
		panic("sprite has no animations")
	}
	blueprint.Config.ActiveAnimIndex = index
	return sb
}
