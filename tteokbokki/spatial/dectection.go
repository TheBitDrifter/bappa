package spatial

import (
	"math"

	"github.com/TheBitDrifter/bappa/blueprint/vector"
)

// Detector is the global collision detector instance
var Detector detector

// detector implements collision detection between shapes
type detector struct{}

// Check determines if two shapes are colliding and returns collision details
func (detector) Check(shapeA, shapeB Shape, posA, posB vector.TwoReader) (bool, Collision) {
	posAConc := vector.Two{
		X: posA.GetX(),
		Y: posA.GetY(),
	}
	posBConc := vector.Two{
		X: posB.GetX(),
		Y: posB.GetY(),
	}
	isAABCollision := shapeA.WorldAAB.Height != 0 && shapeB.WorldAAB.Height != 0
	if isAABCollision {
		return inspectAABCollision(shapeA.WorldAAB, shapeB.WorldAAB, posAConc, posBConc)
	}

	check := broadFilter(shapeA, shapeB, posAConc, posBConc)
	if !check {
		return false, Collision{}
	}
	return inspectPolygonCollision(shapeA.Polygon, shapeB.Polygon)
}

// inspectCircleCollision checks for collision between two circles and returns collision data
func inspectCircleCollision(
	circleA, circleB Circle,
	posA, posB vector.Two,
) (bool, Collision) {
	distanceBetween := posB.Sub(posA)
	radiusSum := circleB.Radius + circleA.Radius
	notColliding := distanceBetween.MagSquared() > (radiusSum * radiusSum)
	if notColliding {
		return false, Collision{}
	}
	normal := distanceBetween.Norm()
	start := posB.Sub(normal.Scale(circleB.Radius))
	end := posA.Add(normal.Scale(circleA.Radius))
	depth := end.Sub(start).Mag()

	// For circles, we don't have traditional edges, so we use a placeholder
	edgeA := CollisionEdge{
		Index:    -1,
		Vertices: []vector.Two{},
	}

	edgeB := CollisionEdge{
		Index:    -1,
		Vertices: []vector.Two{},
	}

	return true, Collision{start, end, normal, depth, edgeA, edgeB}
}

// inspectAABCollision checks for collision between two axis-aligned bounding boxes
func inspectAABCollision(aabA, aabB AAB,
	posA, posB vector.Two,
) (bool, Collision) {
	halfWidthA := aabA.Width / 2
	halfHeightA := aabA.Height / 2
	halfWidthB := aabB.Width / 2
	halfHeightB := aabB.Height / 2
	xA := posA.X
	yA := posA.Y
	xB := posB.X
	yB := posB.Y
	leftA := xA - halfWidthA
	rightA := xA + halfWidthA
	topA := yA - halfHeightA
	bottomA := yA + halfHeightA
	leftB := xB - halfWidthB
	rightB := xB + halfWidthB
	topB := yB - halfHeightB
	bottomB := yB + halfHeightB

	// Early exit if no collision
	if rightA < leftB || leftA > rightB || bottomA < topB || topA > bottomB {
		return false, Collision{}
	}

	// Calculate overlap on each axis
	xOverlap := math.Min(rightA, rightB) - math.Max(leftA, leftB)
	yOverlap := math.Min(bottomA, bottomB) - math.Max(topA, topB)

	// Use the smallest overlap to determine collision normal and depth
	var normal vector.Two
	var depth float64
	var edgeIndexA, edgeIndexB int
	var collidingEdgeVerticesA, collidingEdgeVerticesB []vector.Two

	if xOverlap < yOverlap {
		depth = xOverlap
		if posA.X < posB.X {
			normal = vector.Two{X: 1, Y: 0}
			// Right edge of A, left edge of B
			edgeIndexA = 1 // Right edge
			collidingEdgeVerticesA = []vector.Two{
				{X: rightA, Y: topA},
				{X: rightA, Y: bottomA},
			}
			edgeIndexB = 3 // Left edge
			collidingEdgeVerticesB = []vector.Two{
				{X: leftB, Y: topB},
				{X: leftB, Y: bottomB},
			}
		} else {
			normal = vector.Two{X: -1, Y: 0}
			// Left edge of A, right edge of B
			edgeIndexA = 3 // Left edge
			collidingEdgeVerticesA = []vector.Two{
				{X: leftA, Y: topA},
				{X: leftA, Y: bottomA},
			}
			edgeIndexB = 1 // Right edge
			collidingEdgeVerticesB = []vector.Two{
				{X: rightB, Y: topB},
				{X: rightB, Y: bottomB},
			}
		}
	} else {
		depth = yOverlap
		if posA.Y < posB.Y {
			normal = vector.Two{X: 0, Y: 1}
			// Bottom edge of A, top edge of B
			edgeIndexA = 2 // Bottom edge
			collidingEdgeVerticesA = []vector.Two{
				{X: leftA, Y: bottomA},
				{X: rightA, Y: bottomA},
			}
			edgeIndexB = 0 // Top edge
			collidingEdgeVerticesB = []vector.Two{
				{X: leftB, Y: topB},
				{X: rightB, Y: topB},
			}
		} else {
			normal = vector.Two{X: 0, Y: -1}
			// Top edge of A, bottom edge of B
			edgeIndexA = 0 // Top edge
			collidingEdgeVerticesA = []vector.Two{
				{X: leftA, Y: topA},
				{X: rightA, Y: topA},
			}
			edgeIndexB = 2 // Bottom edge
			collidingEdgeVerticesB = []vector.Two{
				{X: leftB, Y: bottomB},
				{X: rightB, Y: bottomB},
			}
		}
	}

	// Create the collision edges with index information
	collidingEdgeA := CollisionEdge{
		Index:    edgeIndexA,
		Vertices: collidingEdgeVerticesA,
	}

	collidingEdgeB := CollisionEdge{
		Index:    edgeIndexB,
		Vertices: collidingEdgeVerticesB,
	}

	// Calculate contact points based on collision direction
	var start, end vector.Two
	if normal.X != 0 {
		// Horizontal collision
		y := (math.Max(topA, topB) + math.Min(bottomA, bottomB)) / 2
		if normal.X > 0 {
			start = vector.Two{X: rightA, Y: y}
			end = vector.Two{X: leftB, Y: y}
		} else {
			start = vector.Two{X: leftA, Y: y}
			end = vector.Two{X: rightB, Y: y}
		}
	} else {
		// Vertical collision
		x := (math.Max(leftA, leftB) + math.Min(rightA, rightB)) / 2
		if normal.Y > 0 {
			start = vector.Two{X: x, Y: bottomA}
			end = vector.Two{X: x, Y: topB}
		} else {
			start = vector.Two{X: x, Y: topA}
			end = vector.Two{X: x, Y: bottomB}
		}
	}

	return true, Collision{start, end, normal, depth, collidingEdgeA, collidingEdgeB}
}

// broadFilter performs initial collision detection using bounding volumes
func broadFilter(shapeA, shapeB Shape, posA, posB vector.Two) bool {
	isAABCheck := shapeA.Skin.AAB.Height != 0 && shapeB.Skin.AAB.Height != 0
	if isAABCheck {
		check, _ := inspectAABCollision(shapeA.Skin.AAB, shapeB.Skin.AAB, posA, posB)
		return check
	}
	check, _ := inspectCircleCollision(shapeA.Skin.Circle, shapeB.Skin.Circle, posA, posB)
	return check
}

// inspectPolygonCollision performs detailed collision detection between two polygons
func inspectPolygonCollision(polygonA, polygonB Polygon) (bool, Collision) {
	var collision Collision
	minSepA, incidentEdgeIndexA, penPointA := findMinSep(polygonA, polygonB)

	edgeVectorA, v1A, v2A := edge(incidentEdgeIndexA, polygonA)
	// Create colliding edge from polygon A with index information
	collidingEdgeA := CollisionEdge{
		Index:    incidentEdgeIndexA,
		Vertices: []vector.Two{v1A, v2A},
	}

	if minSepA >= 0 {
		return false, collision
	}

	minSepB, incidentEdgeIndexB, penPointB := findMinSep(polygonB, polygonA)

	edgeVectorB, v1B, v2B := edge(incidentEdgeIndexB, polygonB)
	// Create colliding edge from polygon B with index information
	collidingEdgeB := CollisionEdge{
		Index:    incidentEdgeIndexB,
		Vertices: []vector.Two{v1B, v2B},
	}

	if minSepB >= 0 {
		return false, collision
	}

	if minSepA > minSepB {
		depth := -minSepA
		normal := edgeVectorA.Perpendicular().Norm()
		start := penPointA
		end := start.Add(normal.Scale(depth))
		collision = Collision{start, end, normal, depth, collidingEdgeA, collidingEdgeB}
	} else {
		depth := -minSepB
		normal := edgeVectorB.Perpendicular().Norm().Scale(-1)
		start := penPointB.Sub(normal.Scale(depth))
		end := penPointB
		collision = Collision{start, end, normal, depth, collidingEdgeA, collidingEdgeB}
	}

	return true, collision
}

// findMinSep finds the minimum separation between two polygons
// Returns separation distance, reference edge index, and penetration point
func findMinSep(polygonA, polygonB Polygon) (float64, int, vector.Two) {
	sep := -math.MaxFloat64
	var indexReferenceEdge int
	var penPoint vector.Two

	for i := range polygonA.WorldVertices {
		va := polygonA.WorldVertices[i]
		currentEdge, _, _ := edge(i, polygonA)
		normal := currentEdge.Perpendicular().Norm()

		minSep := math.MaxFloat64
		var minVert vector.Two

		for _, vb := range polygonB.WorldVertices {
			projection := vb.Sub(va).ScalarProduct(normal)
			if projection < minSep {
				minSep = projection
				minVert = vb
			}
		}

		if minSep > sep {
			sep = minSep
			indexReferenceEdge = i
			penPoint = minVert
		}
	}
	return sep, indexReferenceEdge, penPoint
}

// edge returns the edge vector and vertices for a given edge index in a polygon
func edge(index int, polygon Polygon) (edge, v1, v2 vector.Two) {
	vertCount := len(polygon.WorldVertices)
	if vertCount == 0 {
		return vector.Two{}, vector.Two{}, vector.Two{}
	}
	nextIndex := (index + 1) % vertCount
	va := polygon.WorldVertices[index]
	vb := polygon.WorldVertices[nextIndex]
	return vb.Sub(va), va, vb
}
