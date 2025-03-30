package spatial

import (
	"github.com/TheBitDrifter/bappa/blueprint/vector"
)

// UpdateWorldVerticesSimple transforms local vertices to world space by applying translation only
func UpdateWorldVerticesSimple(localVerts []vector.Two, pos vector.Two) []vector.Two {
	updated := make([]vector.Two, len(localVerts))
	for i := 0; i < len(localVerts); i++ {
		updated[i].X = localVerts[i].X + pos.X
		updated[i].Y = localVerts[i].Y + pos.Y
	}
	return updated
}

// UpdateWorldVertices transforms local vertices to world space by applying scale, rotation, and translation
func UpdateWorldVertices(localVerts []vector.Two, pos, scale vector.TwoReader, rot float64) []vector.Two {
	updated := make([]vector.Two, len(localVerts))
	for i := 0; i < len(localVerts); i++ {
		// Scale
		scaled := localVerts[i].CloneAsInterface()
		scaled.SetX(scaled.GetX() * scale.GetX())
		scaled.SetY(scaled.GetY() * scale.GetY())
		// Rotate
		rotated := scaled.RotateAsInterface(rot)
		// Translate (add position)
		translated := rotated.CloneAsInterface()
		translated.SetX(translated.GetX() + pos.GetX())
		translated.SetY(translated.GetY() + pos.GetY())
		// Store result
		updated[i].X = translated.GetX()
		updated[i].Y = translated.GetY()
	}
	return updated
}

// UpdateSkinAndAAB updates a shape's skin and axis-aligned bounding box based on scale and rotation
// Uses optimized path for non-rotated shapes
func UpdateSkinAndAAB(shape *Shape, scale vector.TwoReader, rot float64) {
	if rot != 0 {
		if shape.LocalAAB.Width != 0 {
			shape.Skin = CalcSkin(shape.Polygon, AAB{}, NewScale(scale.GetX(), scale.GetY()))
		}
		shape.LocalAAB = AAB{}
		shape.WorldAAB = AAB{}
		return
	}
	// Calculate new world AAB dimensions
	newWidth := shape.LocalAAB.Width * scale.GetX()
	newHeight := shape.LocalAAB.Height * scale.GetY()
	// Check if dimensions have changed to avoid unnecessary skin recalculation
	if shape.WorldAAB.Width != newWidth || shape.WorldAAB.Height != newHeight {
		shape.WorldAAB.Width = newWidth
		shape.WorldAAB.Height = newHeight
		// Only update skin if AAB dimensions changed
		shape.Skin = CalcSkin(shape.Polygon, shape.LocalAAB, scale)
	}
}
