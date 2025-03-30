package clientsystems

import (
	"github.com/TheBitDrifter/blueprint"
	blueprintclient "github.com/TheBitDrifter/blueprint/client"
	blueprintinput "github.com/TheBitDrifter/blueprint/input"
	"github.com/TheBitDrifter/coldbrew"
	"github.com/TheBitDrifter/warehouse"
)

// InputBufferSystem extracts client inputs and passes them to the core system components as StampedInputs
type InputBufferSystem struct{}

// Run processes all active input buffers across scenes
func (InputBufferSystem) Run(cli coldbrew.Client) error {
	for scene := range cli.ActiveScenes() {
		inputBufferCursor := warehouse.Factory.NewCursor(blueprint.Queries.InputBuffer, scene.Storage())
		for range inputBufferCursor.Next() {
			buffer := blueprintinput.Components.InputBuffer.GetFromCursor(inputBufferCursor)
			receiver := cli.Receiver(buffer.ReceiverIndex)
			if !receiver.Active() {
				continue
			}
			poppedInputs := receiver.PopInputs()

			// Transform input coordinates if camera component exists
			hasCam := blueprintclient.Components.CameraIndex.CheckCursor(inputBufferCursor)
			if hasCam {
				camIndex := *blueprintclient.Components.CameraIndex.GetFromCursor(inputBufferCursor)
				cam := cli.Cameras()[camIndex]
				if cam.Active() {
					globalPos, localPos := cam.Positions()
					// Convert global coordinates to local camera space
					for i, sInput := range poppedInputs {
						localX := int(localPos.X + float64(sInput.X) - globalPos.X)
						localY := int(localPos.Y + float64(sInput.Y) - globalPos.Y)
						poppedInputs[i].LocalX = localX
						poppedInputs[i].LocalY = localY
					}
				}
			}
			buffer.AddBatch(poppedInputs)
		}
	}
	return nil
}
