/*
Package coldbrew provides a game client and scene management system for the Bappa Framework.

Coldbrew handles game lifecycle, rendering, input processing, scene transitions, and resource management.
It serves as the top-level interface for game developers, coordinating between the various components
of the Bappa ecosystem.

Core Concepts:

  - Client: The main engine that manages game state, rendering, and input
  - Scene: Isolated game worlds with their own entities, systems, and logic
  - Camera: Viewports that render portions of scenes
  - Systems: Components that process game logic, input, and rendering
  - Receivers: Handle input from various devices (keyboard, mouse, gamepad, touch)

Basic Usage:

	package main

	import (
		"embed"
		"log"

		"github.com/TheBitDrifter/blueprint"
		"github.com/TheBitDrifter/coldbrew"
	)

	//go:embed assets/*
	var assets embed.FS

	func main() {
		// Create a new client with resolution 640x360
		// maxSpritesCached=100, maxSoundsCached=50, maxScenesCached=10
		client := coldbrew.NewClient(640, 360, 100, 50, 10, assets)

		// Configure the client
		client.SetTitle("My Game")
		client.SetWindowSize(1280, 720)
		client.SetResizable(true)

		// Register a scene
		err := client.RegisterScene(
			"MainScene",            // Scene name
			640, 360,               // Scene dimensions
			mainScenePlan,          // Blueprint Plan
			renderSystems,          // Render systems
			clientSystems,          // Client systems
			coreSystems,            // Core systems
		)
		if err != nil {
			log.Fatal(err)
		}

		// Register global systems
		client.RegisterGlobalRenderSystem(rendersystems.GlobalRenderer{})
		client.RegisterGlobalClientSystem(clientsystems.InputBufferSystem{})

		// Setup cameras and inputs
		client.ActivateCamera()
		receiver, _ := client.ActivateReceiver()

		// Start the game
		if err := client.Start(); err != nil {
			log.Fatal(err)
		}
	}

Coldbrew organizes systems into five categories, running in this order each frame:

 1. Global Client Systems: Handle input and state across all scenes
 2. Core Systems: Process game simulation (physics, AI, collision)
 3. Scene Client Systems: Process scene-specific game logic
 4. Global Render Systems: Handle rendering across all scenes
 5. Scene Render Systems: Handle rendering specific to a scene

The framework supports multiple simultaneous active scenes, enabling features like:

  - Split-screen multiplayer
  - UI overlays
  - Level transitions
  - Mini-maps and picture-in-picture views

Coldbrew uses an input abstraction layer that separates physical inputs (keyboard, mouse, gamepad)
from gameplay actions, allowing for flexible control schemes and device independence.

Coldbrew is built on top of the Ebiten game engine (https://github.com/hajimehoshi/ebiten), which provides
the low-level graphics rendering, input detection, and cross-platform functionality. Coldbrew extends
Ebiten with a comprehensive entity-component system and game management features while integrating the
other components of the Bappa Framework: Blueprint, Warehouse, Tteokbokki, and Mask.
*/
package coldbrew
