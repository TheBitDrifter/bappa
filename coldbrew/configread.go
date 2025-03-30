package coldbrew

import (
	"sync/atomic"

	blueprintclient "github.com/TheBitDrifter/blueprint/client"
	"github.com/hajimehoshi/ebiten/v2"
)

// MaxSplit is the maximum number of splits allowed by the blueprint client
const MaxSplit = blueprintclient.MaxSplit

// ClientConfig contains default configuration settings for the game client

var ClientConfig = func() *config {
	defConfig := config{
		configWrite: configWrite{
			Title:       "Bappa!",
			DebugVisual: true,
		},
		minimumLoadTime:    0,
		enforceMinOnActive: true,
		tps:                60,
		debugKey:           ebiten.Key0,
		cameraBorderSize:   0,
		resolution: struct{ x, y int }{
			x: 640,
			y: 360,
		},
		windowSize: struct{ x, y int }{
			x: 1920,
			y: 1080,
		},
	}

	defConfig.maxSpritesCached.Store(200)
	defConfig.maxSoundsCached.Store(200)
	return &defConfig
}()

// config holds all configuration parameters for the game client
type config struct {
	configWrite
	maxSpritesCached   atomic.Uint32
	maxSoundsCached    atomic.Uint32
	minimumLoadTime    int
	enforceMinOnActive bool
	tps                int
	debugKey           ebiten.Key
	cameraBorderSize   int
	resolution         struct {
		x, y int
	}
	baseResolution struct {
		x, y int
	}
	windowSize struct {
		x, y int
	}
}

// BaseResolution returns the base resolution x and y values
func (c config) BaseResolution() (x, y int) { return c.baseResolution.x, c.baseResolution.y }

// Resolution returns the current resolution x and y values
func (c config) Resolution() (x, y int) { return c.resolution.x, c.resolution.y }

// WindowSize returns the window size x and y values
func (c config) WindowSize() (x, y int) { return c.windowSize.x, c.windowSize.y }

// MinimumLoadTime returns the minimum time to show loader when transitioning scenes
func (c config) MinimumLoadTime() int { return c.minimumLoadTime }

// EnforceMinOnActive returns whether to enforce minimum load time rules when transitioning to active scenes
func (c config) EnforceMinOnActive() bool { return c.enforceMinOnActive }

// TPS returns the ticks per second for the game loop
func (c config) TPS() int { return c.tps }

// DebugKey returns the key bound to trigger debug functionality
func (c config) DebugKey() ebiten.Key { return c.debugKey }

// DebugKey returns the key bound to trigger debug functionality
func (c config) CameraBorderSize() int { return c.cameraBorderSize }
