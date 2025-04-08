package coldbrew_clientsystems

import (
	"github.com/TheBitDrifter/bappa/blueprint"
	"github.com/TheBitDrifter/bappa/blueprint/client"
	"github.com/TheBitDrifter/bappa/blueprint/input"
	"github.com/TheBitDrifter/bappa/coldbrew"
	"github.com/TheBitDrifter/bappa/warehouse"
)

// InputBufferSystem extracts client inputs and passes them to the core system components as StampedActions
type InputBufferSystem struct{}

func (InputBufferSystem) Run(cli coldbrew.Client) error {
	for scene := range cli.ActiveScenes() {
		actionBufferCursor := warehouse.Factory.NewCursor(blueprint.Queries.ActionBuffer, scene.Storage())
		for range actionBufferCursor.Next() {
			buffer := input.Components.ActionBuffer.GetFromCursor(actionBufferCursor)
			receiver := cli.Receiver(buffer.ReceiverIndex)
			if !receiver.Active() {
				continue
			}
			poppedActions := receiver.PopActions()

			// Transform input coordinates if camera component exists
			hasCam := client.Components.CameraIndex.CheckCursor(actionBufferCursor)
			if hasCam {
				camIndex := *client.Components.CameraIndex.GetFromCursor(actionBufferCursor)
				cam := cli.Cameras()[camIndex]
				if cam.Active() {
					globalPos, localPos := cam.Positions()
					// Convert global coordinates to local camera space
					for i, sAction := range poppedActions {
						localX := int(localPos.X + float64(sAction.X) - globalPos.X)
						localY := int(localPos.Y + float64(sAction.Y) - globalPos.Y)
						poppedActions[i].LocalX = localX
						poppedActions[i].LocalY = localY
					}
				}
			}
			buffer.AddBatch(poppedActions)
		}
	}
	return nil
}
