package motion

import (
	blueprintmotion "github.com/TheBitDrifter/blueprint/motion"
	"github.com/TheBitDrifter/blueprint/vector"
	"github.com/TheBitDrifter/tteokbokki/spatial"
)

// VerticalResolver is the Global vertical collision resolver instance
var VerticalResolver verticalResolver

// verticalResolver handles collision resolution limited to vertical (Y-axis) movement
type verticalResolver struct{}

// Resolve handles collision resolution for vertical-only movement
func (r verticalResolver) Resolve(posA, posB *vector.Two, dynA, dynB *blueprintmotion.Dynamics, collision spatial.Collision) {
	r.resolvePositions(dynA, dynB, posA, posB, collision)
	r.applyResolutionImpulses(dynA, dynB, posA, posB, collision)
}

// resolvePositions corrects object positions considering only vertical (Y) components
func (verticalResolver) resolvePositions(dynA, dynB *blueprintmotion.Dynamics, posA, posB *vector.Two, collision spatial.Collision) {
	// Only consider Y component of the normal
	yOnlyNormal := vector.Two{X: 0, Y: collision.Normal.Y}
	if yOnlyNormal.Y != 0 {
		yOnlyNormal = yOnlyNormal.Norm()
		correctionA := collision.Depth / (dynA.InverseMass + dynB.InverseMass) * dynA.InverseMass
		correctionB := collision.Depth / (dynA.InverseMass + dynB.InverseMass) * dynB.InverseMass
		*posA = posA.Sub(yOnlyNormal.Scale(correctionA))
		*posB = posB.Add(yOnlyNormal.Scale(correctionB))
	}
}

// applyResolutionImpulses calculates and applies impulses considering only Y-axis components
func (verticalResolver) applyResolutionImpulses(dynA, dynB *blueprintmotion.Dynamics, posA, posB *vector.Two, collision spatial.Collision) {
	combinedElasticity := (dynA.Elasticity + dynB.Elasticity) / 2
	centerToImpactA := collision.End.Sub(*posA)
	centerToImpactB := collision.Start.Sub(*posB)

	// Only consider Y component of velocities
	relativeVelA := vector.Two{
		X: 0,
		Y: dynA.Vel.Y + dynA.AngularVel*centerToImpactA.X,
	}
	relativeVelB := vector.Two{
		X: 0,
		Y: dynB.Vel.Y + dynB.AngularVel*centerToImpactB.X,
	}
	impactVelocity := relativeVelA.Sub(relativeVelB)

	// Use Y-only normal
	yOnlyNormal := vector.Two{X: 0, Y: collision.Normal.Y}
	if yOnlyNormal.Y != 0 {
		yOnlyNormal = yOnlyNormal.Norm()
		normalVelocity := impactVelocity.ScalarProduct(yOnlyNormal)

		// Calculate rotational factors
		rotationFactorA := centerToImpactA.CrossProduct(yOnlyNormal)
		rotationFactorASq := rotationFactorA * rotationFactorA
		rotationFactorB := centerToImpactB.CrossProduct(yOnlyNormal)
		rotationFactorBSq := rotationFactorB * rotationFactorB

		// Compute normal impulse
		totalInverseMass := dynA.InverseMass + dynB.InverseMass
		normalImpulseDenom := totalInverseMass +
			rotationFactorASq*dynA.InverseAngularMass +
			rotationFactorBSq*dynB.InverseAngularMass
		normalImpulseMag := -(1 + combinedElasticity) * normalVelocity / normalImpulseDenom
		normalImpulse := yOnlyNormal.Scale(normalImpulseMag)

		// Apply impulses to both objects
		ApplyImpulse(dynA, normalImpulse, centerToImpactA)
		ApplyImpulse(dynB, normalImpulse.Scale(-1), centerToImpactB)
	}
}
