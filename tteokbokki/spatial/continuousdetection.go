package spatial

import (
	blueprintspatial "github.com/TheBitDrifter/blueprint/spatial"
	"github.com/TheBitDrifter/blueprint/vector"
)

// ContinuousCollisionDetector is the global continuous collision detector instance
var ContinuousCollisionDetector continuousDetector

// ContinuousDetector checks for collisions over a movement path
type continuousDetector struct{}

// Check tests for collisions between two shapes along their movement paths
// It returns whether a collision occurred, collision details, interpolated positions, and the time of impact
//
// Note: This implementation uses simple step-based interpolation which can be
// computationally expensive with high step counts. Use sparingly.
func (continuousDetector) Check(
	shapeA, shapeB blueprintspatial.Shape,
	posA, posB vector.TwoReader,
	prevPosA, prevPosB vector.TwoReader,
	steps int,
) (colliding bool, collision Collision, interPosA, interPosB vector.TwoReader, timeOfImpactRatio float64) {
	// Convert position readers to concrete vector.Two
	posAConc := vector.Two{X: posA.GetX(), Y: posA.GetY()}
	posBConc := vector.Two{X: posB.GetX(), Y: posB.GetY()}
	prevPosAConc := vector.Two{X: prevPosA.GetX(), Y: prevPosA.GetY()}
	prevPosBConc := vector.Two{X: prevPosB.GetX(), Y: prevPosB.GetY()}

	// Calculate movement vectors
	deltaA := posAConc.Sub(prevPosAConc)
	deltaB := posBConc.Sub(prevPosBConc)

	// Step through the movement path checking for collisions
	for step := 0; step <= steps; step++ {
		// Calculate interpolation factor
		t := float64(step) / float64(steps)
		// Interpolate positions
		interpPosA := prevPosAConc.Add(deltaA.Scale(t))
		interpPosB := prevPosBConc.Add(deltaB.Scale(t))

		newWorldVerticesA := UpdateWorldVerticesSimple(shapeA.Polygon.LocalVertices, interpPosA)
		shapeA.Polygon.WorldVertices = newWorldVerticesA

		newWorldVerticesB := UpdateWorldVerticesSimple(shapeB.Polygon.LocalVertices, interpPosB)
		shapeB.Polygon.WorldVertices = newWorldVerticesB

		// Check for collision at this step
		collided, collision := Detector.Check(shapeA, shapeB, interpPosA, interpPosB)
		if collided {
			return true, collision, interpPosA, interpPosB, t
		}
	}
	return false, Collision{}, posA, posB, 0.0
}
