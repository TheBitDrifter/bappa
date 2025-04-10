package main

import (
	"embed"
	"log"

	"github.com/TheBitDrifter/bappa/blueprint"
	"github.com/TheBitDrifter/bappa/blueprint/client"
	"github.com/TheBitDrifter/bappa/blueprint/input"
	"github.com/TheBitDrifter/bappa/coldbrew"
	"github.com/TheBitDrifter/bappa/coldbrew/coldbrew_clientsystems"
	"github.com/TheBitDrifter/bappa/coldbrew/coldbrew_rendersystems"
	"github.com/TheBitDrifter/bappa/tteokbokki/spatial"

	"github.com/TheBitDrifter/bappa/warehouse"
	"github.com/hajimehoshi/ebiten/v2"
)

//go:embed assets/*
var assets embed.FS

var actions = struct {
	Up, Down, Left, Right input.Input
}{
	Up:    input.NewInput(),
	Down:  input.NewInput(),
	Left:  input.NewInput(),
	Right: input.NewInput(),
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
		input.Components.InputBuffer,
		spatial.Components.Position,
		client.Components.SpriteBundle,
	)
	if err != nil {
		return err
	}

	err = spriteArchetype.Generate(1,
		input.Components.InputBuffer,

		spatial.NewPosition(255, 20),
		client.NewSpriteBundle().
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
		And(input.Components.InputBuffer, spatial.Components.Position)

	cursor := scene.NewCursor(query)

	for range cursor.Next() {
		pos := spatial.Components.Position.GetFromCursor(cursor)
		inputBuffer := input.Components.InputBuffer.GetFromCursor(cursor)

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
