package input

// ClientActionMessage holds actions from a client connection for a specific receiver index.
// The server determines the target entity based on the connection sending this message.
type ClientActionMessage struct {
	// ReceiverIndex indicates which action-set this message corresponds to,
	// particularly if a single connection controls multiple inputs for its entity.
	ReceiverIndex int `json:"receiver_index"`

	// Actions contains the actual stamped action events recorded by the client.
	Actions []StampedAction `json:"actions"`
}
