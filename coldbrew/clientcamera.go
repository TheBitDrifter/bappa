package coldbrew

// CameraManager handles the creation and management of camera objects
// and provides access to scene tracking capabilities
type CameraManager interface {
	// CameraSceneTracker returns the tracker for camera scene transitions
	CameraSceneTracker() CameraSceneTracker

	// Cameras returns the array of available cameras
	Cameras() [MaxSplit]Camera

	// ActivateCamera attempts to activate a camera and returns it if successful
	ActivateCamera() (Camera, error)
}
