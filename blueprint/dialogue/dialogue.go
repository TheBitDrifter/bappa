package dialogue

import "github.com/TheBitDrifter/bappa/warehouse"

type Conversation struct {
	Slides           Slides
	ActiveSlideIndex int
	AnimationState
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
	OwnerName     string
	PortraitIndex PortraitEnum
	Text          string
}

type PortraitEnum int

type comps struct {
	Conversation warehouse.AccessibleComponent[Conversation]
}

var Components = comps{
	Conversation: warehouse.FactoryNewComponent[Conversation](),
}
