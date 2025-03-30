package motion

import (
	"github.com/TheBitDrifter/bappa/blueprint/vector"
	"github.com/TheBitDrifter/bappa/tteokbokki/spatial"
)

// Resolver is the global collision resolver instance
var Resolver resolver

// resolver handles collision resolution between objects
type resolver struct{}

// Resolve splits the separation equally between both objects
func (r resolver) Resolve(posA, posB *vector.Two, dynA, dynB *Dynamics, collision spatial.Collision) {
	if dynA.InverseMass == 0 && dynB.InverseMass == 0 {
		return
	}
	r.resolvePositions(dynA, dynB, posA, posB, collision)
	r.applyResolutionImpulses(dynA, dynB, posA, posB, collision)
}

// resolvePositions corrects object positions based on collision depth and mass
func (resolver) resolvePositions(dynA, dynB *Dynamics, posA, posB *vector.Two, collision spatial.Collision) {
	correctionA := collision.Depth / (dynA.InverseMass + dynB.InverseMass) * dynA.InverseMass
	correctionB := collision.Depth / (dynA.InverseMass + dynB.InverseMass) * dynB.InverseMass
	*posA = posA.Sub(collision.Normal.Scale(correctionA))
	*posB = posB.Add(collision.Normal.Scale(correctionB))
}

// applyResolutionImpulses calculates and applies impulses to both objects
func (resolver) applyResolutionImpulses(dynA, dynB *Dynamics, posA, posB *vector.Two, collision spatial.Collision) {
	combinedElasticity := (dynA.Elasticity + dynB.Elasticity) / 2
	combinedFriction := (dynA.Friction + dynB.Friction) / 2
	centerToImpactA := collision.End.Sub(*posA)
	centerToImpactB := collision.Start.Sub(*posB)

	// Calculate relative velocities at impact points
	relativeVelA := dynA.Vel.Add(
		vector.Two{
			X: -dynA.AngularVel * centerToImpactA.Y,
			Y: dynA.AngularVel * centerToImpactA.X,
		})
	relativeVelB := dynB.Vel.Add(
		vector.Two{
			X: -dynB.AngularVel * centerToImpactB.Y,
			Y: dynB.AngularVel * centerToImpactB.X,
		})

	impactVelocity := relativeVelA.Sub(relativeVelB)
	normalVelocity := impactVelocity.ScalarProduct(collision.Normal)
	normalImpulseDir := collision.Normal

	// Calculate rotational factors
	rotationFactorA := centerToImpactA.CrossProduct(collision.Normal)
	rotationFactorASq := rotationFactorA * rotationFactorA
	rotationFactorB := centerToImpactB.CrossProduct(collision.Normal)
	rotationFactorBSq := rotationFactorB * rotationFactorB

	// Compute normal impulse
	totalInverseMass := dynA.InverseMass + dynB.InverseMass
	normalImpulseDenom := totalInverseMass +
		rotationFactorASq*dynA.InverseAngularMass +
		rotationFactorBSq*dynB.InverseAngularMass
	normalImpulseMag := -(1 + combinedElasticity) * normalVelocity / normalImpulseDenom
	normalImpulse := normalImpulseDir.Scale(normalImpulseMag)

	// Compute tangential impulse for friction
	tangentDir := collision.Normal.Perpendicular().Norm()
	tangentVelocity := impactVelocity.ScalarProduct(tangentDir)
	rotationFactorTangentA := centerToImpactA.CrossProduct(tangentDir)
	rotationFactorTangentASq := rotationFactorTangentA * rotationFactorTangentA
	rotationFactorTangentB := centerToImpactB.CrossProduct(tangentDir)
	rotationFactorTangentBSq := rotationFactorTangentB * rotationFactorTangentB

	tangentImpulseDenom := totalInverseMass +
		rotationFactorTangentASq*dynA.InverseAngularMass +
		rotationFactorTangentBSq*dynB.InverseAngularMass
	tangentImpulseMag := combinedFriction * -(1 + combinedElasticity) * tangentVelocity / tangentImpulseDenom
	tangentImpulse := tangentDir.Scale(tangentImpulseMag)

	// Apply combined impulses to both objects
	totalImpulseA := normalImpulse.Add(tangentImpulse)
	totalImpulseB := totalImpulseA.Scale(-1)
	ApplyImpulse(dynA, totalImpulseA, centerToImpactA)
	ApplyImpulse(dynB, totalImpulseB, centerToImpactB)
}
