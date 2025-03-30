package tteo_coresystems

import (
	"github.com/TheBitDrifter/bappa/blueprint"
	"github.com/TheBitDrifter/bappa/tteokbokki/motion"
	"github.com/TheBitDrifter/bappa/tteokbokki/spatial"
	"github.com/TheBitDrifter/bappa/warehouse"
)

// IntegrationSystem handles position and rotation integration based on dynamics
type IntegrationSystem struct{}

// Run performs integration of positions and rotations based on dynamic properties
func (IntegrationSystem) Run(scene blueprint.Scene, dt float64) error {
	// Query for entities with position, rotation, and dynamics components
	withRotation := warehouse.Factory.NewQuery().And(
		spatial.Components.Position,
		spatial.Components.Rotation,
		motion.Components.Dynamics,
	)

	// Query for entities with position and dynamics but without rotation
	onlyLinear := warehouse.Factory.NewQuery().And(
		spatial.Components.Position,
		motion.Components.Dynamics,
		warehouse.Factory.NewQuery().Not(spatial.Components.Rotation),
	)

	// Helper function to integrate positions and rotations
	integrate := func(query warehouse.QueryNode, hasRot bool) {
		cursor := scene.NewCursor(query)
		for range cursor.Next() {
			dyn := motion.Components.Dynamics.GetFromCursor(cursor)
			position := spatial.Components.Position.GetFromCursor(cursor)

			rotV := spatial.Rotation(0)
			rotation := &rotV
			if hasRot {
				rotation = spatial.Components.Rotation.GetFromCursor(cursor)
			}

			// Compute new position and rotation values
			newPos, newRot := motion.Integrate(dyn, position, float64(*rotation), dt)

			// Store current position as previous position if component exists

			// Update position and rotation with new values
			position.X = newPos.X
			position.Y = newPos.Y
			*rotation = spatial.Rotation(newRot)
		}
	}

	// Process entities with rotation
	integrate(withRotation, true)

	// Process entities without rotation
	integrate(onlyLinear, false)

	return nil
}
