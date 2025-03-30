package client

import (
	"errors"

	"github.com/TheBitDrifter/blueprint/vector"
	"github.com/TheBitDrifter/warehouse"
)

// SpriteBlueprint defines visual representation of entities with optional animations
// Like SoundBlueprint, it allows for entity creation without direct sprite dependencies,
// enabling the same entities to function in both client contexts (where sprites render)
// and server contexts (where rendering systems aren't active)
type SpriteBlueprint struct {
	// Location reference to the sprite resource in the cache
	Location warehouse.CacheLocation
	// Animations collection of animation data for this sprite
	Animations [ANIM_LIMIT]AnimationData
	// Config contains rendering and animation settings
	Config struct {
		// Offset position adjustment from entity location
		Offset vector.Two
		// Priority rendering order (higher renders on top)
		Priority int
		// Active whether the sprite should be rendered
		// Static whether the sprite should be affected by camera movement
		// IgnoreDefaultRenderer whether to skip the default rendering pipeline
		Active, Static, IgnoreDefaultRenderer bool
		// ActiveAnimIndex currently playing animation index
		ActiveAnimIndex int
		hasAnim         bool
	}
	// A tilset contains the layout to render the blueprint's sprites as a tiles (cached subimages)
	TileSet []Tile
}

// RegisterAnimations adds animations to the sprite
// Panics if animations exceed the predefined limit
func (s *SpriteBlueprint) RegisterAnimations(anims ...AnimationData) {
	if len(anims) > ANIM_LIMIT {
		panic("todo error sig")
	}
	// I'm not sure how to use copy with array vs slice
	for i, anim := range anims {
		s.Animations[i] = anim
	}
}

// TryAnimation changes the active animation if not already playing
// Resets the previous animation's state
func (s *SpriteBlueprint) TryAnimation(anim AnimationData) {
	current := s.Config.ActiveAnimIndex
	for i, bpAnim := range s.Animations {
		if anim.Name == bpAnim.Name && i != current {
			s.Animations[s.Config.ActiveAnimIndex].StartTick = 0 // Reset first
			s.Config.ActiveAnimIndex = i
		}
	}
}

// Set changes the active animation
// Resets the previous animation's state if needed
func (s *SpriteBlueprint) SetAnimation(anim AnimationData) {
	current := s.Config.ActiveAnimIndex

	for i, bpAnim := range s.Animations {
		if anim.Name == bpAnim.Name && i != current {
			s.Animations[s.Config.ActiveAnimIndex].StartTick = 0 // Reset first
			s.Config.ActiveAnimIndex = i
		}
	}
}

func (s *SpriteBlueprint) TryAnimationFromIndex(index int) {
	if index == s.Config.ActiveAnimIndex {
		return
	}
	s.Animations[s.Config.ActiveAnimIndex].StartTick = 0 // Reset first
	s.Config.ActiveAnimIndex = index
}

// HasAnimations returns whether this sprite has animations registered
func (s *SpriteBlueprint) HasAnimations() bool {
	return s.Config.hasAnim
}

func (sb SpriteBlueprint) GetAnim(anim AnimationData) (AnimationData, error) {
	for _, sbAnim := range sb.Animations {
		if anim.Name == sbAnim.Name {
			return anim, nil
		}
	}
	return AnimationData{}, errors.New("animation not found")
}

func (sb *SpriteBlueprint) Activate() {
	sb.Config.Active = true
}

func (sb *SpriteBlueprint) Deactivate() {
	sb.Config.Active = false
}
