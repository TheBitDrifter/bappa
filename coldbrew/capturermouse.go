package coldbrew

import (
	"log/slog"

	"github.com/TheBitDrifter/bappa/blueprint/input"
	"github.com/TheBitDrifter/bark"
	"github.com/hajimehoshi/ebiten/v2"
)

// mouseCapturer handles mouse input detection and processing
type mouseCapturer struct {
	client *clientImpl
	logger *slog.Logger
}

func newMouseCapturer(client *clientImpl) *mouseCapturer {
	return &mouseCapturer{
		client: client,
		logger: bark.For("mouse"),
	}
}

func (handler *mouseCapturer) Capture() {
	client := handler.client
	for i := range client.receivers {
		if err := handler.populateReceiver(client.receivers[i]); err != nil {
			handler.logger.Error("failed to populate receiver",
				bark.KeyError, err,
				"receiver_index", i)
		}
	}
}

func (handler *mouseCapturer) populateReceiver(receiverPtr *receiver) error {
	if !receiverPtr.active {
		return nil
	}
	x, y := ebiten.CursorPosition()
	pressedButtons := make([]ebiten.MouseButton, 0)
	for _, eMouseBtn := range receiverPtr.mouseLayout.mouseButtonsRaw {
		if ebiten.IsMouseButtonPressed(eMouseBtn) {
			pressedButtons = append(pressedButtons, eMouseBtn)
			receiverPtr.actions.mouse = append(receiverPtr.actions.mouse, input.StampedAction{
				Val:  receiverPtr.mouseLayout.mouseButtons[eMouseBtn],
				Tick: tick,
				X:    x,
				Y:    y,
			})
			handler.logger.Debug("mouse buttons pressed",
				"buttons", pressedButtons,
				"x", x,
				"y", y,
				"tick", tick,
			)

		}
	}
	return nil
}
