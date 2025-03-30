package main

import (
	"embed"
	"log"
	"math"

	"github.com/TheBitDrifter/blueprint"
	blueprintclient "github.com/TheBitDrifter/blueprint/client"
	blueprintinput "github.com/TheBitDrifter/blueprint/input"
	blueprintspatial "github.com/TheBitDrifter/blueprint/spatial"
	"github.com/TheBitDrifter/coldbrew"
	coldbrew_clientsystems "github.com/TheBitDrifter/coldbrew/clientsystems"
	coldbrew_rendersystems "github.com/TheBitDrifter/coldbrew/rendersystems"
	"github.com/TheBitDrifter/warehouse"
	"github.com/hajimehoshi/ebiten/v2"
)

//go:embed assets/*
var assets embed.FS

var actions = struct {
	Movement blueprintinput.Input
}{
	Movement: blueprintinput.NewInput(),
}

func lerp(start, end, t float64) float64 {
	return start + t*(end-start)
}

func main() {
	client := coldbrew.NewClient(
		640,
		360,
		10,
		10,
		10,
		assets,
	)
	client.SetTitle("Capturing Mouse Inputs")
	err := client.RegisterScene(
		"Example Scene",
		640,
		360,
		exampleScenePlan,
		[]coldbrew.RenderSystem{},
		[]coldbrew.ClientSystem{},
		[]blueprint.CoreSystem{
			&inputSystem{},
		},
	)
	if err != nil {
		log.Fatal(err)
	}
	client.RegisterGlobalRenderSystem(coldbrew_rendersystems.GlobalRenderer{})
	client.RegisterGlobalClientSystem(coldbrew_clientsystems.InputBufferSystem{})
	client.ActivateCamera()
	receiver, _ := client.ActivateReceiver()
	receiver.RegisterMouseButton(ebiten.MouseButtonLeft, actions.Movement)
	if err := client.Start(); err != nil {
		log.Fatal(err)
	}
}

func exampleScenePlan(height, width int, sto warehouse.Storage) error {
	spriteArchetype, err := sto.NewOrExistingArchetype(
		blueprintinput.Components.InputBuffer,
		blueprintspatial.Components.Position,
		blueprintclient.Components.SpriteBundle,
	)
	if err != nil {
		return err
	}
	err = spriteArchetype.Generate(1,
		blueprintinput.Components.InputBuffer,
		blueprintspatial.NewPosition(255, 20),
		blueprintclient.NewSpriteBundle().
			AddSprite("sprite.png", true),
	)
	if err != nil {
		return err
	}
	return nil
}

type inputSystem struct {
	LastMovementX float64
	LastMovementY float64
	HasTarget     bool
}

func (sys *inputSystem) Run(scene blueprint.Scene, dt float64) error {
	query := warehouse.Factory.NewQuery().
		And(blueprintinput.Components.InputBuffer, blueprintspatial.Components.Position)
	cursor := scene.NewCursor(query)

	for range cursor.Next() {
		pos := blueprintspatial.Components.Position.GetFromCursor(cursor)
		inputBuffer := blueprintinput.Components.InputBuffer.GetFromCursor(cursor)

		if stampedMovement, ok := inputBuffer.ConsumeInput(actions.Movement); ok {
			sys.LastMovementX = float64(stampedMovement.X)
			sys.LastMovementY = float64(stampedMovement.Y)
			sys.HasTarget = true
		}

		if sys.HasTarget {
			dx := sys.LastMovementX - pos.X
			dy := sys.LastMovementY - pos.Y
			distance := math.Sqrt(dx*dx + dy*dy)

			if distance < 5 {
				sys.HasTarget = false
			} else {
				lerpFactor := 0.05
				pos.X = lerp(pos.X, sys.LastMovementX, lerpFactor)
				pos.Y = lerp(pos.Y, sys.LastMovementY, lerpFactor)
			}
		}
	}

	return nil
}
