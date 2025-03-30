package main

import (
	"embed"
	"log"

	"github.com/TheBitDrifter/blueprint"
	blueprintmotion "github.com/TheBitDrifter/blueprint/motion"
	blueprintspatial "github.com/TheBitDrifter/blueprint/spatial"
	"github.com/TheBitDrifter/coldbrew"
	coldbrew_rendersystems "github.com/TheBitDrifter/coldbrew/rendersystems"
	"github.com/TheBitDrifter/warehouse"

	tteo_coresystems "github.com/TheBitDrifter/tteokbokki/coresystems"
	"github.com/TheBitDrifter/tteokbokki/motion"
	"github.com/TheBitDrifter/tteokbokki/spatial"
)

var assets embed.FS

var floorTag = warehouse.FactoryNewComponent[struct{}]()

func main() {
	client := coldbrew.NewClient(
		640,
		360,
		10,
		10,
		10,
		assets,
	)

	client.SetTitle("Simple Collision and Physics")

	err := client.RegisterScene(
		"Example Scene",
		640,
		360,
		exampleScenePlan,
		[]coldbrew.RenderSystem{},
		[]coldbrew.ClientSystem{},
		[]blueprint.CoreSystem{
			gravitySystem{},
			tteo_coresystems.IntegrationSystem{},
			tteo_coresystems.TransformSystem{},
			collisionBounceSystem{},
		},
	)
	if err != nil {
		log.Fatal(err)
	}

	client.RegisterGlobalRenderSystem(
		coldbrew_rendersystems.GlobalRenderer{},
		&coldbrew_rendersystems.DebugRenderer{},
	)

	client.ActivateCamera()

	if err := client.Start(); err != nil {
		log.Fatal(err)
	}
}

func exampleScenePlan(height, width int, sto warehouse.Storage) error {
	boxArchetype, err := sto.NewOrExistingArchetype(
		blueprintspatial.Components.Position,
		blueprintspatial.Components.Rotation,
		blueprintspatial.Components.Shape,
		blueprintmotion.Components.Dynamics,
	)
	if err != nil {
		return err
	}
	for i := 0; i < 10; i++ {
		err = boxArchetype.Generate(1,
			blueprintspatial.NewPosition(float64(i*100), float64(20*i)),
			blueprintmotion.NewDynamics(10),
			blueprintspatial.NewRectangle(30, 40),
		)
		if err != nil {
			return err
		}
	}

	floorArchetype, err := sto.NewOrExistingArchetype(
		floorTag,
		blueprintspatial.Components.Position,
		blueprintspatial.Components.Rotation,
		blueprintspatial.Components.Shape,
		blueprintmotion.Components.Dynamics,
	)
	if err != nil {
		return err
	}

	err = floorArchetype.Generate(1,
		blueprintspatial.NewPosition(320, 300),
		blueprintmotion.NewDynamics(0),
		blueprintspatial.NewRectangle(800, 40),
	)

	return nil
}

type gravitySystem struct{}

func (gravitySystem) Run(scene blueprint.Scene, _ float64) error {
	const (
		DEFAULT_GRAVITY  = 9.8
		PIXELS_PER_METER = 50.0
	)

	cursor := scene.NewCursor(blueprint.Queries.Dynamics)
	for range cursor.Next() {
		dyn := blueprintmotion.Components.Dynamics.GetFromCursor(cursor)
		mass := 1 / dyn.InverseMass
		gravity := motion.Forces.Generator.NewGravityForce(mass, DEFAULT_GRAVITY, PIXELS_PER_METER)
		motion.Forces.AddForce(dyn, gravity)
	}
	return nil
}

type collisionBounceSystem struct{}

func (collisionBounceSystem) Run(scene blueprint.Scene, _ float64) error {
	boxQuery := warehouse.Factory.NewQuery().And(
		blueprintspatial.Components.Shape,
		warehouse.Factory.NewQuery().Not(floorTag),
	)
	floorQuery := warehouse.Factory.NewQuery().And(
		blueprintspatial.Components.Shape,
		floorTag,
	)

	boxCursor := scene.NewCursor(boxQuery)
	floorCursor := scene.NewCursor(floorQuery)

	for range boxCursor.Next() {
		for range floorCursor.Next() {
			boxPos := blueprintspatial.Components.Position.GetFromCursor(boxCursor)
			boxShape := blueprintspatial.Components.Shape.GetFromCursor(boxCursor)
			boxDyn := blueprintmotion.Components.Dynamics.GetFromCursor(boxCursor)

			// Get the block pos, shape, and dynamics
			floorPos := blueprintspatial.Components.Position.GetFromCursor(floorCursor)
			floorShape := blueprintspatial.Components.Shape.GetFromCursor(floorCursor)
			floorDyn := blueprintmotion.Components.Dynamics.GetFromCursor(floorCursor)

			// Check for a collision
			if ok, collisionResult := spatial.Detector.Check(
				*boxShape, *floorShape, boxPos.Two, floorPos.Two,
			); ok {
				motion.Resolver.Resolve(
					&boxPos.Two,
					&floorPos.Two,
					boxDyn,
					floorDyn,
					collisionResult,
				)

				boxDyn.Vel.Y -= 500
			}
		}
	}
	return nil
}
