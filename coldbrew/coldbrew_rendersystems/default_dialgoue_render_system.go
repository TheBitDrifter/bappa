package coldbrew_rendersystems

import (
	"github.com/TheBitDrifter/bappa/blueprint/client"
	"github.com/TheBitDrifter/bappa/blueprint/dialogue"
	"github.com/TheBitDrifter/bappa/blueprint/vector"
	"github.com/TheBitDrifter/bappa/coldbrew"
	"github.com/TheBitDrifter/bappa/tteokbokki/spatial"
	"github.com/TheBitDrifter/bappa/warehouse"
	"github.com/hajimehoshi/ebiten/v2/text/v2"
)

type DefaultDialogueRenderSystem struct {
	FONT_FACE               *text.GoTextFace
	PORTRAIT_FONT_FACE      *text.GoTextFace
	PADDING_X               int // TEXT PADDING_X
	PADDING_Y               int // TEXT PADDING Y
	PORTRAIT_PADDING_X      int
	PORTRAIT_PADDING_Y      int
	PORTRAIT_TEXT_PADDING_X int
	PORTRAIT_TEXT_PADDING_Y int
}

func (sys DefaultDialogueRenderSystem) validate() {
	if sys.FONT_FACE == nil {
		panic("missing font face defaultDialogueRenderSystem")
	}
}

func (sys DefaultDialogueRenderSystem) Render(scene coldbrew.Scene, screen coldbrew.Screen, cli coldbrew.LocalClient) {
	sys.validate()

	query := warehouse.Factory.NewQuery().And(
		dialogue.Components.Conversation,
		client.Components.SpriteBundle,
		spatial.Components.Position,
	)
	cursor := scene.NewCursor(query)

	cameras := cli.ActiveCamerasFor(scene)

	for _, c := range cameras {
		for range cursor.Next() {
			convo := dialogue.Components.Conversation.GetFromCursor(cursor)
			if convo.ActiveSlideIndex >= len(convo.Slides) {
				return
			}
			activeSlide := convo.Slides[convo.ActiveSlideIndex]

			bundle := client.Components.SpriteBundle.GetFromCursor(cursor)

			dialogueBoxSheet, err := coldbrew.MaterializeSprite(bundle, 0)
			if err != nil {
				panic("dialogue missing box sprites")
			}
			portraitSprite, err := coldbrew.MaterializeSprite(bundle, int(activeSlide.PortraitIndex))
			if err != nil {
				panic("dialogue sys missing portrait sprites")
			}

			pos := spatial.Components.Position.GetFromCursor(cursor)

			RenderSpriteSheetAnimation(
				dialogueBoxSheet,
				&bundle.Blueprints[0],
				bundle.Blueprints[0].Config.ActiveAnimIndex,
				pos.Two,
				0,
				vector.Two{X: 1, Y: 1},
				spatial.NewDirectionRight(),
				vector.Two{},
				true,
				c,
				scene.CurrentTick(),
				nil,
				nil,
			)

			portPos := vector.Two{X: pos.X + float64(sys.PORTRAIT_PADDING_X), Y: pos.Y + float64(sys.PORTRAIT_PADDING_Y)}
			RenderSprite(
				portraitSprite,
				portPos,
				0,
				vector.Two{X: 1, Y: 1},
				vector.Two{},
				spatial.NewDirectionRight(),
				true,
				c,
			)
			if convo.AnimationState.DisplayedText != "" {
				textOpts := &text.DrawOptions{}

				textOpts.LineSpacing = float64(sys.FONT_FACE.Size + 2)

				texPos := vector.Two{X: pos.X + float64(sys.PADDING_X), Y: pos.Y + float64(sys.PADDING_Y)}
				c.DrawTextStatic(
					convo.AnimationState.DisplayedText,
					textOpts,
					sys.FONT_FACE,
					texPos,
				)
			}

			if convo.Slides[convo.ActiveSlideIndex].OwnerName != "" {
				nameText := convo.Slides[convo.ActiveSlideIndex].OwnerName
				textOpts := &text.DrawOptions{}

				textOpts.LineSpacing = float64(sys.PORTRAIT_FONT_FACE.Size + 2)
				texPos := vector.Two{X: pos.X + float64(sys.PORTRAIT_TEXT_PADDING_X), Y: pos.Y + float64(sys.PORTRAIT_TEXT_PADDING_Y)}

				c.DrawTextStatic(
					nameText,
					textOpts,
					sys.PORTRAIT_FONT_FACE,
					texPos,
				)
			}
		}

		c.PresentToScreen(screen, coldbrew.ClientConfig.CameraBorderSize())
	}
}
