package coldbrew

import "github.com/hajimehoshi/ebiten/v2"

// TickManager provides game tick counting functionality and implements ebiten.Game
type TickManager interface {
	CurrentTick() int
	ebiten.Game
}

type tickManager struct{}

// newTickManager creates a new tick manager instance
func newTickManager() *tickManager {
	return &tickManager{}
}

// CurrentTick returns the current game tick count
func (g *tickManager) CurrentTick() int {
	return tick
}

// Update increments the tick counter
// Implements part of the ebiten.Game interface
func (g *tickManager) Update() error {
	tick++
	return nil
}
