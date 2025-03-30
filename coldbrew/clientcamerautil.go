package coldbrew

import (
	"fmt"

	"github.com/hajimehoshi/ebiten/v2"
)

// CameraUtility provides methods for managing camera states and retrieving
// active cameras for scenes
type CameraUtility interface {
	// ActiveCamerasFor returns all active cameras assigned to the given scene
	ActiveCamerasFor(Scene) []Camera

	// Ready determines if a camera is ready to be used based on timing constraints
	Ready(Camera) bool
}

type cameraUtility struct {
	cameras            [MaxSplit]Camera
	cameraSceneTracker CameraSceneTracker
}

// newCameraUtility creates and initializes a new camera utility with default cameras
func newCameraUtility() *cameraUtility {
	cm := &cameraUtility{
		cameraSceneTracker: CameraSceneTracker{},
	}
	for k := range cm.cameras {
		cm.cameras[k] = &camera{
			index: k,
			surface: &sprite{
				image: ebiten.NewImage(1, 1),
				name:  fmt.Sprintf("camera %d", k+1),
			},
		}
	}
	return cm
}

// ActiveCamerasFor returns all active cameras that are assigned to the specified scene
func (cm *cameraUtility) ActiveCamerasFor(scene Scene) []Camera {
	result := []Camera{}
	for _, cam := range cm.cameras {
		if !cam.Active() {
			continue
		}
		sceneRecord, ok := cm.cameraSceneTracker[cam]
		if !ok {
			continue
		}
		if sceneRecord.Scene == scene {
			result = append(result, cam)
		}
	}
	return result
}

// Ready checks if a camera is ready to be used based on timing constraints
// and its active status
func (cm *cameraUtility) Ready(c Camera) bool {
	sceneRecord, ok := cm.cameraSceneTracker[c]
	if !ok {
		return false
	}
	cameraLastChanged := sceneRecord.Tick
	cutoff := 0
	if ClientConfig.enforceMinOnActive {
		cutoff = cameraLastChanged
	} else {
		cutoff = sceneRecord.Scene.LastActivatedTick()
	}
	return tick-cutoff >= ClientConfig.minimumLoadTime && c.Active()
}
