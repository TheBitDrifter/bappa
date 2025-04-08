package input

// nextAction stores the counter for generating sequential action identifiers
var nextAction Action = 0

// Action represents a unique identifier for an action event
type Action uint32

// StampedAction contains input data along with position and timing information
type StampedAction struct {
	Tick           int    // Tick when the input occurred
	Val            Action // The action identifier
	X, Y           int    // Screen coordinates where the action occurred
	LocalX, LocalY int    // Position relative to an entity's camera view
}

// NewAction generates a new unique Input identifier
func NewAction() Action {
	action := nextAction
	nextAction++
	return action
}
