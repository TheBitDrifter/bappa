// Package coldbrew provides input management functionality for various input devices.
package coldbrew

import (
	"errors"

	"github.com/TheBitDrifter/bark"
)

// InputManager defines the interface for managing input receivers.
type InputManager interface {
	ActivateReceiver() (Receiver, error)
	Receiver(index int) Receiver
}

// inputManager implements the InputManager interface with support for
// multiple input devices and receivers.
type inputManager struct {
	receivers [MaxSplit]*receiver
	*capturers
}

// capturers holds input capture mechanisms for different device types.
type capturers struct {
	keyboard InputCapturer
	touch    InputCapturer
	mouse    InputCapturer
	gamepad  InputCapturer
}

// newInputManager creates and initializes a new input manager for the given client.
func newInputManager(cli *client) *inputManager {
	m := &inputManager{
		capturers: &capturers{
			keyboard: newKeyboardCapturer(cli),
			mouse:    newMouseCapturer(cli),
			gamepad:  newGamepadCapturer(cli),
			touch:    newTouchCapturer(cli),
		},
	}
	for i := range m.receivers {
		m.receivers[i] = &receiver{}
	}
	return m
}

// ActivateReceiver finds an inactive receiver and activates it with initialized layouts.
// Returns an error if no receiver is available.
func (m *inputManager) ActivateReceiver() (Receiver, error) {
	for i := range m.receivers {
		if !m.receivers[i].active {
			m.receivers[i].active = true
			if m.receivers[i].keyLayout == nil {
				m.receivers[i].keyLayout = &keyLayout{}
			}
			if m.receivers[i].mouseLayout == nil {
				m.receivers[i].mouseLayout = &mouseLayout{}
			}
			if m.receivers[i].padLayout == nil {
				m.receivers[i].padLayout = &padLayout{}
			}
			if m.receivers[i].touchLayout == nil {
				m.receivers[i].touchLayout = &touchLayout{}
			}
			m.receivers[i].padID = -1
			return m.receivers[i], nil
		}
	}
	return nil, bark.AddTrace(errors.New("no available receiver"))
}

// Receiver returns the receiver with the specified ID.
func (m *inputManager) Receiver(id int) Receiver {
	return m.receivers[id]
}
