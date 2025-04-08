package coldbrew_clientsystems

import (
	"encoding/json"
	"log"

	"github.com/TheBitDrifter/bappa/blueprint/input"
	"github.com/TheBitDrifter/bappa/coldbrew"
)

// InputSenderSystem collects inputs from local receivers and sends them over the network as StampedActions
type InputSenderSystem struct{}

func (s InputSenderSystem) Run(cli coldbrew.Client) error {
	networkCli, ok := cli.(coldbrew.NetworkClient)
	if !ok {
		log.Println("InputSenderSystem: Client is not a NetworkClient, cannot send actions")
		return nil
	}

	// Don't attempt to send if the client isn't connected.
	if !networkCli.IsConnected() {
		return nil
	}

	// Check each potential receiver slot.
	for i := 0; i < coldbrew.MaxSplit; i++ {
		receiver := cli.Receiver(i)
		// Skip receivers that are not marked as active.
		if !receiver.Active() {
			continue
		}

		// Retrieve and clear the action buffered for this receiver since the last frame.
		poppedActions := receiver.PopActions()
		// Skip if there were no actions for this receiver in this frame.
		if len(poppedActions) == 0 {
			continue
		}

		// Construct the message payload.
		message := input.ClientActionMessage{
			ReceiverIndex: i,
			Actions:       poppedActions,
		}

		// Serialize the message to JSON format.
		jsonData, err := json.Marshal(message)
		if err != nil {
			log.Printf("InputSenderSystem: Error marshalling action for receiver %d: %v", i, err)
			continue // Skip sending this message if marshalling fails.
		}

		// Send the serialized data over the network connection.
		err = networkCli.Send(jsonData)
		if err != nil {
			log.Printf("InputSenderSystem: Error sending action for receiver %d: %v", i, err)
		}

	}
	return nil
}
