package main

import (
	"embed"
	"log"
	"math"
	"math/rand"

	"github.com/TheBitDrifter/blueprint"
	blueprintmotion "github.com/TheBitDrifter/blueprint/motion"
	blueprintspatial "github.com/TheBitDrifter/blueprint/spatial"
	"github.com/TheBitDrifter/blueprint/vector"
	"github.com/TheBitDrifter/coldbrew"
	coldbrew_rendersystems "github.com/TheBitDrifter/coldbrew/rendersystems"
	"github.com/TheBitDrifter/warehouse"

	tteo_coresystems "github.com/TheBitDrifter/tteokbokki/coresystems"
	"github.com/TheBitDrifter/tteokbokki/motion"
	"github.com/TheBitDrifter/tteokbokki/spatial"
)

// Embedded assets
var assets embed.FS

// Component tags for entity identification
var (
	boundaryTag  = warehouse.FactoryNewComponent[struct{}]()
	circleTag    = warehouse.FactoryNewComponent[struct{}]()
	rectangleTag = warehouse.FactoryNewComponent[struct{}]()
	triangleTag  = warehouse.FactoryNewComponent[struct{}]()
	trapezoidTag = warehouse.FactoryNewComponent[struct{}]()
	hexagonTag   = warehouse.FactoryNewComponent[struct{}]()
	rhombusTag   = warehouse.FactoryNewComponent[struct{}]()
)

func main() {
	// Initialize the client with dimensions and assets
	client := coldbrew.NewClient(
		800, 600, 10, 10, 10, assets,
	)

	client.SetTitle("Physics Demo - Angular")

	// Register the main scene with systems
	err := client.RegisterScene(
		"Physics Demo",
		800, 600,
		playgroundScenePlan,
		[]coldbrew.RenderSystem{},
		[]coldbrew.ClientSystem{},
		[]blueprint.CoreSystem{
			newMovementSystem(),                  // Gravity and movement forces
			tteo_coresystems.IntegrationSystem{}, // Physics integration
			tteo_coresystems.TransformSystem{},   // Transform updates
			boundaryCollisionSystem{},            // Wall collisions
			shapeCollisionSystem{},               // Shape-to-shape collisions
		},
	)
	if err != nil {
		log.Fatal(err)
	}

	// Register global rendering systems
	client.RegisterGlobalRenderSystem(
		coldbrew_rendersystems.GlobalRenderer{},
		&coldbrew_rendersystems.DebugRenderer{},
	)

	client.ActivateCamera()

	// Start the application
	if err := client.Start(); err != nil {
		log.Fatal(err)
	}
}

// Creates the initial scene layout
func playgroundScenePlan(height, width int, sto warehouse.Storage) error {
	if err := createBoundaries(width, height, sto); err != nil {
		return err
	}

	if err := createShapes(width, height, sto); err != nil {
		return err
	}

	return nil
}

// Creates boundary walls around the scene
func createBoundaries(width, height int, sto warehouse.Storage) error {
	boundaryArchetype, err := sto.NewOrExistingArchetype(
		boundaryTag,
		blueprintspatial.Components.Position,
		blueprintspatial.Components.Shape,
		blueprintmotion.Components.Dynamics,
	)
	if err != nil {
		return err
	}

	wallThickness := 60.0
	gap := 1.0

	// Bottom wall (floor)
	floorDyn := blueprintmotion.NewDynamics(0)
	floorWidth := float64(width) - (2 * wallThickness) - (2 * gap)
	floorShape := blueprintspatial.NewRectangle(floorWidth, wallThickness)
	err = boundaryArchetype.Generate(1,
		blueprintspatial.NewPosition(float64(width)/2, float64(height)-(wallThickness/2)),
		floorShape,
		floorDyn,
	)
	if err != nil {
		return err
	}

	// Top wall (ceiling)
	ceilingDyn := blueprintmotion.NewDynamics(0)
	ceilingWidth := float64(width) - (2 * wallThickness) - (2 * gap)
	ceilingShape := blueprintspatial.NewRectangle(ceilingWidth, wallThickness)
	err = boundaryArchetype.Generate(1,
		blueprintspatial.NewPosition(float64(width)/2, wallThickness/2),
		ceilingShape,
		ceilingDyn,
	)
	if err != nil {
		return err
	}

	// Left wall
	leftWallDyn := blueprintmotion.NewDynamics(0)
	wallHeight := float64(height*2) - (2 * wallThickness) - (2 * gap)
	leftWallShape := blueprintspatial.NewRectangle(wallThickness, wallHeight)
	err = boundaryArchetype.Generate(1,
		blueprintspatial.NewPosition(wallThickness/2, float64(height)/2),
		leftWallShape,
		leftWallDyn,
	)
	if err != nil {
		return err
	}

	// Right wall
	rightWallDyn := blueprintmotion.NewDynamics(0)
	rightWallShape := blueprintspatial.NewRectangle(wallThickness, wallHeight)
	rightWallDyn.SetDefaultAngularMass(rightWallShape)
	err = boundaryArchetype.Generate(1,
		blueprintspatial.NewPosition(float64(width)-(wallThickness/2), float64(height)/2),
		rightWallShape,
		rightWallDyn,
	)
	if err != nil {
		return err
	}

	return nil
}

// Creates physics objects of different shapes
func createShapes(width, height int, sto warehouse.Storage) error {
	shapeCount := 0
	maxShapes := 7

	// Calculate even spacing for shapes
	shapeSpacing := float64(width-100) / float64(maxShapes+1)

	// Create rectangle shape
	if shapeCount < maxShapes {
		rectArchetype, err := sto.NewOrExistingArchetype(
			rectangleTag,
			blueprintspatial.Components.Position,
			blueprintspatial.Components.Rotation,
			blueprintspatial.Components.Shape,
			blueprintmotion.Components.Dynamics,
		)
		if err != nil {
			return err
		}

		rectWidth := 30.0 + rand.Float64()*40
		rectHeight := 30.0 + rand.Float64()*40
		rect := blueprintspatial.NewRectangle(rectWidth, rectHeight)
		dyn := blueprintmotion.NewDynamics(8 + rand.Float64()*3)
		dyn.SetDefaultAngularMass(rect)
		dyn.Elasticity = 0.2 + rand.Float64()*0.1
		dyn.Friction = 0.2 + rand.Float64()*0.1

		dyn.Vel.X = -10 + rand.Float64()*20
		dyn.Vel.Y = -5 + rand.Float64()*10

		xPos := 70 + shapeSpacing*float64(shapeCount)
		yPos := 150 + rand.Float64()*50

		err = rectArchetype.Generate(1,
			blueprintspatial.NewPosition(xPos, yPos),
			rect,
			dyn,
		)
		if err != nil {
			return err
		}
		shapeCount++
	}

	// Create triangle shape
	if shapeCount < maxShapes {
		triangleArchetype, err := sto.NewOrExistingArchetype(
			triangleTag,
			blueprintspatial.Components.Position,
			blueprintspatial.Components.Rotation,
			blueprintspatial.Components.Shape,
			blueprintmotion.Components.Dynamics,
		)
		if err != nil {
			return err
		}

		triWidth := 40.0 + rand.Float64()*40
		triHeight := 40.0 + rand.Float64()*40
		tri := blueprintspatial.NewTriangularPlatform(triWidth, triHeight)
		dyn := blueprintmotion.NewDynamics(6 + rand.Float64()*2)
		dyn.SetDefaultAngularMass(tri)
		dyn.Elasticity = 0.2 + rand.Float64()*0.1
		dyn.Friction = 0.2 + rand.Float64()*0.1

		dyn.Vel.X = -10 + rand.Float64()*20
		dyn.Vel.Y = -5 + rand.Float64()*10

		xPos := 70 + shapeSpacing*float64(shapeCount)
		yPos := 220 + rand.Float64()*50

		err = triangleArchetype.Generate(1,
			blueprintspatial.NewPosition(xPos, yPos),
			tri,
			dyn,
		)
		if err != nil {
			return err
		}
		shapeCount++
	}

	// Create trapezoid shape
	if shapeCount < maxShapes {
		trapezoidArchetype, err := sto.NewOrExistingArchetype(
			trapezoidTag,
			blueprintspatial.Components.Position,
			blueprintspatial.Components.Rotation,
			blueprintspatial.Components.Shape,
			blueprintmotion.Components.Dynamics,
		)
		if err != nil {
			return err
		}

		trapWidth := 50.0 + rand.Float64()*40
		trapHeight := 30.0 + rand.Float64()*30
		slopeRatio := 0.4 + rand.Float64()*0.3
		trap := blueprintspatial.NewTrapezoidPlatform(trapWidth, trapHeight, slopeRatio)
		dyn := blueprintmotion.NewDynamics(9 + rand.Float64()*3)
		dyn.SetDefaultAngularMass(trap)
		dyn.Elasticity = 0.2 + rand.Float64()*0.1
		dyn.Friction = 0.2 + rand.Float64()*0.1

		dyn.Vel.X = -10 + rand.Float64()*20
		dyn.Vel.Y = -5 + rand.Float64()*10

		xPos := 70 + shapeSpacing*float64(shapeCount)
		yPos := 290 + rand.Float64()*50

		err = trapezoidArchetype.Generate(1,
			blueprintspatial.NewPosition(xPos, yPos),
			trap,
			dyn,
		)
		if err != nil {
			return err
		}
		shapeCount++
	}

	// Create ramp shape
	if shapeCount < maxShapes {
		rampArchetype, err := sto.NewOrExistingArchetype(
			triangleTag,
			blueprintspatial.Components.Position,
			blueprintspatial.Components.Rotation,
			blueprintspatial.Components.Shape,
			blueprintmotion.Components.Dynamics,
		)
		if err != nil {
			return err
		}

		rampWidth := 60.0 + rand.Float64()*30
		rampHeight := 40.0 + rand.Float64()*20
		leftToRight := rand.Float64() > 0.5
		ramp := blueprintspatial.NewSingleRamp(rampWidth, rampHeight, leftToRight)
		dyn := blueprintmotion.NewDynamics(7 + rand.Float64()*3)
		dyn.SetDefaultAngularMass(ramp)
		dyn.Elasticity = 0.2 + rand.Float64()*0.1
		dyn.Friction = 0.2 + rand.Float64()*0.1

		dyn.Vel.X = -10 + rand.Float64()*20
		dyn.Vel.Y = -5 + rand.Float64()*10

		xPos := 70 + shapeSpacing*float64(shapeCount)
		yPos := 150 + rand.Float64()*50

		err = rampArchetype.Generate(1,
			blueprintspatial.NewPosition(xPos, yPos),
			ramp,
			dyn,
		)
		if err != nil {
			return err
		}
		shapeCount++
	}

	// Create hexagon shape
	if shapeCount < maxShapes {
		hexagonArchetype, err := sto.NewOrExistingArchetype(
			hexagonTag,
			blueprintspatial.Components.Position,
			blueprintspatial.Components.Rotation,
			blueprintspatial.Components.Shape,
			blueprintmotion.Components.Dynamics,
		)
		if err != nil {
			return err
		}

		hexSize := 30.0 + rand.Float64()*15
		hexagon := newHexagon(hexSize)
		dyn := blueprintmotion.NewDynamics(9 + rand.Float64()*3)
		dyn.SetDefaultAngularMass(hexagon)
		dyn.Elasticity = 0.2 + rand.Float64()*0.1
		dyn.Friction = 0.2 + rand.Float64()*0.1

		dyn.Vel.X = -10 + rand.Float64()*20
		dyn.Vel.Y = -5 + rand.Float64()*10

		xPos := 70 + shapeSpacing*float64(shapeCount)
		yPos := 220 + rand.Float64()*50

		err = hexagonArchetype.Generate(1,
			blueprintspatial.NewPosition(xPos, yPos),
			hexagon,
			dyn,
		)
		if err != nil {
			return err
		}
		shapeCount++
	}

	// Create rhombus shape
	if shapeCount < maxShapes {
		rhombusArchetype, err := sto.NewOrExistingArchetype(
			rhombusTag,
			blueprintspatial.Components.Position,
			blueprintspatial.Components.Rotation,
			blueprintspatial.Components.Shape,
			blueprintmotion.Components.Dynamics,
		)
		if err != nil {
			return err
		}

		rhombusWidth := 60.0 + rand.Float64()*20
		rhombusHeight := 30.0 + rand.Float64()*20
		rhombus := newRhombus(rhombusWidth, rhombusHeight)
		dyn := blueprintmotion.NewDynamics(8 + rand.Float64()*2)
		dyn.SetDefaultAngularMass(rhombus)
		dyn.Elasticity = 0.2 + rand.Float64()*0.1
		dyn.Friction = 0.2 + rand.Float64()*0.1

		dyn.Vel.X = -10 + rand.Float64()*20
		dyn.Vel.Y = -5 + rand.Float64()*10

		xPos := 70 + shapeSpacing*float64(shapeCount)
		yPos := 290 + rand.Float64()*50

		err = rhombusArchetype.Generate(1,
			blueprintspatial.NewPosition(xPos, yPos),
			rhombus,
			dyn,
		)
		if err != nil {
			return err
		}
		shapeCount++
	}

	return nil
}

// Helper function to create a regular hexagon shape
func newHexagon(size float64) blueprintspatial.Shape {
	vertices := make([]vector.Two, 6)

	// Generate vertices in a circle
	for i := 0; i < 6; i++ {
		angle := float64(i) * (math.Pi / 3.0)
		vertices[i] = vector.Two{
			X: size * math.Cos(angle),
			Y: size * math.Sin(angle),
		}
	}

	return blueprintspatial.NewPolygon(vertices)
}

// Helper function to create a rhombus (diamond) shape
func newRhombus(width, height float64) blueprintspatial.Shape {
	vertices := make([]vector.Two, 4)
	halfWidth := width / 2
	halfHeight := height / 2

	// Create a diamond shape
	vertices[0] = vector.Two{X: 0, Y: -halfHeight} // Top point
	vertices[1] = vector.Two{X: halfWidth, Y: 0}   // Right point
	vertices[2] = vector.Two{X: 0, Y: halfHeight}  // Bottom point
	vertices[3] = vector.Two{X: -halfWidth, Y: 0}  // Left point

	return blueprintspatial.NewPolygon(vertices)
}

// System that applies gravity and horizontal movement forces
type movementSystem struct {
	tickCounter int
	gravityDir  float64 // Direction of gravity (1.0 = down, -1.0 = up)
	moveDir     float64 // Direction of horizontal force (1.0 = right, -1.0 = left)
	moveTicks   int     // Counter for horizontal movement timing
}

// Creates a new movement system with initial values
func newMovementSystem() *movementSystem {
	return &movementSystem{
		tickCounter: 0,
		gravityDir:  1.0,
		moveDir:     1.0,
		moveTicks:   0,
	}
}

// Applies forces to all dynamic objects each frame
func (m *movementSystem) Run(scene blueprint.Scene, _ float64) error {
	const (
		GRAVITY               = 3.5
		PIXELS_PER_METER      = 30.0
		GRAVITY_FLIP_INTERVAL = 400 // Ticks before gravity reverses
		MOVE_FORCE            = 2.0
		MOVE_FLIP_INTERVAL    = 350 // Ticks before horizontal force reverses
	)

	// Update timers
	m.tickCounter++
	m.moveTicks++

	// Check if it's time to flip gravity direction
	if m.tickCounter >= GRAVITY_FLIP_INTERVAL {
		m.gravityDir = -m.gravityDir
		m.tickCounter = 0
	}

	// Check if it's time to flip horizontal movement direction
	if m.moveTicks >= MOVE_FLIP_INTERVAL {
		m.moveDir = -m.moveDir
		m.moveTicks = 0
	}

	// Apply forces to all dynamic objects
	cursor := scene.NewCursor(blueprint.Queries.Dynamics)
	for range cursor.Next() {
		dyn := blueprintmotion.Components.Dynamics.GetFromCursor(cursor)

		// Skip static objects (walls)
		if dyn.InverseMass <= 0 {
			continue
		}

		mass := 1 / dyn.InverseMass

		// Apply gravity force
		gravity := motion.Forces.Generator.NewGravityForce(
			mass,
			GRAVITY*m.gravityDir,
			PIXELS_PER_METER,
		)
		motion.Forces.AddForce(dyn, gravity)

		// Apply horizontal force
		horizontalForce := vector.Two{
			X: mass * MOVE_FORCE * PIXELS_PER_METER * m.moveDir,
			Y: 0,
		}
		motion.Forces.AddForce(dyn, horizontalForce)
	}

	return nil
}

// System that handles collisions between shapes and boundaries
type boundaryCollisionSystem struct{}

// Detects and resolves collisions between movable shapes and walls
func (boundaryCollisionSystem) Run(scene blueprint.Scene, _ float64) error {
	// Query for movable shapes (non-boundary objects)
	movableQuery := warehouse.Factory.NewQuery().And(
		blueprintspatial.Components.Shape,
		blueprintmotion.Components.Dynamics,
		warehouse.Factory.NewQuery().Not(boundaryTag),
	)

	// Query for boundary objects (walls)
	boundaryQuery := warehouse.Factory.NewQuery().And(
		blueprintspatial.Components.Shape,
		boundaryTag,
	)

	movableCursor := scene.NewCursor(movableQuery)
	boundaryCursor := scene.NewCursor(boundaryQuery)

	// Check each movable shape against each boundary
	for range movableCursor.Next() {
		shapePos := blueprintspatial.Components.Position.GetFromCursor(movableCursor)
		shapeShape := blueprintspatial.Components.Shape.GetFromCursor(movableCursor)
		shapeDyn := blueprintmotion.Components.Dynamics.GetFromCursor(movableCursor)

		for range boundaryCursor.Next() {
			boundaryPos := blueprintspatial.Components.Position.GetFromCursor(boundaryCursor)
			boundaryShape := blueprintspatial.Components.Shape.GetFromCursor(boundaryCursor)
			boundaryDyn := blueprintmotion.Components.Dynamics.GetFromCursor(boundaryCursor)

			// Check for collision and resolve it if detected
			if ok, collisionResult := spatial.Detector.Check(
				*shapeShape, *boundaryShape, shapePos.Two, boundaryPos.Two,
			); ok {
				motion.Resolver.Resolve(
					&shapePos.Two,
					&boundaryPos.Two,
					shapeDyn,
					boundaryDyn,
					collisionResult,
				)
			}
		}
		boundaryCursor.Reset()
	}
	return nil
}

// System that handles collisions between movable shapes
type shapeCollisionSystem struct{}

// Entity data for collision processing
type ShapeEntity struct {
	Position *blueprintspatial.Position
	Shape    *blueprintspatial.Shape
	Dynamics *blueprintmotion.Dynamics
	ID       int
}

// Detects and resolves collisions between movable shapes
func (shapeCollisionSystem) Run(scene blueprint.Scene, _ float64) error {
	// Query for all movable shapes
	movableQuery := warehouse.Factory.NewQuery().And(
		blueprintspatial.Components.Position,
		blueprintspatial.Components.Shape,
		blueprintmotion.Components.Dynamics,
		warehouse.Factory.NewQuery().Not(boundaryTag),
	)

	// Gather all shape entities
	var shapes []ShapeEntity
	movableCursor := scene.NewCursor(movableQuery)
	entityID := 0

	for range movableCursor.Next() {
		shapes = append(shapes, ShapeEntity{
			Position: blueprintspatial.Components.Position.GetFromCursor(movableCursor),
			Shape:    blueprintspatial.Components.Shape.GetFromCursor(movableCursor),
			Dynamics: blueprintmotion.Components.Dynamics.GetFromCursor(movableCursor),
			ID:       entityID,
		})
		entityID++
	}

	// Check each shape pair for collisions
	for i := 0; i < len(shapes); i++ {
		for j := i + 1; j < len(shapes); j++ {
			entity1 := shapes[i]
			entity2 := shapes[j]

			// Check for collision and resolve if detected
			if ok, collisionResult := spatial.Detector.Check(
				*entity1.Shape, *entity2.Shape, entity1.Position.Two, entity2.Position.Two,
			); ok {
				motion.Resolver.Resolve(
					&entity1.Position.Two,
					&entity2.Position.Two,
					entity1.Dynamics,
					entity2.Dynamics,
					collisionResult,
				)

				// Occasionally apply small random impulses to keep simulation interesting
				if rand.Float64() < 0.005 {
					randomImpulse := vector.Two{
						X: -5 + rand.Float64()*10,
						Y: -5 + rand.Float64()*10,
					}

					// Randomly select which object gets the impulse
					if rand.Float64() < 0.5 {
						motion.ApplyImpulse(entity1.Dynamics, randomImpulse, vector.Two{})
					} else {
						motion.ApplyImpulse(entity2.Dynamics, randomImpulse, vector.Two{})
					}
				}
			}
		}
	}
	return nil
}
