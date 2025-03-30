/*
Package blueprint provides common components and interfaces for the Bappa Framework,
a game development system based on an entity-component-system (ECS) architecture.

Blueprint serves as the foundation layer of the Bappa Framework, defining standard
component types and interfaces that other packages in the framework build upon.
It establishes common patterns and data structures used throughout the ecosystem.

# Core Concepts

Blueprint defines these key elements:

  - Scene: A minimal interface for accessing entities and systems
  - CoreSystem: Interface for systems that operate on scenes
  - Common component types across various domains
  - Predefined queries for frequently used component combinations

# Subpackages

Blueprint organizes components by domain into several subpackages:

  - blueprint/client: Visual and audio components (sprites, animations, sounds)
  - blueprint/input: User interaction components (input buffers)
  - blueprint/motion: Physical movement components (dynamics)
  - blueprint/spatial: Positioning and geometry components (position, rotation, shapes)
  - blueprint/vector: 2D vector mathematics utilities

# Working with Queries

The package provides predefined queries for common component combinations:

	// Access entities with position components
	cursor := scene.NewCursor(blueprint.Queries.Position)

	// Access entities with dynamics components
	cursor = scene.NewCursor(blueprint.Queries.Dynamics)

	// Access entities with sprite components
	cursor = scene.NewCursor(blueprint.Queries.SpriteBundle)

# Background Utilities

The package includes functions for creating backgrounds and parallax effects:

	// Create a parallax background with multiple layers
	builder := blueprint.NewParallaxBackgroundBuilder(storage)
	builder.AddLayer("backgrounds/mountains.png", 0.2, 0.0)
	builder.AddLayer("backgrounds/clouds.png", 0.5, 0.1)
	builder.Build()

	// Create a static background
	blueprint.CreateStillBackground(storage, "backgrounds/scene.png")
*/
package blueprint
