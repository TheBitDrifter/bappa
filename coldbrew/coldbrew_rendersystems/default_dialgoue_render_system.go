package coldbrew_rendersystems

import (
	"log"

	"github.com/TheBitDrifter/bappa/blueprint/client"
	"github.com/TheBitDrifter/bappa/blueprint/dialogue"
	"github.com/TheBitDrifter/bappa/blueprint/vector"
	"github.com/TheBitDrifter/bappa/coldbrew"
	"github.com/TheBitDrifter/bappa/tteokbokki/spatial"
	"github.com/TheBitDrifter/bappa/warehouse"
	"github.com/hajimehoshi/ebiten/v2/text/v2"
)

type DefaultDialogueRenderSystem struct {
	FONT_FACE                *text.GoTextFace
	PORTRAIT_FONT_FACE       *text.GoTextFace
	PADDING_X                int // TEXT PADDING_X
	PADDING_Y                int // TEXT PADDING Y
	PORTRAIT_PADDING_X       int
	PORTRAIT_PADDING_Y       int
	PORTRAIT_TEXT_PADDING_X  int
	PORTRAIT_TEXT_PADDING_Y  int
	PORTRAIT_NAME_BOX_WIDTH  float64
	NEXT_INDICATOR_PADDING_X float64
	NEXT_INDICATOR_PADDING_Y float64
	NEXT_MIN_DELAY           float64
	NEXT_LAST_SHOWN          float64
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
			slides := dialogue.SlidesRegistry[convo.SlidesID]
			if convo.ActiveSlideIndex >= len(slides) {
				return
			}
			activeSlide := slides[convo.ActiveSlideIndex]

			bundle := client.Components.SpriteBundle.GetFromCursor(cursor)

			dialogueBoxSheet, err := coldbrew.MaterializeSprite(bundle, 0)
			if err != nil {
				panic("dialogue missing box sprites")
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

			portDex, portOK := convo.PortraitIDForSpriteBundleBlueprintIndex[activeSlide.PortraitID]
			if portOK {

				portraitSprite, err := coldbrew.MaterializeSprite(bundle, portDex)
				if err != nil {
					log.Println(err, portDex, activeSlide.OwnerName, activeSlide.PortraitID)
					panic("dialogue sys missing portrait sprites")
				}

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

			}

			if len(convo.AnimationState.DisplayedText) >= len(activeSlide.Text) && scene.CurrentTick()-convo.AnimationState.FinalUpdateTick > int(sys.NEXT_MIN_DELAY) {
				nextSheet, err := coldbrew.MaterializeSprite(bundle, 1)
				if err != nil {
					panic("missing next sprite")
				}

				nextPos := vector.Two{X: pos.X + float64(sys.NEXT_INDICATOR_PADDING_X), Y: pos.Y + float64(sys.NEXT_INDICATOR_PADDING_Y)}

				RenderSpriteSheetAnimation(
					nextSheet,
					&bundle.Blueprints[1],
					bundle.Blueprints[1].Config.ActiveAnimIndex,
					nextPos,
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
			}

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

			if slides[convo.ActiveSlideIndex].OwnerName != "" {
				nameText := slides[convo.ActiveSlideIndex].OwnerName

				textOpts := &text.DrawOptions{}
				textOpts.LineSpacing = float64(sys.PORTRAIT_FONT_FACE.Size + 2)

				textWidth, _ := text.Measure(nameText, sys.PORTRAIT_FONT_FACE, textOpts.LineSpacing)
				boxStartX := pos.X + float64(sys.PORTRAIT_TEXT_PADDING_X)

				centeredX := boxStartX + (float64(sys.PORTRAIT_NAME_BOX_WIDTH) / 2) - (textWidth / 2)
				texPos := vector.Two{X: centeredX, Y: pos.Y + float64(sys.PORTRAIT_TEXT_PADDING_Y)}

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
