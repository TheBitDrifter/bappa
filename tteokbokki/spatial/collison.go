package spatial

import (
	"github.com/TheBitDrifter/bappa/blueprint/vector"
)

// Collision represents information about a collision between two objects
type Collision struct {
	Start, End, Normal vector.Two
	Depth              float64
	CollidingEdgeA     CollisionEdge
	CollidingEdgeB     CollisionEdge
}

// CollisionEdge represents an edge involved in a collision
type CollisionEdge struct {
	Index    int
	Vertices []vector.Two
}

// IsTop determines if the collision occurred from the top of object A relative to object B
// In a coordinate system where lower Y values are higher up, a top collision normal has Y < 0
func (c Collision) IsTop() bool {
	return c.Normal.Y < 0
}

// IsTopB determines if the collision occurred from the top of object B relative to object A
// From B's perspective, the collision normal direction is inverted
func (c Collision) IsTopB() bool {
	// Invert the normal to get the direction from B to A
	return c.Normal.Y > 0 // Opposite of IsTop()
}
