package main

import (
	"embed"
	"log"

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
	Up, Down, Left, Right blueprintinput.Input
}{
	Up:    blueprintinput.NewInput(),
	Down:  blueprintinput.NewInput(),
	Left:  blueprintinput.NewInput(),
	Right: blueprintinput.NewInput(),
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

	client.SetTitle("Capturing Keyboard Inputs")

	err := client.RegisterScene(
		"Example Scene",
		640,
		360,
		exampleScenePlan,
		[]coldbrew.RenderSystem{},
		[]coldbrew.ClientSystem{},
		[]blueprint.CoreSystem{
			inputSystem{},
		},
	)
	if err != nil {
		log.Fatal(err)
	}

	client.RegisterGlobalRenderSystem(coldbrew_rendersystems.GlobalRenderer{})
	client.RegisterGlobalClientSystem(coldbrew_clientsystems.InputBufferSystem{})
	client.ActivateCamera()

	receiver, _ := client.ActivateReceiver()

	receiver.RegisterKey(ebiten.KeyUp, actions.Up)
	receiver.RegisterKey(ebiten.KeyW, actions.Up)

	receiver.RegisterKey(ebiten.KeyDown, actions.Down)
	receiver.RegisterKey(ebiten.KeyS, actions.Down)

	receiver.RegisterKey(ebiten.KeyLeft, actions.Left)
	receiver.RegisterKey(ebiten.KeyA, actions.Left)

	receiver.RegisterKey(ebiten.KeyRight, actions.Right)
	receiver.RegisterKey(ebiten.KeyD, actions.Right)

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

type inputSystem struct{}

func (inputSystem) Run(scene blueprint.Scene, _ float64) error {
	query := warehouse.Factory.NewQuery().
		And(blueprintinput.Components.InputBuffer, blueprintspatial.Components.Position)

	cursor := scene.NewCursor(query)

	for range cursor.Next() {
		pos := blueprintspatial.Components.Position.GetFromCursor(cursor)
		inputBuffer := blueprintinput.Components.InputBuffer.GetFromCursor(cursor)

		if stampedAction, ok := inputBuffer.ConsumeInput(actions.Up); ok {
			log.Println("Tick", stampedAction.Tick)
			pos.Y -= 2
		}
		if stampedAction, ok := inputBuffer.ConsumeInput(actions.Down); ok {
			log.Println("Tick", stampedAction.Tick)
			pos.Y += 2
		}
		if stampedAction, ok := inputBuffer.ConsumeInput(actions.Left); ok {
			log.Println("Tick", stampedAction.Tick)
			pos.X -= 2
		}
		if stampedAction, ok := inputBuffer.ConsumeInput(actions.Right); ok {
			log.Println("Tick", stampedAction.Tick)
			pos.X += 2
		}

	}
	return nil
}
