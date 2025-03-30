package main

import (
	"embed"
	"log"

	"github.com/TheBitDrifter/blueprint"
	"github.com/TheBitDrifter/coldbrew"
	coldbrew_clientsystems "github.com/TheBitDrifter/coldbrew/clientsystems"
	coldbrew_rendersystems "github.com/TheBitDrifter/coldbrew/rendersystems"
	"github.com/TheBitDrifter/warehouse"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"

	blueprintclient "github.com/TheBitDrifter/blueprint/client"
	blueprintinput "github.com/TheBitDrifter/blueprint/input"
	blueprintspatial "github.com/TheBitDrifter/blueprint/spatial"
)

//go:embed assets/*
var assets embed.FS

const (
	sceneOneName = "s1"
	sceneTwoName = "s2"
)

func main() {
	client := coldbrew.NewClient(
		640,
		360,
		10,
		10,
		10,
		assets,
	)

	client.SetTitle("Scene Change Example")
	client.SetMinimumLoadTime(20)

	err := client.RegisterScene(
		sceneOneName,
		640,
		360,
		sceneOnePlan,
		[]coldbrew.RenderSystem{},
		[]coldbrew.ClientSystem{},
		[]blueprint.CoreSystem{},
	)
	if err != nil {
		log.Fatal(err)
	}
	err = client.RegisterScene(
		sceneTwoName,
		640,
		360,
		sceneTwoPlan,
		[]coldbrew.RenderSystem{},
		[]coldbrew.ClientSystem{},
		[]blueprint.CoreSystem{},
	)
	if err != nil {
		log.Fatal(err)
	}

	client.RegisterGlobalRenderSystem(coldbrew_rendersystems.GlobalRenderer{})
	client.RegisterGlobalClientSystem(basicTransferSystem{}, &coldbrew_clientsystems.CameraSceneAssignerSystem{})

	client.ActivateCamera()

	if err := client.Start(); err != nil {
		log.Fatal(err)
	}
}

func sceneOnePlan(height, width int, sto warehouse.Storage) error {
	spriteArchetype, err := sto.NewOrExistingArchetype(
		blueprintspatial.Components.Position,
		blueprintclient.Components.SpriteBundle,
		blueprintclient.Components.CameraIndex,
	)
	if err != nil {
		return err
	}

	err = spriteArchetype.Generate(1,
		blueprintinput.Components.InputBuffer,

		blueprintspatial.NewPosition(255, 20),
		blueprintclient.NewSpriteBundle().
			AddSprite("sprite.png", true),

		blueprintclient.CameraIndex(0),
	)
	err = blueprint.NewParallaxBackgroundBuilder(sto).
		AddLayer("sky.png", 0.1, 0.1).
		AddLayer("far.png", 0.3, 0.3).
		AddLayer("mid.png", 0.4, 0.4).
		AddLayer("near.png", 0.8, 0.8).
		Build()
	if err != nil {
		return err
	}
	return nil
}

func sceneTwoPlan(height, width int, sto warehouse.Storage) error {
	err := blueprint.NewParallaxBackgroundBuilder(sto).
		AddLayer("sky.png", 0.1, 0.1).
		Build()
	if err != nil {
		return err
	}
	return nil
}

type transfer struct {
	targetSceneName string
	playerEntity    warehouse.Entity
}

type basicTransferSystem struct{}

func (basicTransferSystem) Run(cli coldbrew.Client) error {
	var pending []transfer

	for activeScene := range cli.ActiveScenes() {
		if !activeScene.Ready() {
			continue
		}
		cursor := activeScene.NewCursor(blueprint.Queries.CameraIndex)
		for range cursor.Next() {
			if inpututil.IsKeyJustPressed(ebiten.Key1) {

				currentPlayerEntity, err := cursor.CurrentEntity()
				if err != nil {
					return err
				}

				// --- Determine target scene ---
				// Simple toggle between scenes
				var sceneTargetName string
				if activeScene.Name() == sceneOneName {
					sceneTargetName = sceneTwoName
				} else {
					sceneTargetName = sceneOneName
				}

				transfer := transfer{
					targetSceneName: sceneTargetName,
					playerEntity:    currentPlayerEntity,
				}
				pending = append(pending, transfer)
			}
		}
	}

	for _, transfer := range pending {
		cli.ChangeSceneByName(transfer.targetSceneName, transfer.playerEntity)
		// Note:
		// Use cli.ActivateSceneByName() if doing multi scene setups!
	}

	return nil
}
