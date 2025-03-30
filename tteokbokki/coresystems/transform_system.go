package coresystems

import (
	"github.com/TheBitDrifter/bappa/blueprint"
	"github.com/TheBitDrifter/bappa/blueprint/vector"
	"github.com/TheBitDrifter/bappa/tteokbokki/spatial"
)

// TransformSystem updates world coordinates for shapes based on position, rotation, and scale
type TransformSystem struct{}

// Run processes all entities with Shape components and updates their world vertices
func (TransformSystem) Run(scene blueprint.Scene, dt float64) error {
	cursor := scene.NewCursor(blueprint.Queries.Shape)
	for range cursor.Next() {
		shape := spatial.Components.Shape.GetFromCursor(cursor)
		hasPos, pos := spatial.Components.Position.GetFromCursorSafe(cursor)
		hasRot, rot := spatial.Components.Rotation.GetFromCursorSafe(cursor)
		hasScale, scale := spatial.Components.Scale.GetFromCursorSafe(cursor)

		// Initialize default transform values
		var posToUse, scaleToUse vector.Two
		scaleToUse.X = 1
		scaleToUse.Y = 1

		if hasPos {
			posToUse = pos.Two
		}

		if hasScale {
			scaleToUse = scale.Two
			if scaleToUse.X == 0 {
				scale.X = 1
			}
			if scaleToUse.Y == 0 {
				scaleToUse.Y = 1
			}
		}

		rot64 := 0.0
		if hasRot {
			rot64 = rot.AsFloat64()
		}

		newWorldVertices := spatial.UpdateWorldVertices(shape.Polygon.LocalVertices, posToUse, scaleToUse, rot64)
		shape.Polygon.WorldVertices = newWorldVertices
		spatial.UpdateSkinAndAAB(shape, scaleToUse, rot64)
	}
	return nil
}
