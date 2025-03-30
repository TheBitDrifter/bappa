package clientsystems

import (
	"github.com/TheBitDrifter/coldbrew"
)

// SplitScreenLayoutSystem manages camera layouts based on resolution, window size and active camera count
type SplitScreenLayoutSystem struct {
	lastCount int
}

// Internal layout configuration for organizing camera views
type layout struct {
	rows, cols int
	isVertical bool
}

// Run processes active cameras and updates layout when count changes
func (s *SplitScreenLayoutSystem) Run(cli coldbrew.Client) error {
	cameras := cli.Cameras()
	count := 0
	for _, c := range cameras {
		if c.Active() {
			count++
		}
	}
	if count != s.lastCount {
		s.lastCount = count
		s.updateLayout(count, cli)
	}
	return nil
}

// updateLayout recalculates and applies new screen layout based on active camera count
func (s *SplitScreenLayoutSystem) updateLayout(count int, cli coldbrew.Client) {
	resX, resY := coldbrew.ClientConfig.Resolution()
	windowX, windowY := coldbrew.ClientConfig.WindowSize()
	maxHorizontal := windowX / resX
	maxVertical := windowY / resY
	maxScreens := maxHorizontal * maxVertical
	layout := s.calculateLayout(count)
	if count <= maxScreens {
		s.applyResolutionLayout(layout, resX, resY, cli)
		return
	}
	panic("unable to apply split screen layout given window size and base resolution")
}

// calculateLayout determines optimal grid configuration based on active camera count
func (SplitScreenLayoutSystem) calculateLayout(count int) layout {
	switch count {
	case 1:
		return layout{rows: 1, cols: 1, isVertical: false}
	case 2:
		return layout{rows: 2, cols: 1, isVertical: true}
	case 3, 4:
		return layout{rows: 2, cols: 2, isVertical: false}
	case 5, 6:
		return layout{rows: 2, cols: 3, isVertical: false}
	case 7, 8:
		return layout{rows: 2, cols: 4, isVertical: false}
	default:
		return layout{rows: 1, cols: 1, isVertical: false}
	}
}

// applyResolutionLayout sets camera dimensions and positions based on calculated layout
func (s *SplitScreenLayoutSystem) applyResolutionLayout(l layout, resX, resY int, cli coldbrew.Client) {
	cli.SetResolution(resX*l.cols, resY*l.rows)
	position := 0
	for _, cam := range cli.Cameras() {
		if !cam.Active() {
			continue
		}
		row := position / l.cols
		col := position % l.cols
		cam.SetDimensions(resX, resY)
		screenPos, _ := cam.Positions()
		screenPos.X = float64(col * resX)
		screenPos.Y = float64(row * resY)
		position++
	}
}
