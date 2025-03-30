package coldbrew

// RenderSystem handles rendering at the scene level and runs after global render systems when its parent scene is active
type RenderSystem interface {
	Render(Scene, Screen, CameraUtility)
}

// GlobalRenderSystem handles rendering at the global/client level with access to all scenes
type GlobalRenderSystem interface {
	Render(Client, Screen)
}

// GlobalClientSystem runs in the update loop with access to all scenes via client
// Typically handles sound processing and transforms raw inputs into a usable state for core simulation systems
// Can  run before or after blueprint.CoreSystems depending on registration type
type GlobalClientSystem interface {
	Run(Client) error
}

// ClientSystem runs in the update loop at the scene level
// Processes local sound effects and input handling without access to other scenes
type ClientSystem interface {
	Run(LocalClient, Scene) error
}
