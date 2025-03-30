package coldbrew

import (
	"log/slog"
	"math"

	"github.com/TheBitDrifter/bark"
	blueprintinput "github.com/TheBitDrifter/blueprint/input"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
)

var _ InputCapturer = &gamepadCapturer{}

// gamepadCapturer handles gamepad input detection and processing
type gamepadCapturer struct {
	client  *client
	logger  *slog.Logger
	buttons map[ebiten.GamepadID][]ebiten.GamepadButton
	axes    map[ebiten.GamepadID][]float64
	sticks  map[ebiten.GamepadID]stickState
	active  bool

	// Map logical padIDs (what users register with RegisterPad) to physical GamepadIDs
	padMapping         map[int]ebiten.GamepadID
	mappingNeedsUpdate bool
}

// stick represents an analog stick's position
type stick struct {
	X float64
	Y float64
}

// stickState tracks analog stick positions for a gamepad
type stickState struct {
	Left  stick
	Right stick
}

// newGamepadCapturer creates a new gamepad input capturer
func newGamepadCapturer(client *client) *gamepadCapturer {
	return &gamepadCapturer{
		client:             client,
		logger:             bark.For("gamepad"),
		buttons:            make(map[ebiten.GamepadID][]ebiten.GamepadButton),
		axes:               make(map[ebiten.GamepadID][]float64),
		sticks:             make(map[ebiten.GamepadID]stickState),
		padMapping:         make(map[int]ebiten.GamepadID),
		mappingNeedsUpdate: true,
	}
}

// Capture processes all gamepad inputs
func (h *gamepadCapturer) Capture() {
	h.handleDisconnections()
	h.handleNewConnections()

	// Update the mapping between logical and physical controller IDs when needed
	if h.mappingNeedsUpdate {
		h.updateControllerMapping()
		h.mappingNeedsUpdate = false
	}

	h.captureInputs()
	h.processReceiverInputs()
}

// updateControllerMapping maps logical padIDs to physical GamepadIDs
func (h *gamepadCapturer) updateControllerMapping() {
	// Get all available gamepads by checking all possible IDs
	// Ebiten doesn't provide a direct GamepadCount function
	availableGamepads := make([]ebiten.GamepadID, 0)

	// Check the first 16 possible gamepad IDs
	// This is a reasonable limit that should cover most use cases
	for i := 0; i < 16; i++ {
		id := ebiten.GamepadID(i)
		// Check if this ID is in our tracked buttons map, which means it's connected
		if _, exists := h.buttons[id]; exists {
			availableGamepads = append(availableGamepads, id)
		}
	}

	h.logger.Debug("updating gamepad mapping",
		"available_gamepads", len(availableGamepads))

	// Clear the mapping
	h.padMapping = make(map[int]ebiten.GamepadID)

	// Map each receiver's padID to an available physical gamepad
	for i, receiver := range h.client.receivers {
		if !receiver.active {
			continue
		}

		logicalID := receiver.padID

		// If this logical ID is in range of available gamepads, map it
		if logicalID >= 0 && logicalID < len(availableGamepads) {
			physicalID := availableGamepads[logicalID]
			h.padMapping[logicalID] = physicalID

			h.logger.Debug("mapped logical pad to physical gamepad",
				"receiver", i,
				"logical_id", logicalID,
				"physical_id", physicalID,
				"gamepad_name", ebiten.GamepadName(physicalID))
		} else {
			h.logger.Warn("no physical gamepad available for logical ID",
				"receiver", i,
				"logical_id", logicalID,
				"available_count", len(availableGamepads))
		}
	}
}

// handleDisconnections cleans up state for disconnected gamepads
func (h *gamepadCapturer) handleDisconnections() {
	disconnected := false
	for id := range h.buttons {
		if inpututil.IsGamepadJustDisconnected(id) {
			h.logger.Info("gamepad disconnected",
				"id", id,
				"name", ebiten.GamepadName(id),
				"sdl_id", ebiten.GamepadSDLID(id),
			)
			delete(h.buttons, id)
			delete(h.axes, id)
			delete(h.sticks, id)
			disconnected = true
		}
	}

	if disconnected {
		h.mappingNeedsUpdate = true
	}
}

// handleNewConnections initializes state for newly connected gamepads
func (h *gamepadCapturer) handleNewConnections() {
	newIds := inpututil.AppendJustConnectedGamepadIDs([]ebiten.GamepadID{})
	if len(newIds) > 0 {
		h.mappingNeedsUpdate = true
	}

	for _, id := range newIds {
		h.buttons[id] = []ebiten.GamepadButton{}
		h.axes[id] = []float64{}
		h.sticks[id] = stickState{}
		h.logger.Info("gamepad connected",
			"id", id,
			"name", ebiten.GamepadName(id),
			"sdl_id", ebiten.GamepadSDLID(id),
		)
	}

	// Instead of using GamepadCount, we'll check for gamepads that are connected
	// but not yet in our tracking maps
	// Check the first 16 possible gamepad IDs
	for i := 0; i < MaxSplit; i++ {
		id := ebiten.GamepadID(i)
		// A connected gamepad should have a non-empty name
		if ebiten.GamepadName(id) != "" {
			if _, exists := h.buttons[id]; !exists {
				h.buttons[id] = []ebiten.GamepadButton{}
				h.axes[id] = []float64{}
				h.sticks[id] = stickState{}
				h.logger.Debug("tracking previously connected gamepad",
					"id", id,
					"name", ebiten.GamepadName(id),
				)
				h.mappingNeedsUpdate = true
			}
		}
	}
}

// captureInputs reads the current state of all connected gamepads
func (h *gamepadCapturer) captureInputs() {
	for id := range h.buttons {
		h.captureButtonState(id)
		h.captureAxesState(id)
	}
}

// captureButtonState records which buttons are currently pressed
func (h *gamepadCapturer) captureButtonState(id ebiten.GamepadID) {
	var pressedButtons []ebiten.GamepadButton
	pressedButtons = inpututil.AppendPressedGamepadButtons(id, pressedButtons)
	h.buttons[id] = pressedButtons
}

// captureAxesState captures analog input state, preferring standard gamepad layout
// when available, falling back to raw axis values
func (h *gamepadCapturer) captureAxesState(id ebiten.GamepadID) {
	if ebiten.IsStandardGamepadLayoutAvailable(id) {
		h.sticks[id] = stickState{
			Left: processStickInput(
				ebiten.StandardGamepadAxisValue(id, ebiten.StandardGamepadAxisLeftStickHorizontal),
				ebiten.StandardGamepadAxisValue(id, ebiten.StandardGamepadAxisLeftStickVertical),
			),
			Right: processStickInput(
				ebiten.StandardGamepadAxisValue(id, ebiten.StandardGamepadAxisRightStickHorizontal),
				ebiten.StandardGamepadAxisValue(id, ebiten.StandardGamepadAxisRightStickVertical),
			),
		}
		return
	}

	// Fallback to raw axes if standard layout isn't available
	maxAxis := ebiten.GamepadAxisCount(id)
	axisValues := make([]float64, maxAxis)

	for a := 0; a < maxAxis; a++ {
		axisValues[a] = ebiten.GamepadAxisValue(id, ebiten.GamepadAxisType(a))
	}
	h.axes[id] = axisValues

	if len(axisValues) >= 4 {
		h.sticks[id] = stickState{
			Left:  processStickInput(axisValues[0], axisValues[1]),
			Right: processStickInput(axisValues[2], axisValues[3]),
		}
	} else {
		h.logger.Debug("insufficient axes for stick mapping",
			"gamepad_id", id,
			"axes_count", len(axisValues),
		)
	}
}

// processReceiverInputs forwards gamepad inputs to active receivers
func (h *gamepadCapturer) processReceiverInputs() {
	x, y := ebiten.CursorPosition()

	for i, receiver := range h.client.receivers {
		if !receiver.active {
			continue
		}

		// Get the physical gamepad ID for this receiver's logical padID
		physicalID, exists := h.padMapping[receiver.padID]
		if !exists {
			// Skip this receiver if we don't have a physical gamepad for it
			continue
		}

		h.processButtonInputsForGamepad(i, receiver, physicalID, x, y)
		h.processStickInputsForGamepad(i, receiver, physicalID, x, y)
	}
}

// processButtonInputsForGamepad handles button presses for a specific receiver with explicit gamepad ID
func (h *gamepadCapturer) processButtonInputsForGamepad(receiverIndex int, receiver *receiver, gamepadID ebiten.GamepadID, x, y int) {
	buttons := h.buttons[gamepadID]
	for _, btn := range buttons {
		if !receiver.padLayout.mask.Contains(uint32(btn)) {
			continue
		}

		// Make sure button index is within the slice bounds
		if int(btn) >= len(receiver.padLayout.buttons) {
			h.logger.Debug("button index out of range",
				"button", btn,
				"buttons_length", len(receiver.padLayout.buttons))
			continue
		}

		input := receiver.padLayout.buttons[btn]
		h.client.receivers[receiverIndex].inputs.pad = append(
			h.client.receivers[receiverIndex].inputs.pad,
			blueprintinput.StampedInput{
				Tick: tick,
				X:    x,
				Y:    y,
				Val:  input,
			},
		)

		h.logger.Debug("button input processed",
			"receiver", receiverIndex,
			"gamepad_id", gamepadID,
			"button", btn)
	}
}

// processStickInputsForGamepad handles analog stick movement for a specific receiver with explicit gamepad ID
func (h *gamepadCapturer) processStickInputsForGamepad(receiverIndex int, receiver *receiver, gamepadID ebiten.GamepadID, x, y int) {
	stickState, ok := h.sticks[gamepadID]
	if !ok {
		return
	}

	// Process left stick if enabled
	if (stickState.Left.X != 0 || stickState.Left.Y != 0) && h.client.receivers[receiverIndex].leftAxes {
		h.client.receivers[receiverIndex].inputs.pad = append(
			h.client.receivers[receiverIndex].inputs.pad,
			blueprintinput.StampedInput{
				Tick: tick,
				// Use stick X/Y values directly as the vector components
				X:   int(stickState.Left.X * 100),  // Scale to a reasonable range if needed
				Y:   int(-stickState.Left.Y * 100), // Invert Y and scale to a reasonable range
				Val: h.client.receivers[receiverIndex].leftAxesInput,
			},
		)
		h.logger.Debug("gamepad left stick processed",
			"receiver", receiverIndex,
			"gamepad_id", gamepadID,
			"x", stickState.Left.X,
			"y", -stickState.Left.Y, // Showing inverted Y value in logs
			"val", h.client.receivers[receiverIndex].leftAxesInput,
		)
	}

	// Process right stick if enabled
	if (stickState.Right.X != 0 || stickState.Right.Y != 0) && h.client.receivers[receiverIndex].rightAxes {
		h.client.receivers[receiverIndex].inputs.pad = append(
			h.client.receivers[receiverIndex].inputs.pad,
			blueprintinput.StampedInput{
				Tick: tick,
				// Use stick X/Y values directly as the vector components
				X:   int(stickState.Right.X * 100),  // Scale to a reasonable range if needed
				Y:   int(-stickState.Right.Y * 100), // Invert Y and scale to a reasonable range
				Val: h.client.receivers[receiverIndex].rightAxesInput,
			},
		)
		h.logger.Debug("gamepad right stick processed",
			"receiver", receiverIndex,
			"gamepad_id", gamepadID,
			"x", stickState.Right.X,
			"y", -stickState.Right.Y, // Showing inverted Y value in logs
			"val", h.client.receivers[receiverIndex].rightAxesInput,
		)
	}
}

// The original processButtonInputs and processStickInputs are kept for backward compatibility
// but they delegate to the new implementations that take explicit gamepad IDs

// processButtonInputs handles button presses for a specific receiver
func (h *gamepadCapturer) processButtonInputs(receiverIndex int, receiver *receiver, x, y int) {
	physicalID, exists := h.padMapping[receiver.padID]
	if !exists {
		// Use the old way as fallback
		physicalID = ebiten.GamepadID(receiver.padID)
	}
	h.processButtonInputsForGamepad(receiverIndex, receiver, physicalID, x, y)
}

// processStickInputs handles analog stick movement for a specific receiver
func (h *gamepadCapturer) processStickInputs(receiverIndex int, receiver *receiver, x, y int) {
	physicalID, exists := h.padMapping[receiver.padID]
	if !exists {
		// Use the old way as fallback
		physicalID = ebiten.GamepadID(receiver.padID)
	}
	h.processStickInputsForGamepad(receiverIndex, receiver, physicalID, x, y)
}

// processStickInput processes raw stick input with deadzone and normalization
// Returns normalized stick position after deadzone processing
func processStickInput(x, y float64) stick {
	const deadzone = 0.25
	x = math.Max(math.Min(x, 1.0), -1.0)
	y = math.Max(math.Min(y, 1.0), -1.0)

	magnitude := math.Sqrt(x*x + y*y)
	if magnitude < deadzone || magnitude == 0 {
		return stick{X: 0, Y: 0}
	}

	normalizedX := x / magnitude
	normalizedY := y / magnitude

	adjustedMagnitude := (magnitude - deadzone) / (1 - deadzone)
	adjustedMagnitude = math.Max(0, math.Min(1, adjustedMagnitude))
	return stick{
		X: normalizedX * adjustedMagnitude,
		Y: normalizedY * adjustedMagnitude,
	}
}
