package motion

import "github.com/TheBitDrifter/bappa/blueprint/vector"

// ApplyImpulse applies both linear and angular impulse to a dynamics object
func ApplyImpulse(dyn *Dynamics, linearImpulse, torqueArm vector.Two) {
	linearImpulseScaled := linearImpulse.Scale(dyn.InverseMass)
	dyn.Vel = dyn.Vel.Add(linearImpulseScaled)
	angularImpulseScaled := torqueArm.CrossProduct(linearImpulse) * dyn.InverseAngularMass
	dyn.AngularVel = dyn.AngularVel + angularImpulseScaled
}
