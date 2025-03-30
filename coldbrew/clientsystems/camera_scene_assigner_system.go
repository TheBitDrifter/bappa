package clientsystems

import (
	"maps"

	"github.com/TheBitDrifter/blueprint"
	blueprintclient "github.com/TheBitDrifter/blueprint/client"
	"github.com/TheBitDrifter/blueprint/vector"
	"github.com/TheBitDrifter/coldbrew"
)

// CameraSceneAssignerSystem manages cameras across scenes.
// It uses the 'CameraIndex' component to determine which cameras should be active.
// When all scenes are inactive (loading or similar states), the system will
// only keep the first camera active to display loaders or other UI elements.
type CameraSceneAssignerSystem struct {
	prevScenes            coldbrew.CameraSceneTracker
	originalCamDimensions vector.Two
	originalCamScreenPos  vector.Two
}

// Run executes the camera scene assignment logic
func (sys *CameraSceneAssignerSystem) Run(cli coldbrew.Client) error {
	// Initialize previous scenes tracking if needed
	if sys.prevScenes == nil {
		sys.prevScenes = maps.Clone(cli.CameraSceneTracker())
	}
	defer func() {
		sys.prevScenes = maps.Clone(cli.CameraSceneTracker())
	}()

	cameras := cli.Cameras()

	// Track active cameras and assign them to scenes
	allScenesInactive := sys.processCamerasForActiveScenes(cli, cameras)

	// Deactivate cameras not assigned to any scene
	sys.deactivateUnusedCameras(cli, cameras)

	// Count active cameras
	activeCameraCount := countActiveCameras(cameras)

	// Handle scene transitions
	scenesChanging := sys.handleSceneTransitions(cli, cameras, activeCameraCount)

	// If all scenes are inactive or changing, keep only the first camera active
	if (allScenesInactive || scenesChanging) && activeCameraCount > 0 {
		sys.setDefaultCameraState(cli, cameras)
	} else if sys.originalCamDimensions.X != 0 && sys.originalCamDimensions.Y != 0 {
		cameras[0].SetDimensions(int(sys.originalCamDimensions.X), int(sys.originalCamDimensions.Y))
		screenPos, _ := cameras[0].Positions()
		screenPos.X = sys.originalCamScreenPos.X
		screenPos.Y = sys.originalCamScreenPos.Y
	}

	return nil
}

// processCamerasForActiveScenes assigns cameras to active scenes and returns whether all scenes are inactive
func (sys *CameraSceneAssignerSystem) processCamerasForActiveScenes(cli coldbrew.Client, cameras [coldbrew.MaxSplit]coldbrew.Camera) bool {
	allScenesInactive := true

	for scene := range cli.ActiveScenes() {
		if !scene.Ready() {
			continue
		}

		// At least one scene is active
		allScenesInactive = false

		// Assign cameras to this scene
		cameraIndexCursor := scene.NewCursor(blueprint.Queries.CameraIndex)
		for range cameraIndexCursor.Next() {
			camIndex := *blueprintclient.Components.CameraIndex.GetFromCursor(cameraIndexCursor)
			cam := cameras[camIndex]
			entry := cli.CameraSceneTracker()[cam]
			entry.Scene = scene
			cli.CameraSceneTracker()[cam] = entry
			cam.Activate()
		}
	}

	return allScenesInactive
}

// deactivateUnusedCameras turns off cameras not assigned to any scene
func (sys *CameraSceneAssignerSystem) deactivateUnusedCameras(cli coldbrew.Client, cameras [coldbrew.MaxSplit]coldbrew.Camera) {
	for _, cam := range cameras {
		_, isAssigned := cli.CameraSceneTracker()[cam]
		if !isAssigned {
			cam.Deactivate()
		}
	}
}

// handleSceneTransitions detects and manages cameras during scene transitions
// Returns true if all active cameras are in a transition state
func (sys *CameraSceneAssignerSystem) handleSceneTransitions(cli coldbrew.Client, cameras [coldbrew.MaxSplit]coldbrew.Camera, activeCameraCount int) bool {
	camerasInTransition := 0
	minimumLoadTime := coldbrew.ClientConfig.MinimumLoadTime()

	for _, cam := range cameras {
		prevSceneTracker := sys.prevScenes[cam]
		currSceneTracker := cli.CameraSceneTracker()[cam]

		// Detect scene change and update transition timing
		if prevSceneTracker.Scene != currSceneTracker.Scene {
			currSceneTracker.Tick = cli.CurrentTick()
			cli.CameraSceneTracker()[cam] = currSceneTracker
		}

		// Check if camera is within transition period
		sceneChangeTick := currSceneTracker.Tick
		if cli.CurrentTick()-sceneChangeTick < minimumLoadTime && sceneChangeTick != 0 {
			camerasInTransition++
		}
	}

	return camerasInTransition == activeCameraCount && activeCameraCount > 0
}

// setDefaultCameraState sets only the first camera to active and resets resolution
func (sys *CameraSceneAssignerSystem) setDefaultCameraState(cli coldbrew.Client, cameras [coldbrew.MaxSplit]coldbrew.Camera) {
	// Deactivate all cameras
	for _, cam := range cameras {
		cam.Deactivate()
	}
	// Save original dimensions and position
	screenPos, _ := cameras[0].Positions()
	if sys.originalCamDimensions.X == 0 || sys.originalCamDimensions.Y == 0 {
		w, h := cameras[0].Dimensions()
		sys.originalCamDimensions.X = float64(w)
		sys.originalCamDimensions.Y = float64(h)
		sys.originalCamScreenPos = *screenPos
	}
	// Activate only the first camera
	cameras[0].Activate()

	// Reset resolution to base values
	rx, ry := coldbrew.ClientConfig.BaseResolution()
	cli.SetResolution(rx, ry)

	// Full Screen
	cameras[0].SetDimensions(rx, ry)
	screenPos.X = 0
	screenPos.Y = 0
}

// countActiveCameras returns the number of currently active cameras
func countActiveCameras(cameras [coldbrew.MaxSplit]coldbrew.Camera) int {
	count := 0
	for _, cam := range cameras {
		if cam.Active() {
			count++
		}
	}
	return count
}
