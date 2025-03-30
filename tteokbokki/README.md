# Tteokbokki (physics)

A lightweight 2D physics engine for Go, designed to work seamlessly with the [Bappa Framework]. However, it can
work standalone.

## Features

- **Collision Detection**: Detect collisions between polygons and axis-aligned boxes
- **Physics Integration**: Update positions and rotations based on physical forces
- **Force Application**: Apply gravity, friction, and custom forces to objects
- **Impulse Resolution**: Handle realistic bouncing and momentum transfer
- **Specialized Resolvers**: Options for standard, vertical-only, and static object resolution

## Installation

```bash
go get github.com/TheBitDrifter/tteokbokki
```

## Quick Start

```go
package main

import (
    blueprintmotion "github.com/TheBitDrifter/blueprint/motion"
    blueprintspatial "github.com/TheBitDrifter/blueprint/spatial"
    "github.com/TheBitDrifter/blueprint/vector"
    "github.com/TheBitDrifter/tteokbokki/motion"
    "github.com/TheBitDrifter/tteokbokki/spatial"
)

func main() {
    // Create two objects with positions
    posA := blueprintspatial.NewPosition(100, 100)
    posB := blueprintspatial.NewPosition(120, 110)
    
    // Create shapes for collision
    shapeA := blueprintspatial.NewRectangle(50, 50)
    shapeB := blueprintspatial.NewRectangle(40, 40)
    
    // Update world vertices for collision detection
    shapeA.Polygon.WorldVertices = spatial.UpdateWorldVerticesSimple(
        shapeA.Polygon.LocalVertices, posA.Two)
    shapeB.Polygon.WorldVertices = spatial.UpdateWorldVerticesSimple(
        shapeB.Polygon.LocalVertices, posB.Two)
    
    // Create dynamics objects
    dynA := blueprintmotion.NewDynamics(1.0)
    dynA.Vel = vector.Two{X: 10.0, Y: 5.0}
    dynA.Elasticity = 0.5
    dynA.SetDefaultAngularMass(shapeA)
    
    // Create a heavier object
    dynB := blueprintmotion.NewDynamics(2.0)
    dynB.Elasticity = 0.3
    dynB.SetDefaultAngularMass(shapeB)
    
    // Check for collision
    if colliding, collision := spatial.Detector.Check(
        shapeA, shapeB, posA.Two, posB.Two,
    ); colliding {
        // Resolve the collision with physics
        motion.Resolver.Resolve(
            &posA.Two,
            &posB.Two,
            &dynA,
            &dynB,
            collision,
        )
        
        // Apply forces (like gravity)
        gravity := motion.Forces.Generator.NewGravityForce(1.0, 9.8, 100.0)
        motion.Forces.AddForce(&dynA, gravity)
        
        // Manually integrate for a time step
        newPosA := motion.IntegrateLinear(&dynA, &posA.Two, 0.016)
        posA.Two = newPosA
    }
}
```

## Core Concepts

### Systems

Physics simulation is handled by two main systems:

- **TransformSystem**: Updates world coordinates based on position, rotation, and scale
- **IntegrationSystem**: Integrates physical forces to update positions and rotations

### Physics Components

Entities can have these physics-related components:

- **Position**: Location in 2D space
- **Rotation**: Angular orientation
- **Dynamics**: Mass, velocity, forces, and physical properties
- **Shape**: Collision geometry with world and local coordinates

### Collision Handling

The engine provides multiple ways to handle collisions:

```go
// Standard collision detection
isColliding, collision := spatial.Detector.Check(shapeA, shapeB, posA, posB)

// Continuous collision detection for fast objects
isColliding, collision, timeOfImpact := spatial.ContinuousCollisionDetector.Check(
    shapeA, shapeB, posA, posB, prevPosA, prevPosB, steps
)

// Resolve collision with physics
motion.Resolver.Resolve(posA, posB, dynA, dynB, collision)

// Vertical-only resolution for platformers
motion.VerticalResolver.Resolve(posA, posB, dynA, dynB, collision)
```

### Force Management

Apply and manage forces through the global Forces handler:

```go
// Apply a force
motion.Forces.AddForce(dynamics, forceVector)

// Apply rotational force
motion.Forces.AddTorque(dynamics, torqueValue)

// Create common forces
gravityForce := motion.Forces.Generator.NewGravityForce(mass, gravity, pixelsPerMeter)
frictionForce := motion.Forces.Generator.NewFrictionForce(velocity, frictionCoefficient)
```

## License

MIT License - see the [LICENSE](LICENSE) file for details.
