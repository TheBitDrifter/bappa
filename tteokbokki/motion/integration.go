package motion

import (
	blueprintmotion "github.com/TheBitDrifter/blueprint/motion"
	vector "github.com/TheBitDrifter/blueprint/vector"
)

// Integrate updates position and rotation based on dynamics over the given time step
func Integrate(dyn *blueprintmotion.Dynamics, position vector.TwoReader, rotation, dt float64) (newPosition vector.Two, newRotation float64) {
	return IntegrateLinear(dyn, position, dt), IntegrateAngular(dyn, rotation, dt)
}

// IntegrateLinear calculates new position based on forces, acceleration and velocity
func IntegrateLinear(dyn *blueprintmotion.Dynamics, pos vector.TwoReader, dt float64) (newPos vector.Two) {
	posConc := vector.Two{
		X: pos.GetX(),
		Y: pos.GetY(),
	}
	if dyn.InverseMass == 0 {
		return posConc
	}
	dyn.Accel = dyn.SumForces.Scale(dyn.InverseMass)
	dyn.Vel = dyn.Vel.Add(dyn.Accel.Scale(dt))
	newPos = posConc.Add(dyn.Vel.Scale(dt))
	Forces.ClearForces(dyn)
	return newPos
}

// IntegrateAngular calculates new rotation based on torque, angular acceleration and velocity
func IntegrateAngular(dyn *blueprintmotion.Dynamics, rotation float64, dt float64) (newRotation float64) {
	if dyn.InverseAngularMass == 0 {
		return rotation
	}
	dyn.AngularAccel = dyn.SumTorque * dyn.InverseAngularMass
	dyn.AngularVel = dyn.AngularVel + dyn.AngularAccel*dt
	newRotation = rotation + dyn.AngularVel*dt
	Forces.ClearTorque(dyn)
	return newRotation
}
