package clientsystems

import (
	"math"

	"github.com/TheBitDrifter/blueprint"
	blueprintclient "github.com/TheBitDrifter/blueprint/client"
	"github.com/TheBitDrifter/coldbrew"
)

// BackgroundScrollSystem handles parallax scrolling effects for backgrounds based on camera movement
type BackgroundScrollSystem struct{}

func (BackgroundScrollSystem) Run(cli coldbrew.LocalClient, scene coldbrew.Scene) error {
	const movementThreshold = 0.01 // Minimum movement to register
	const precision = 0.25         // Round to nearest quarter-pixel (0.25, 0.5, 0.75, 1.0)

	activeCameras := cli.ActiveCamerasFor(scene)
	for _, cam := range activeCameras {
		cursor := scene.NewCursor(blueprint.Queries.ParallaxBackground)
		_, worldPosition := cam.Positions()

		for range cursor.Next() {
			config := blueprintclient.Components.ParallaxBackground.GetFromCursor(cursor)

			// Calculate raw new translations
			rawX := worldPosition.X * config.SpeedX * -1
			rawY := worldPosition.Y * config.SpeedY * -1

			// Round to nearest precision step
			newX := math.Round(rawX/precision) * precision
			newY := math.Round(rawY/precision) * precision

			// Current translations
			currentX := config.RelativeTranslations[cam.Index()].X
			currentY := config.RelativeTranslations[cam.Index()].Y

			// Only update if change exceeds threshold
			if math.Abs(newX-currentX) > movementThreshold {
				config.RelativeTranslations[cam.Index()].X = newX
			}

			if math.Abs(newY-currentY) > movementThreshold {
				config.RelativeTranslations[cam.Index()].Y = newY
			}
		}
	}
	return nil
}
