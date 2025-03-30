package motion

import (
	"fmt"
	"math"

	blueprintspatial "github.com/TheBitDrifter/blueprint/spatial"
	"github.com/TheBitDrifter/blueprint/vector"
)

// Dynamics represents the physical properties and state of a movable entity
type Dynamics struct {
	UnstoppableLinear, UnstoppableAngular bool
	Accel, Vel, SumForces                 vector.Two
	InverseMass                           float64
	AngularVel, AngularAccel, SumTorque   float64
	InverseAngularMass                    float64
	Friction, Elasticity                  float64
}

// NewDynamics creates a new Dynamics component with the specified mass
func NewDynamics(mass float64) Dynamics {
	inverseMass := mass
	if mass != 0 {
		inverseMass = 1 / mass
	}
	return Dynamics{
		Accel:       vector.Two{},
		Vel:         vector.Two{},
		SumForces:   vector.Two{},
		InverseMass: inverseMass,
	}
}

// SetMass updates the inverse mass based on the provided mass
func (dyn *Dynamics) SetMass(mass float64) {
	if mass == 0 {
		dyn.InverseMass = mass
	} else {
		dyn.InverseMass = 1 / mass
	}
}

// SetAngularMass updates the inverse angular mass based on the provided angular mass
func (dyn *Dynamics) SetAngularMass(angularMass float64) {
	if angularMass == 0 {
		dyn.InverseAngularMass = angularMass
	} else {
		dyn.InverseAngularMass = 1 / angularMass
	}
}

// SetDefaultAngularMass calculates and sets appropriate angular mass based on shape type
func (dyn *Dynamics) SetDefaultAngularMass(shape blueprintspatial.Shape) error {
	isAAB := shape.LocalAAB.Height != 0
	if isAAB {
		angularMass := getMomentOfInertiaWithoutMassRect(shape.WorldAAB)
		dyn.SetAngularMass(angularMass)
		return nil
	}
	isPoly := len(shape.Polygon.LocalVertices) > 0
	if isPoly {
		angularMass := getMomentOfInertiaWithoutMassPolygon(shape.Polygon)
		dyn.SetAngularMass(angularMass)
		return nil
	}
	return fmt.Errorf("unable to calc and set angular mass")
}

// getMomentOfInertiaWithoutMassRect calculates moment of inertia for a rectangular shape
func getMomentOfInertiaWithoutMassRect(aab blueprintspatial.AAB) float64 {
	return 0.083333 * ((aab.Width * aab.Width) + (aab.Height * aab.Height))
}

// getMomentOfInertiaWithoutMassPolygon calculates moment of inertia for a polygon shape
func getMomentOfInertiaWithoutMassPolygon(p blueprintspatial.Polygon) float64 {
	centroid := centroid(p)
	localVertices := p.LocalVertices

	var acc0, acc1 float64
	for i := 0; i < len(localVertices); i++ {
		a := localVertices[i].Sub(centroid)
		b := localVertices[(i+1)%len(localVertices)].Sub(centroid)
		cross := math.Abs(a.CrossProduct(b))
		acc0 += cross * (a.ScalarProduct(a) + b.ScalarProduct(b) + a.ScalarProduct(b))
		acc1 += cross
	}

	return acc0 / 6 / acc1
}

// centroid calculates the geometric center of a polygon
func centroid(p blueprintspatial.Polygon) vector.Two {
	localVertices := p.LocalVertices
	centroid := localVertices[0].Clone()
	centroid.X = 0
	centroid.Y = 0

	for i := 0; i < len(localVertices); i++ {
		j := (i + 1) % len(localVertices)
		cross := localVertices[i].CrossProduct(localVertices[j])
		sum := localVertices[i].Add(localVertices[j]).Scale(cross)
		centroid = centroid.Add(sum)
	}

	return centroid.Scale(1.0 / 6.0 / area(p))
}

// area calculates the area of a polygon
func area(p blueprintspatial.Polygon) float64 {
	localVertices := p.LocalVertices
	area := 0.0

	for i := 0; i < len(localVertices); i++ {
		j := (i + 1) % len(localVertices)
		area += localVertices[i].CrossProduct(localVertices[j])
	}

	return area / 2.0
}
