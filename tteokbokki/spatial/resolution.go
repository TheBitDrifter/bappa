package spatial

import (
	blueprint_spatial "github.com/TheBitDrifter/blueprint/spatial"
	blueprint_vector "github.com/TheBitDrifter/blueprint/vector"
)

// Resolver is the global collision resolver instance (without physics)
var Resolver resolver

// resolver handles collision resolution between objects (without physics)
type resolver struct{}

// Resolve splits the separation equally between both objects
func (resolver) Resolve(posA, posB *blueprint_vector.Two, collision Collision) {
	correction := collision.Normal.Scale(collision.Depth / 2.0)
	*posA = posA.Sub(correction)
	*posB = posB.Add(correction)
}

// ResolveAStatic resolves the collision by only moving object B, treating A as immovable
func (resolver) ResolveAStatic(shapeA, shapeB blueprint_spatial.Shape, posA, posB *blueprint_vector.Two, collision Collision) {
	correction := collision.Normal.Scale(collision.Depth)
	*posB = posB.Add(correction)
}

// ResolveBStatic resolves the collision by only moving object A, treating B as immovable
func (resolver) ResolveBStatic(shapeA, shapeB blueprint_spatial.Shape, posA, posB *blueprint_vector.Two, collision Collision) {
	correction := collision.Normal.Scale(collision.Depth)
	*posA = posA.Sub(correction)
}
