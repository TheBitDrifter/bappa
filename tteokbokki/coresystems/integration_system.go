package coresystems

import (
	"github.com/TheBitDrifter/blueprint"
	blueprintmotion "github.com/TheBitDrifter/blueprint/motion"
	blueprintspatial "github.com/TheBitDrifter/blueprint/spatial"
	"github.com/TheBitDrifter/tteokbokki/motion"
	"github.com/TheBitDrifter/warehouse"
)

// IntegrationSystem handles position and rotation integration based on dynamics
type IntegrationSystem struct{}

// Run performs integration of positions and rotations based on dynamic properties
func (IntegrationSystem) Run(scene blueprint.Scene, dt float64) error {
	// Query for entities with position, rotation, and dynamics components
	withRotation := warehouse.Factory.NewQuery().And(
		blueprintspatial.Components.Position,
		blueprintspatial.Components.Rotation,
		blueprintmotion.Components.Dynamics,
	)

	// Query for entities with position and dynamics but without rotation
	onlyLinear := warehouse.Factory.NewQuery().And(
		blueprintspatial.Components.Position,
		blueprintmotion.Components.Dynamics,
		warehouse.Factory.NewQuery().Not(blueprintspatial.Components.Rotation),
	)

	// Helper function to integrate positions and rotations
	integrate := func(query warehouse.QueryNode, hasRot bool) {
		cursor := scene.NewCursor(query)
		for range cursor.Next() {
			dyn := blueprintmotion.Components.Dynamics.GetFromCursor(cursor)
			position := blueprintspatial.Components.Position.GetFromCursor(cursor)

			rotV := blueprintspatial.Rotation(0)
			rotation := &rotV
			if hasRot {
				rotation = blueprintspatial.Components.Rotation.GetFromCursor(cursor)
			}

			// Compute new position and rotation values
			newPos, newRot := motion.Integrate(dyn, position, float64(*rotation), dt)

			// Store current position as previous position if component exists

			// Update position and rotation with new values
			position.X = newPos.X
			position.Y = newPos.Y
			*rotation = blueprintspatial.Rotation(newRot)
		}
	}

	// Process entities with rotation
	integrate(withRotation, true)

	// Process entities without rotation
	integrate(onlyLinear, false)

	return nil
}
