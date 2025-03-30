package vector

import (
	"math"
)

// TwoFace provides a complete interface for 2D vector operations
type TwoFace interface {
	TwoWriter
	TwoReader
	TwoCalculator
}

// TwoWriter defines methods for modifying 2D vector values
type TwoWriter interface {
	SetFromInterface(TwoReader)
	SetX(x float64)
	SetY(y float64)
}

// TwoReader defines methods for accessing 2D vector values
type TwoReader interface {
	GetX() float64
	GetY() float64
}

// TwoCalculator defines methods for performing calculations with 2D vectors
type TwoCalculator interface {
	CloneAsInterface() TwoFace
	RotateAsInterface(radians float64) TwoFace
	AddAsInterface(TwoReader) TwoFace
	SubAsInterface(TwoReader) TwoFace
	CrossProductAsInterface(TwoReader) float64
	ScaleAsInterface(float64) TwoFace
}

// Two represents a 2D vector with X and Y components
type Two struct {
	X, Y float64
}

// Scale multiplies the vector by a scalar value
func (v2 Two) Scale(n float64) Two {
	return Two{v2.X * n, v2.Y * n}
}

// ScalarProduct calculates the dot product of two vectors
func (v2a Two) ScalarProduct(v2b Two) float64 {
	return (v2a.X * v2b.X) + (v2a.Y * v2b.Y)
}

// CrossProduct calculates the cross product of two vectors
func (v2a Two) CrossProduct(v2b Two) float64 {
	return (v2a.X * v2b.Y) - (v2a.Y * v2b.X)
}

// Add returns the sum of two vectors
func (v2a Two) Add(v2b Two) Two { return Two{X: v2a.X + v2b.X, Y: v2a.Y + v2b.Y} }

// Sub returns the difference of two vectors
func (v2a Two) Sub(v2b Two) Two { return Two{X: v2a.X - v2b.X, Y: v2a.Y - v2b.Y} }

// Perpendicular returns a vector perpendicular to this one
func (v2 Two) Perpendicular() Two {
	X := v2.X
	Y := v2.Y
	v2.X = Y
	v2.Y = -X
	return v2
}

// Mag returns the magnitude (length) of the vector
func (v2 Two) Mag() float64 {
	return math.Sqrt((v2.X * v2.X) + (v2.Y * v2.Y))
}

// MagSquared returns the squared magnitude of the vector
func (v2 Two) MagSquared() float64 {
	return (v2.X * v2.X) + (v2.Y * v2.Y)
}

// Norm returns a normalized (unit length) version of the vector
func (v2 Two) Norm() Two {
	len := v2.Mag()
	if len != 0 {
		v2.X = v2.X / len
		v2.Y = v2.Y / len
	}
	return v2
}

// Rotate rotates the vector by the specified angle in radians
func (v2 Two) Rotate(radians float64) Two {
	newX := v2.X*math.Cos(radians) - v2.Y*math.Sin(radians)
	newY := v2.X*math.Sin(radians) + v2.Y*math.Cos(radians)
	v2.X = newX
	v2.Y = newY
	return v2
}

// RotateAroundPoint rotates the vector around a specified point
func (v2 Two) RotateAroundPoint(radians float64, point Two) Two {
	origin := Two{X: 0, Y: 0}
	offset := origin.Sub(point)
	result := v2.Add(offset)
	result = result.Rotate(radians)
	result = result.Sub(offset)
	return result
}

// Equal checks if two vectors have identical components
func (v2 Two) Equal(v2b Two) bool {
	return v2.X == v2b.X && v2.Y == v2b.Y
}

// Clone returns a copy of the vector
func (v2 Two) Clone() Two {
	return Two{X: v2.X, Y: v2.Y}
}

// CloneAsInterface returns a copy of the vector as a TwoFace interface
func (v2 Two) CloneAsInterface() TwoFace {
	clone := v2.Clone()
	return &clone
}

// SetX sets the X component of the vector
func (v2 *Two) SetX(x float64) {
	v2.X = x
}

// SetY sets the Y component of the vector
func (v2 *Two) SetY(y float64) {
	v2.Y = y
}

// SetFromInterface sets vector components from a TwoReader interface
func (v2 *Two) SetFromInterface(tr TwoReader) {
	v2.X = tr.GetX()
	v2.Y = tr.GetY()
}

// GetX returns the X component of the vector
func (v2 Two) GetX() float64 {
	return v2.X
}

// GetY returns the Y component of the vector
func (v2 Two) GetY() float64 {
	return v2.Y
}

// RotateAsInterface rotates the vector and returns the result as a TwoFace interface
func (v2 Two) RotateAsInterface(radians float64) TwoFace {
	rotated := v2.Rotate(radians)
	return &rotated
}

// AddAsInterface adds another vector and returns the result as a TwoFace interface
func (v2 Two) AddAsInterface(v2B TwoReader) TwoFace {
	v2B2 := Two{
		X: v2B.GetX(),
		Y: v2B.GetY(),
	}
	added := v2.Add(v2B2)
	return &added
}

// SubAsInterface subtracts another vector and returns the result as a TwoFace interface
func (v2 Two) SubAsInterface(v2B TwoReader) TwoFace {
	v2B2 := Two{
		X: v2B.GetX(),
		Y: v2B.GetY(),
	}
	sub := v2.Sub(v2B2)
	return &sub
}

// CrossProductAsInterface calculates the cross product with another vector
func (v2 Two) CrossProductAsInterface(v2B TwoReader) float64 {
	v2B2 := Two{
		X: v2B.GetX(),
		Y: v2B.GetY(),
	}
	return v2.CrossProduct(v2B2)
}

// ScaleAsInterface scales the vector and returns the result as a TwoFace interface
func (v2 Two) ScaleAsInterface(n float64) TwoFace {
	scaled := v2.Scale(n)
	return &scaled
}
