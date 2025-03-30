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
	Movement, Up, Down, Left, Right blueprintinput.Input
}{
	Movement: blueprintinput.NewInput(),
	Up:       blueprintinput.NewInput(),
	Down:     blueprintinput.NewInput(),
	Left:     blueprintinput.NewInput(),
	Right:    blueprintinput.NewInput(),
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
	client.SetTitle("Capturing Gamepad Inputs")
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
	receiver.RegisterPad(0)
	receiver.RegisterGamepadAxes(true, actions.Movement)

	receiver.RegisterGamepadButton(ebiten.GamepadButton4, actions.Up)
	receiver.RegisterGamepadButton(ebiten.GamepadButton0, actions.Down)

	receiver.RegisterGamepadButton(ebiten.GamepadButton1, actions.Right)
	receiver.RegisterGamepadButton(ebiten.GamepadButton3, actions.Left)

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
	StickX float64
	StickY float64
}

func (sys *inputSystem) Run(scene blueprint.Scene, dt float64) error {
	query := warehouse.Factory.NewQuery().
		And(blueprintinput.Components.InputBuffer, blueprintspatial.Components.Position)

	cursor := scene.NewCursor(query)

	for range cursor.Next() {
		pos := blueprintspatial.Components.Position.GetFromCursor(cursor)
		inputBuffer := blueprintinput.Components.InputBuffer.GetFromCursor(cursor)

		if stampedMovement, ok := inputBuffer.ConsumeInput(actions.Movement); ok {
			sys.StickX = float64(stampedMovement.X)
			sys.StickY = float64(stampedMovement.Y)

			magnitude := math.Sqrt(sys.StickX*sys.StickX + sys.StickY*sys.StickY)
			deadzone := 0.15

			if magnitude > deadzone {
				if magnitude > 1.0 {
					sys.StickX /= magnitude
					sys.StickY /= magnitude
				}

				const moveSpeed = 3.0
				pos.X += sys.StickX * moveSpeed
				pos.Y -= sys.StickY * moveSpeed
			}
		}

		moveSpeed := 2.0

		if _, ok := inputBuffer.ConsumeInput(actions.Up); ok {
			pos.Y -= moveSpeed
		}
		if _, ok := inputBuffer.ConsumeInput(actions.Down); ok {
			pos.Y += moveSpeed
		}
		if _, ok := inputBuffer.ConsumeInput(actions.Left); ok {
			pos.X -= moveSpeed
		}
		if _, ok := inputBuffer.ConsumeInput(actions.Right); ok {
			pos.X += moveSpeed
		}
	}

	return nil
}
