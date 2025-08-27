package coldbrew_clientsystems

import (
	"github.com/TheBitDrifter/bappa/blueprint/client"
	"github.com/TheBitDrifter/bappa/blueprint/dialogue"
	"github.com/TheBitDrifter/bappa/coldbrew"
	"github.com/TheBitDrifter/bappa/coldbrew/text"
	"github.com/TheBitDrifter/bappa/warehouse"
)

type DialogueSoundSystem struct {
	Volume                     float64
	TEXT_REVEAL_DELAY_IN_TICKS int
	SoundOnWord                bool
}

func (sys DialogueSoundSystem) Run(cli coldbrew.LocalClient, scene coldbrew.Scene) error {
	query := warehouse.Factory.NewQuery().And(
		client.Components.SoundBundle,
		dialogue.Components.Conversation,
	)
	cursor := scene.NewCursor(query)
	currentTick := scene.CurrentTick()

	for range cursor.Next() {
		convo := dialogue.Components.Conversation.GetFromCursor(cursor)
		slides := dialogue.SlidesRegistry[convo.SlidesID]
		tpc := sys.TEXT_REVEAL_DELAY_IN_TICKS
		if slides[convo.ActiveSlideIndex].CustomSpeedTicks != 0 {
			tpc = slides[convo.ActiveSlideIndex].CustomSpeedTicks
		}

		var play bool
		if sys.SoundOnWord {
			play = text.ShouldPlayRevealSoundForWord(convo.AnimationState.RevealStartTick, currentTick, tpc, convo.DisplayedText)
		} else {
			play = text.ShouldPlayRevealSound(convo.AnimationState.RevealStartTick, currentTick, tpc, convo.DisplayedText)
		}

		if play {

			revealedText := convo.AnimationState.DisplayedText
			if len(revealedText) == 0 {
				continue
			}
			runes := []rune(revealedText)
			lastChar := runes[len(runes)-1]

			if lastChar == ' ' {
				continue
			}

			soundBundle := client.Components.SoundBundle.GetFromCursor(cursor)
			sounds := coldbrew.MaterializeSounds(soundBundle)
			if len(sounds) > 0 {
				player := sounds[0].GetAny()
				player.SetVolume(sys.Volume)
				player.Rewind()
				player.Play()
			}
		}
	}
	return nil
}
