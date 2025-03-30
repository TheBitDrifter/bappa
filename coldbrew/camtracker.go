package coldbrew

// CameraSceneTracker tracks which scene each camera is rendering and when it changed
type CameraSceneTracker map[Camera]CameraSceneRecord

// CameraSceneRecord stores information about a camera's current scene and tick
type CameraSceneRecord struct {
	Scene Scene
	Tick  int
}
