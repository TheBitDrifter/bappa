# Blueprint

Blueprint is a foundation package for the [Bappa Framework](https://dl43t3h5ccph3.cloudfront.net/), providing common components, interfaces, and utilities for game development.

## Overview

Blueprint serves as the base layer of the Bappa Framework, defining standard component types and interfaces that other packages in the framework build upon. It establishes shared patterns and data structures used throughout the ecosystem.

## Features

- **Core Interfaces**: Defines essential interfaces like `Scene` and `CoreSystem`
- **Common Components**: Provides standardized component definitions across multiple domains
- **Vector Mathematics**: Complete 2D vector implementation with comprehensive operations
- **Predefined Queries**: Ready-to-use queries for common component combinations
- **Background Utilities**: Tools for creating static and parallax backgrounds
- **Shape Factory**: Various shape creation functions for collision geometry

## Installation

```bash
go get github.com/TheBitDrifter/blueprint
```

## Component Subpackages

Blueprint organizes components by domain:

- **blueprint/client**: Visual and audio components
  - `SpriteBundle`: Sprites with animation support
  - `SoundBundle`: Audio resources
  - `ParallaxBackground`: Multi-layered scrolling backgrounds

- **blueprint/input**: User interaction components
  - `InputBuffer`: Collection and management of user inputs
  - `StampedInput`: Inputs with timing and position information

- **blueprint/motion**: Physics components
  - `Dynamics`: Physical properties for movement and forces

- **blueprint/spatial**: Positioning and geometry components
  - `Position`, `Rotation`, `Scale`: Basic spatial properties
  - `Shape`: Collision geometry with various factory methods
  - `Direction`: Directional orientation

- **blueprint/vector**: 2D vector mathematics
  - `Two`: Vector with extensive operations (add, subtract, rotate, etc.)
  - Vector interfaces for flexible implementation

## Quick Start

### Using Predefined Queries

```go
// Create a cursor for entities with position components
cursor := scene.NewCursor(blueprint.Queries.Position)

// Process matching entities
for range cursor.Next() {
    pos := blueprintspatial.Components.Position.GetFromCursor(cursor)
    // Process entity...
}
```

### Creating Backgrounds

```go
// Create a parallax background with multiple layers
builder := blueprint.NewParallaxBackgroundBuilder(storage)
builder.AddLayer("backgrounds/mountains.png", 0.2, 0.0) // Slow-moving background
builder.AddLayer("backgrounds/clouds.png", 0.5, 0.1)    // Mid-speed layer
builder.WithOffset(vector.Two{X: 0, Y: 20})             // Optional offset
builder.Build()

// Create a static background
blueprint.CreateStillBackground(storage, "backgrounds/scene.png")
```

### Creating and Using Shapes

```go
// Create a rectangle
rect := blueprintspatial.NewRectangle(50, 30)

// Create a ramp
ramp := blueprintspatial.NewSingleRamp(100, 40, true) // ascending left-to-right

// Create a platform
platform := blueprintspatial.NewTrapezoidPlatform(80, 20, 0.7) // 0.7 = bottom width ratio
```

## License

MIT License - see the [LICENSE](LICENSE) file for details.
