package motion

import (
	blueprint_motion "github.com/TheBitDrifter/blueprint/motion"
	blueprint_vector "github.com/TheBitDrifter/blueprint/vector"
)

// ApplyImpulse applies both linear and angular impulse to a dynamics object
func ApplyImpulse(dyn *blueprint_motion.Dynamics, linearImpulse, torqueArm blueprint_vector.Two) {
	linearImpulseScaled := linearImpulse.Scale(dyn.InverseMass)
	dyn.Vel = dyn.Vel.Add(linearImpulseScaled)
	angularImpulseScaled := torqueArm.CrossProduct(linearImpulse) * dyn.InverseAngularMass
	dyn.AngularVel = dyn.AngularVel + angularImpulseScaled
}
