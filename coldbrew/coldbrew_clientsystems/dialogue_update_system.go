package coldbrew_clientsystems

import (
	"errors"

	"github.com/TheBitDrifter/bappa/blueprint/dialogue"
	bptext "github.com/TheBitDrifter/bappa/coldbrew/text"

	"github.com/TheBitDrifter/bappa/coldbrew"
	"github.com/TheBitDrifter/bappa/warehouse"
	"github.com/hajimehoshi/ebiten/v2/text/v2"
)

type DialogueTextSystem struct {
	TEXT_REVEAL_DELAY_IN_TICKS int
	FONT_SIZE                  int
	FONT_FACE                  *text.GoTextFace
	MAX_LINE_WIDTH             int
}

func (sys DialogueTextSystem) validate() error {
	if sys.FONT_SIZE == 0 {
		return errors.New("missing fontsize dialogueUpdateSystem")
	}
	if sys.FONT_FACE == nil {
		return errors.New("missing fontface dialogueUpdateSystem")
	}
	if sys.MAX_LINE_WIDTH == 0 {
		return errors.New("missing max line width dialogueUpdateSystem")
	}
	return nil
}

func (sys DialogueTextSystem) Run(cli coldbrew.LocalClient, scene coldbrew.Scene) error {
	err := sys.validate()
	if err != nil {
		return err
	}
	query := warehouse.Factory.NewQuery().And(dialogue.Components.Conversation)
	cursor := scene.NewCursor(query)
	currentTick := scene.CurrentTick()

	for range cursor.Next() {
		convo := dialogue.Components.Conversation.GetFromCursor(cursor)

		if convo.AnimationState.RevealStartTick == 0 {
			convo.AnimationState.RevealStartTick = currentTick
		}

		finished, count := bptext.CurrentIndexInTextReveal(
			convo.AnimationState.RevealStartTick,
			currentTick,
			sys.TEXT_REVEAL_DELAY_IN_TICKS,
			convo.AnimationState.WrappedText,
		)

		fullRunes := []rune(convo.AnimationState.WrappedText)
		convo.AnimationState.DisplayedText = string(fullRunes[:count])

		if finished && convo.AnimationState.FinalUpdateTick == 0 {
			convo.AnimationState.FinalUpdateTick = currentTick
		}
	}
	return nil
}

func InitConversation(scene coldbrew.Scene, convo *dialogue.Conversation, fontFace *text.GoTextFace, maxWidth float64) bool {
	if convo.ActiveSlideIndex >= len(convo.Slides) {
		return false
	}
	convo.AnimationState.IsRevealing = true
	convo.WrappedText = bptext.WrapText(convo.Slides[convo.ActiveSlideIndex].Text, fontFace, float64(maxWidth))
	convo.AnimationState.FinalUpdateTick = 0
	convo.DisplayedText = ""

	return true
}
