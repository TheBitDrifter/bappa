package input

// nextInput stores the counter for generating sequential input identifiers
var nextInput Input = 0

// Input represents a unique identifier for an input event
type Input uint32

// StampedInput contains input data along with position and timing information
type StampedInput struct {
	Tick           int   // Tick when the input occurred
	Val            Input // The input identifier
	X, Y           int   // Screen coordinates where the input occurred
	LocalX, LocalY int   // Position relative to an entity's camera view
}

// NewInput generates a new unique Input identifier
func NewInput() Input {
	input := nextInput
	nextInput++
	return input
}
