package dialogue

import (
	"github.com/TheBitDrifter/bappa/blueprint"
	"github.com/TheBitDrifter/bappa/warehouse"
)

type Conversation struct {
	SlidesID         SlidesEnum
	CallbackID       CallbackEnum
	ActiveSlideIndex int
	AnimationState
	PortraitIDForSpriteBundleBlueprintIndex map[PortraitEnum]int
}

type AnimationState struct {
	WrappedText     string
	DisplayedText   string
	FinalUpdateTick int
	RevealStartTick int
	IsRevealing     bool
}

type Slides []Slide

type Slide struct {
	OwnerName        string
	PortraitID       PortraitEnum
	Text             string
	CustomSpeedTicks int
}

type PortraitEnum int

type SlidesEnum int

type CallbackEnum int

var CallbackRegistry map[CallbackEnum]func(scene blueprint.Scene) error = map[CallbackEnum]func(scene blueprint.Scene) error{}

var SlidesRegistry map[SlidesEnum]Slides = map[SlidesEnum]Slides{}

type comps struct {
	Conversation warehouse.AccessibleComponent[Conversation]
}

var Components = comps{
	Conversation: warehouse.FactoryNewComponent[Conversation](),
}
