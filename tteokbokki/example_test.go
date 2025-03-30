package tteokbokki_test

import (
	"fmt"

	"github.com/TheBitDrifter/bappa/blueprint/vector"
	"github.com/TheBitDrifter/bappa/tteokbokki/motion"
	"github.com/TheBitDrifter/bappa/tteokbokki/spatial"
)

// Example_basicCollision demonstrates detecting and resolving a basic collision between two objects
func Example_basicCollision() {
	// Create box shapes
	playerShape := spatial.NewRectangle(40, 80)  // Player box (40x80)
	groundShape := spatial.NewRectangle(200, 20) // Ground platform (200x20)

	// Create positions - player slightly overlapping with ground
	playerPos := spatial.NewPosition(100, 110)
	groundPos := spatial.NewPosition(100, 150)

	// Update world vertices for collision detection
	playerShape.Polygon.WorldVertices = spatial.UpdateWorldVerticesSimple(
		playerShape.Polygon.LocalVertices, playerPos.Two)
	groundShape.Polygon.WorldVertices = spatial.UpdateWorldVerticesSimple(
		groundShape.Polygon.LocalVertices, groundPos.Two)

	// Setup dynamics
	playerDyn := motion.NewDynamics(1.0)
	playerDyn.Vel = vector.Two{X: 0, Y: 40.0} // Moving down
	playerDyn.Elasticity = 0.3
	playerDyn.SetDefaultAngularMass(playerShape)

	// Ground is static (infinite mass)
	groundDyn := motion.NewDynamics(0.0) // Zero mass = static
	groundDyn.Elasticity = 0.5

	fmt.Printf("Initial: Player at (%.1f, %.1f) with velocity (%.1f, %.1f)\n",
		playerPos.X, playerPos.Y, playerDyn.Vel.X, playerDyn.Vel.Y)

	// Detect collision
	colliding, collision := spatial.Detector.Check(
		playerShape, groundShape, playerPos.Two, groundPos.Two)

	if colliding {
		fmt.Printf("Collision detected!\n")
		fmt.Printf("  Normal: (%.1f, %.1f)\n", collision.Normal.X, collision.Normal.Y)
		fmt.Printf("  Depth: %.1f\n", collision.Depth)

		// Resolve the collision
		motion.Resolver.Resolve(
			&playerPos.Two, &groundPos.Two,
			&playerDyn, &groundDyn,
			collision)

		fmt.Printf("After resolution: Player at (%.1f, %.1f) with velocity (%.1f, %.1f)\n",
			playerPos.X, playerPos.Y, playerDyn.Vel.X, playerDyn.Vel.Y)
	}

	// Output:
	// Initial: Player at (100.0, 110.0) with velocity (0.0, 40.0)
	// Collision detected!
	//   Normal: (0.0, 1.0)
	//   Depth: 10.0
	// After resolution: Player at (100.0, 100.0) with velocity (0.0, -16.0)
}
