package blueprintspatial

import vector "github.com/TheBitDrifter/blueprint/vector"

// Position represents a 2D point in space
type Position struct{ vector.Two }

// PreviousPosition stores the previous position for continuous collision detection
type PreviousPosition struct{ vector.Two }

// Rotation represents an angle in radians
type Rotation float64

// AsFloat64 returns the rotation value as a float64
func (r Rotation) AsFloat64() float64 {
	return float64(r)
}

// Scale represents scaling factors for x and y dimensions
type Scale struct{ vector.Two }

// NewPosition creates a new Position at the specified coordinates
func NewPosition(x, y float64) Position {
	return Position{vector.Two{X: x, Y: y}}
}

// NewScale creates a new Scale with the specified x and y scaling factors
func NewScale(x, y float64) Scale {
	return Scale{vector.Two{X: x, Y: y}}
}

func (r *Rotation) Set(val float64) {
	*r = Rotation(val)
}
