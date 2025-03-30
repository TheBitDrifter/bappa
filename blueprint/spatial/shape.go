package blueprintspatial

import (
	"slices"

	vector "github.com/TheBitDrifter/blueprint/vector"
)

// Shape represents a geometric shape with local and world bounds, polygon definition, and skin
type Shape struct {
	LocalAAB AAB
	WorldAAB AAB
	Polygon  Polygon
	Skin     Skin
}

// Circle represents a circular shape with radius
type Circle struct {
	Radius float64
}

// Polygon represents a shape defined by vertices in both local and world space
type Polygon struct {
	LocalVertices []vector.Two
	WorldVertices []vector.Two
}

// Skin contains collision detection primitives for a shape
type Skin struct {
	AAB    AAB
	Circle Circle
}

// AAB represents an axis-aligned bounding box
type AAB struct {
	Width  float64
	Height float64
}

// NewPolygon creates a new shape from a set of local vertices
func NewPolygon(localVertices []vector.Two) Shape {
	// Create deep copy of vertices for both local and world vertices
	localCopy := make([]vector.Two, len(localVertices))
	worldCopy := make([]vector.Two, len(localVertices))

	for i, vertex := range localVertices {
		localCopy[i] = vertex.Clone()
		worldCopy[i] = vertex.Clone()
	}

	shape := Shape{
		Polygon: Polygon{
			LocalVertices: localCopy,
			WorldVertices: worldCopy,
		},
	}
	shape.Skin = CalcSkin(shape.Polygon, AAB{}, NewScale(1, 1))
	return shape
}

// NewRectangle creates a rectangular shape with specified width and height
func NewRectangle(width, height float64) Shape {
	shape := NewPolygon(make([]vector.Two, 4))
	// Clockwise order
	shape.Polygon.LocalVertices[0] = vector.Two{-width / 2, -height / 2} // top-left
	shape.Polygon.LocalVertices[1] = vector.Two{width / 2, -height / 2}  // top-right
	shape.Polygon.LocalVertices[2] = vector.Two{width / 2, height / 2}   // bottom-right
	shape.Polygon.LocalVertices[3] = vector.Two{-width / 2, height / 2}  // bottom-left
	shape.LocalAAB = AAB{
		Width:  width,
		Height: height,
	}
	shape.WorldAAB = shape.LocalAAB
	shape.Polygon.WorldVertices = slices.Clone(shape.Polygon.LocalVertices)
	shape.Skin = CalcSkin(shape.Polygon, shape.LocalAAB, NewScale(1, 1))
	return shape
}

// CalcSkin calculates collision primitives for a polygon
func CalcSkin(p Polygon, aab AAB, scale vector.TwoReader) Skin {
	var skin Skin

	// Handle empty polygon
	if len(p.LocalVertices) == 0 {
		return skin
	}

	// Calculate the circle skin (longest distance from center, accounting for scale)
	origin := vector.Two{X: 0, Y: 0}
	longestDistanceFromCenter := 0.0

	for _, vert := range p.LocalVertices {
		// Create a scaled copy of the vertex
		scaledVert := vert
		scaledVert.X *= scale.GetX()
		scaledVert.Y *= scale.GetY()

		distance := scaledVert.Sub(origin).Mag()
		if distance > longestDistanceFromCenter {
			longestDistanceFromCenter = distance
		}
	}

	// Always create the circle skin
	skin.Circle = Circle{
		Radius: longestDistanceFromCenter,
	}

	// Use the provided AAB if it has height, applying scale
	if aab.Height > 0 {
		skin.AAB = AAB{
			Width:  aab.Width * scale.GetX(),
			Height: aab.Height * scale.GetY(),
		}
	}

	return skin
}

// NewTriangularPlatform creates a triangular shape with a flat top of the specified width
// and angled sides meeting at a point at the specified height below the top
// This is useful for one-way platforms where you want to prevent side collisions
func NewTriangularPlatform(width, height float64) Shape {
	// Create a polygon with 3 vertices
	shape := NewPolygon(make([]vector.Two, 3))

	// Clockwise order starting from top-left
	shape.Polygon.LocalVertices[0] = vector.Two{-width / 2, -height / 2} // top-left
	shape.Polygon.LocalVertices[1] = vector.Two{width / 2, -height / 2}  // top-right
	shape.Polygon.LocalVertices[2] = vector.Two{0, height / 2}           // bottom-center (point)

	// Clone the local vertices to world vertices
	shape.Polygon.WorldVertices = slices.Clone(shape.Polygon.LocalVertices)

	// Calculate skin
	shape.Skin = CalcSkin(shape.Polygon, shape.LocalAAB, NewScale(1, 1))

	return shape
}

// NewTrapezoidPlatform creates a trapezoid shape with a flat top of the specified width,
// angled sides, and a narrower bottom. The slopeRatio controls how much narrower the bottom is
// This is useful for creating one-way platforms with more gradual angled sides
func NewTrapezoidPlatform(width, height float64, slopeRatio float64) Shape {
	// Create a polygon with 4 vertices
	shape := NewPolygon(make([]vector.Two, 4))

	// Calculate bottom width based on slope ratio (0.0 = triangle, 1.0 = rectangle)
	bottomWidth := width * slopeRatio

	// Clockwise order
	shape.Polygon.LocalVertices[0] = vector.Two{-width / 2, -height / 2}      // top-left
	shape.Polygon.LocalVertices[1] = vector.Two{width / 2, -height / 2}       // top-right
	shape.Polygon.LocalVertices[2] = vector.Two{bottomWidth / 2, height / 2}  // bottom-right
	shape.Polygon.LocalVertices[3] = vector.Two{-bottomWidth / 2, height / 2} // bottom-left

	// Clone the local vertices to world vertices
	shape.Polygon.WorldVertices = slices.Clone(shape.Polygon.LocalVertices)

	// Calculate skin
	shape.Skin = CalcSkin(shape.Polygon, shape.LocalAAB, NewScale(1, 1))

	return shape
}

// NewSingleRamp creates a triangular ramp shape with specified dimensions
// When leftToRight is true, the ramp ascends from left to right
func NewSingleRamp(width, height float64, leftToRight bool) Shape {
	// Create a triangle for the ramp
	ramp := NewPolygon(make([]vector.Two, 3))

	if leftToRight {
		// Peak on the right
		ramp.Polygon.LocalVertices[0] = vector.Two{-width / 2, height / 2} // Bottom-left
		ramp.Polygon.LocalVertices[1] = vector.Two{width / 2, -height / 2} // Top-right (peak)
		ramp.Polygon.LocalVertices[2] = vector.Two{width / 2, height / 2}  // Bottom-right
	} else {
		// Peak on the left
		ramp.Polygon.LocalVertices[0] = vector.Two{-width / 2, -height / 2} // Top-left (peak)
		ramp.Polygon.LocalVertices[1] = vector.Two{width / 2, height / 2}   // Bottom-right
		ramp.Polygon.LocalVertices[2] = vector.Two{-width / 2, height / 2}  // Bottom-left
	}

	ramp.Polygon.WorldVertices = slices.Clone(ramp.Polygon.LocalVertices)
	ramp.Skin = CalcSkin(ramp.Polygon, ramp.LocalAAB, NewScale(1, 1))
	return ramp
}

// NewDoubleRamp creates a two-sided ramp (trapezoid with slopes on both sides)
// width: total width of the ramp
// height: height of the ramp
// topWidthRatio: width of the flat section on top (as a ratio of total width, 0-1)
// Note: In this engine, lower Y values are higher in the world (y=0 is top of screen)
func NewDoubleRamp(width, height float64, topWidthRatio float64) Shape {
	// Create a trapezoid for the ramp
	ramp := NewPolygon(make([]vector.Two, 4))

	// Ensure topWidth is between 0 and 1
	if topWidthRatio < 0 {
		topWidthRatio = 0
	}
	if topWidthRatio > 1 {
		topWidthRatio = 1
	}

	// Calculate the horizontal offset for the top vertices
	topOffset := (width * (1 - topWidthRatio)) / 2

	// Define trapezoid vertices (clockwise order)
	// Remember: lower Y values = higher in world
	ramp.Polygon.LocalVertices[0] = vector.Two{-width/2 + topOffset, -height / 2} // Top-left
	ramp.Polygon.LocalVertices[1] = vector.Two{width/2 - topOffset, -height / 2}  // Top-right
	ramp.Polygon.LocalVertices[2] = vector.Two{width / 2, height / 2}             // Bottom-right
	ramp.Polygon.LocalVertices[3] = vector.Two{-width / 2, height / 2}            // Bottom-left

	ramp.Polygon.WorldVertices = slices.Clone(ramp.Polygon.LocalVertices)
	ramp.Skin = CalcSkin(ramp.Polygon, ramp.LocalAAB, NewScale(1, 1))
	return ramp
}
