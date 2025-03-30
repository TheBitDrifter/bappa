package main

import (
	"embed"
	"log"

	"github.com/TheBitDrifter/blueprint"
	blueprintclient "github.com/TheBitDrifter/blueprint/client"
	blueprintspatial "github.com/TheBitDrifter/blueprint/spatial"
	"github.com/TheBitDrifter/coldbrew"
	coldbrew_rendersystems "github.com/TheBitDrifter/coldbrew/rendersystems"
	"github.com/TheBitDrifter/warehouse"
)

//go:embed assets/*
var assets embed.FS

var idleAnimation = blueprintclient.AnimationData{
	Name:        "idle",
	RowIndex:    0,
	FrameCount:  6,
	FrameWidth:  144,
	FrameHeight: 116,
	Speed:       8,
}

func main() {
	client := coldbrew.NewClient(
		320,
		180,
		10,
		10,
		10,
		assets,
	)

	client.SetTitle("Animating a Sprite Sheet")

	err := client.RegisterScene(
		"Example Scene",
		320,
		180,
		exampleScenePlan,
		[]coldbrew.RenderSystem{},
		[]coldbrew.ClientSystem{},
		[]blueprint.CoreSystem{},
	)
	if err != nil {
		log.Fatal(err)
	}

	client.RegisterGlobalRenderSystem(coldbrew_rendersystems.GlobalRenderer{})

	client.ActivateCamera()

	if err := client.Start(); err != nil {
		log.Fatal(err)
	}
}

func exampleScenePlan(height, width int, sto warehouse.Storage) error {
	spriteArchetype, err := sto.NewOrExistingArchetype(
		blueprintspatial.Components.Position,
		blueprintclient.Components.SpriteBundle,
	)
	if err != nil {
		return err
	}
	err = spriteArchetype.Generate(1,
		blueprintspatial.NewPosition(90, 20),
		blueprintclient.NewSpriteBundle().
			AddSprite("sprite_sheet.png", true).
			WithAnimations(idleAnimation),
	)
	if err != nil {
		return err
	}
	return nil
}
