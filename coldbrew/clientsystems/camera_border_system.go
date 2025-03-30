package clientsystems

import (
	"github.com/TheBitDrifter/coldbrew"
)

// CameraBorderLockSystem ensures cameras stay within scene boundaries
type CameraBorderLockSystem struct{}

// Run implements the system that constrains camera movement to scene boundaries
func (sys CameraBorderLockSystem) Run(cli coldbrew.LocalClient, scene coldbrew.Scene) error {
	cameras := cli.ActiveCamerasFor(scene)
	for _, cam := range cameras {
		sceneWidth := scene.Width()
		sceneHeight := scene.Height()
		camWidth, camHeight := cam.Dimensions()

		// Calculate maximum X position to keep camera within scene bounds
		maxX := sceneWidth - camWidth
		_, localPos := cam.Positions()

		// Constrain camera X position to scene boundaries
		if localPos.X > float64(maxX) {
			localPos.X = float64(maxX)
		}
		if localPos.X < 0 {
			localPos.X = 0
		}

		// Calculate and constrain camera Y position
		maxY := (float64(sceneHeight)) - float64(camHeight)
		if localPos.Y > float64(maxY) {
			localPos.Y = float64(maxY)
		}
		if localPos.Y < 0 {
			localPos.Y = 0
		}
	}
	return nil
}
