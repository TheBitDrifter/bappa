package motion

import (
	"math"

	blueprintmotion "github.com/TheBitDrifter/blueprint/motion"
	"github.com/TheBitDrifter/blueprint/vector"
)

// Forces is the global forces handler instance
var Forces forcesHandler

// forcesHandler manages force and torque operations on dynamics objects
type forcesHandler struct {
	Generator forcesGenerator
}

// AddForce applies a force vector to the dynamics object
func (forcesHandler) AddForce(dyn *blueprintmotion.Dynamics, force vector.TwoReader) {
	forceConc := vector.Two{
		X: force.GetX(),
		Y: force.GetY(),
	}
	dyn.SumForces = dyn.SumForces.Add(forceConc)
}

// ClearForces resets the accumulated forces to zero
func (forcesHandler) ClearForces(dyn *blueprintmotion.Dynamics) {
	dyn.SumForces = vector.Two{}
}

// AddTorque applies rotational force to the dynamics object
func (forcesHandler) AddTorque(dyn *blueprintmotion.Dynamics, torque float64) {
	dyn.SumTorque = dyn.SumTorque + torque
}

// ClearTorque resets the accumulated torque to zero
func (forcesHandler) ClearTorque(dyn *blueprintmotion.Dynamics) {
	dyn.SumTorque = 0
}

// forcesGenerator creates common physics forces
type forcesGenerator struct{}

// NewGravityForce creates a gravity force vector based on mass and gravity constants
func (forcesGenerator) NewGravityForce(mass, gravity, pixelsPerMeter float64) vector.Two {
	return vector.Two{X: 0.0, Y: mass * gravity * pixelsPerMeter}
}

// NewFrictionForce creates a friction force opposing the direction of movement
func (forcesGenerator) NewFrictionForce(velocity vector.Two, frictionCoefficient float64) vector.Two {
	// Calculate velocity magnitude
	velMagSquared := velocity.MagSquared()

	// If velocity is very small, just return a force that will zero it out completely
	if velMagSquared < 1.0 {
		// Return a force strong enough to zero out the velocity this frame
		return vector.Two{
			X: -velocity.X * 10.0, // Multiplier to ensure quick stopping
			Y: -velocity.Y * 10.0,
		}
	}

	// For normal velocities, scale the friction by a higher coefficient and by velocity
	// This makes friction much stronger overall and proportional to speed
	scaledCoefficient := frictionCoefficient * 10.0 // Make friction much stronger

	return vector.Two{
		X: -velocity.X * scaledCoefficient,
		Y: -velocity.Y * scaledCoefficient,
	}
}

// When friction isn't enough (0.9 for example will reduce vel 10% per frame)
func (forcesGenerator) ApplyVelocityDamping(dyn *blueprintmotion.Dynamics, dampingFactor float64) {
	// Apply a strong damping factor to slow down quickly

	// Apply damping
	dyn.Vel = dyn.Vel.Scale(dampingFactor)

	// Zero out very small velocities completely
	if dyn.Vel.MagSquared() < 0.5 {
		dyn.Vel = vector.Two{X: 0, Y: 0}
	}

	// Also damp angular velocity
	dyn.AngularVel = dyn.AngularVel * dampingFactor
	if math.Abs(dyn.AngularVel) < 0.01 {
		dyn.AngularVel = 0
	}
}

// NewHorizontalFrictionForce creates a friction force that only affects horizontal (X) movement
func (forcesGenerator) NewHorizontalFrictionForce(velocity vector.Two, frictionCoefficient float64) vector.Two {
	// Only use the X component of velocity
	xVelSquared := velocity.X * velocity.X

	// If X velocity is very small, return a force that will zero it out completely
	if xVelSquared < 1.0 {
		return vector.Two{
			X: -velocity.X * 10.0, // Strong multiplier to ensure quick stopping
			Y: 0,                  // No vertical component
		}
	}

	// For normal velocities, scale the friction by a higher coefficient
	scaledCoefficient := frictionCoefficient * 10.0

	return vector.Two{
		X: -velocity.X * scaledCoefficient,
		Y: 0, // No vertical component
	}
}

// ApplyHorizontalDamping reduces only the horizontal velocity component
func (forcesGenerator) ApplyHorizontalDamping(dyn *blueprintmotion.Dynamics, dampingFactor float64) {
	// Apply damping to X component only
	dyn.Vel.X = dyn.Vel.X * dampingFactor

	// Zero out very small horizontal velocities completely
	if math.Abs(dyn.Vel.X) < 0.5 {
		dyn.Vel.X = 0
	}

	// Also damp angular velocity
	dyn.AngularVel = dyn.AngularVel * dampingFactor
	if math.Abs(dyn.AngularVel) < 0.01 {
		dyn.AngularVel = 0
	}
}
