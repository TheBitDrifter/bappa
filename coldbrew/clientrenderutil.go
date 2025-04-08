package coldbrew

import (
	"fmt"

	"github.com/hajimehoshi/ebiten/v2"
)

type RenderUtility interface {
	// ActiveCamerasFor returns all active cameras assigned to the given scene
	ActiveCamerasFor(Scene) []Camera

	// Ready determines if a camera is ready to be used based on timing constraints
	Ready(Camera) bool
}

type renderUtility struct {
	cameras            [MaxSplit]Camera
	cameraSceneTracker CameraSceneTracker
}

// newRenderUtility creates and initializes a new camera utility with default cameras
func newRenderUtility() *renderUtility {
	u := &renderUtility{
		cameraSceneTracker: CameraSceneTracker{},
	}
	for k := range u.cameras {
		u.cameras[k] = &camera{
			index: k,
			surface: &sprite{
				image: ebiten.NewImage(1, 1),
				name:  fmt.Sprintf("camera %d", k+1),
			},
		}
	}
	return u
}

// ActiveCamerasFor returns all active cameras that are assigned to the specified scene
func (u *renderUtility) ActiveCamerasFor(scene Scene) []Camera {
	result := []Camera{}
	for _, cam := range u.cameras {
		if !cam.Active() {
			continue
		}
		sceneRecord, ok := u.cameraSceneTracker[cam]
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
func (u *renderUtility) Ready(c Camera) bool {
	sceneRecord, ok := u.cameraSceneTracker[c]
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
