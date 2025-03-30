package main

import (
	"embed"
	"log"

	"github.com/TheBitDrifter/blueprint"
	"github.com/TheBitDrifter/coldbrew"
	coldbrew_clientsystems "github.com/TheBitDrifter/coldbrew/clientsystems"
	coldbrew_rendersystems "github.com/TheBitDrifter/coldbrew/rendersystems"
	"github.com/TheBitDrifter/warehouse"
)

//go:embed assets/*
var assets embed.FS

func main() {
	client := coldbrew.NewClient(
		640,
		360,
		10,
		10,
		10,
		assets,
	)
	client.SetTitle("Split Screen Camera")
	err := client.RegisterScene(
		"Example Scene",
		640,
		360,
		exampleScenePlan,
		[]coldbrew.RenderSystem{},
		[]coldbrew.ClientSystem{
			&coldbrew_clientsystems.BackgroundScrollSystem{},
			&cameraMovementSystem{},
		},
		[]blueprint.CoreSystem{},
	)
	if err != nil {
		log.Fatal(err)
	}
	client.RegisterGlobalRenderSystem(coldbrew_rendersystems.GlobalRenderer{})

	client.SetCameraBorderSize(6)

	quarterWidth := 640 / 2
	quarterHeight := 360 / 2
	cam1, err := client.ActivateCamera()
	if err != nil {
		log.Fatal(err)
	}
	cam2, err := client.ActivateCamera()
	if err != nil {
		log.Fatal(err)
	}
	cam3, err := client.ActivateCamera()
	if err != nil {
		log.Fatal(err)
	}
	cam4, err := client.ActivateCamera()
	if err != nil {
		log.Fatal(err)
	}
	cam1.SetDimensions(quarterWidth, quarterHeight)
	cam2.SetDimensions(quarterWidth, quarterHeight)
	cam3.SetDimensions(quarterWidth, quarterHeight)
	cam4.SetDimensions(quarterWidth, quarterHeight)
	cam1Pos, _ := cam1.Positions()
	cam1Pos.X = 0
	cam1Pos.Y = 0
	cam2Pos, cam2ScenePos := cam2.Positions()
	cam2Pos.X = float64(quarterWidth)
	cam2Pos.Y = 0
	cam2ScenePos.Y = 240
	cam3Pos, cam3ScenePos := cam3.Positions()
	cam3Pos.X = 0
	cam3Pos.Y = float64(quarterHeight)
	cam3ScenePos.Y = 160
	cam4Pos, cam4ScenePos := cam4.Positions()
	cam4Pos.X = float64(quarterWidth)
	cam4Pos.Y = float64(quarterHeight)
	cam4ScenePos.Y = 80
	if err := client.Start(); err != nil {
		log.Fatal(err)
	}
}

func exampleScenePlan(height, width int, sto warehouse.Storage) error {
	err := blueprint.NewParallaxBackgroundBuilder(sto).
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

type cameraMovementSystem struct {
	flipX bool
}

func (sys *cameraMovementSystem) Run(cli coldbrew.LocalClient, scene coldbrew.Scene) error {
	cameras := cli.ActiveCamerasFor(scene)
	for _, cam := range cameras {
		_, cameraPositionInScene := cam.Positions()
		if cameraPositionInScene.X > 640 {
			sys.flipX = true
		} else if cameraPositionInScene.X < 0 {
			sys.flipX = false
		}

		if !sys.flipX {
			cameraPositionInScene.X += 1
		} else {
			cameraPositionInScene.X -= 1
		}
	}
	return nil
}
