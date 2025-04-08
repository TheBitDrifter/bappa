package drip_seversystems

import (
	"log"

	"github.com/TheBitDrifter/bappa/blueprint/input"
	"github.com/TheBitDrifter/bappa/drip"
	"github.com/TheBitDrifter/bappa/warehouse"
)

// ActionBufferSystem injects network received actions into the core sim action buffers
type ActionBufferSystem struct{}

func (ActionBufferSystem) Run(s drip.Server) error {
	actionsToProcess := s.ConsumeAllActions()

	if len(actionsToProcess) > 0 {

		activeScenesCopy := s.ActiveScenes()

		// Iterate through the actions that were received since the last tick.
		for _, item := range actionsToProcess {
			var targetEntity warehouse.Entity = nil
			var found bool = false

			// Find the target entity within the currently active scenes.
			// We're currently only supporting one scene, but a loop cant can't hurt for now ¯\_(ツ)_/¯
			for _, scene := range activeScenesCopy {
				// Attempt to get the entity from the scene's storage.
				potentialEntity, err := scene.Storage().Entity(int(item.TargetEntityID))
				// Check if found, valid, and the recycled count matches (prevents using stale IDs).
				if err == nil && potentialEntity.Valid() && potentialEntity.Recycled() == item.Recycled {
					targetEntity = potentialEntity
					found = true
					break
				}
			}

			// If the entity wasn't found in any active scene, log and skip.
			if !found {
				log.Printf("Update: Could not find valid entity ID %d (recycled %d) in active scenes.", item.TargetEntityID, item.Recycled)
				continue
			}

			actionBuffer := input.Components.ActionBuffer.GetFromEntity(targetEntity)
			if actionBuffer == nil {
				log.Printf("Update: Entity ID %d missing InputBuffer component.", item.TargetEntityID)
				continue
			}

			if actionBuffer.ReceiverIndex != item.ReceiverIndex {
				log.Printf("Update: Mismatched ReceiverIndex for entity %d. Input msg: %d, Component: %d. Discarding.",
					item.TargetEntityID, item.ReceiverIndex, actionBuffer.ReceiverIndex)
				continue
			}

			// Add the received actions to the entity's standard ActionBuffer.
			actionBuffer.AddBatch(item.Actions)
		}
	}
	return nil
}
