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

	client.SetTitle("Rendering Parallax Background")

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

	client.ActivateCamera()

	if err := client.Start(); err != nil {
		log.Fatal(err)
	}
}

func exampleScenePlan(height, width int, sto warehouse.Storage) error {
	// Use 0.0 and a single layer if you want a still background
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
	flipY bool
}

func (sys *cameraMovementSystem) Run(cli coldbrew.LocalClient, scene coldbrew.Scene) error {
	cam := cli.ActiveCamerasFor(scene)[0]
	_, cameraPositionInScene := cam.Positions()

	if cameraPositionInScene.X > 640 {
		sys.flipX = true
	} else if cameraPositionInScene.X < 0 {
		sys.flipX = false
	}

	if cameraPositionInScene.Y > 100 {
		sys.flipY = true
	} else if cameraPositionInScene.Y < 0 {
		sys.flipY = false
	}

	if !sys.flipX {
		cameraPositionInScene.X += 2
	} else {
		cameraPositionInScene.X -= 2
	}

	if !sys.flipY {
		cameraPositionInScene.Y += 2
	} else {
		cameraPositionInScene.Y -= 2
	}

	return nil
}
