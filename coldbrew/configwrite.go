package coldbrew

import (
	"github.com/hajimehoshi/ebiten/v2"
)

// ConfigManager defines the interface for managing game configuration settings
type ConfigManager interface {
	SetTitle(string)
	SetWindowSize(x, y int)
	SetResolution(x, y int)
	SetResizable(bool)
	SetFullScreen(bool)
	SetMinimumLoadTime(ticks int)
	SetEnforceMinOnActive(bool)
	SetTPS(int)
	BindDebugKey(ebiten.Key)
	SetDebugMode(bool)
	SetCameraBorderSize(int)
}

// configWrite contains configuration parameters that can be written to
type configWrite struct {
	Title       string
	DebugVisual bool
}

// Ensure client implements ConfigManager
var _ ConfigManager = &client{}

// configManager implements the ConfigManager interface
type configManager struct{}

// newConfigManager creates a new configuration manager
func newConfigManager() *configManager {
	return &configManager{}
}

// SetTitle updates the game window title
func (cm *configManager) SetTitle(title string) {
	ebiten.SetWindowTitle(title)
	ClientConfig.Title = title
}

// SetWindowSize updates the window dimensions and applies them
func (cm *configManager) SetWindowSize(x, y int) {
	ClientConfig.windowSize.x = x
	ClientConfig.windowSize.y = y
	ebiten.SetWindowSize(x, y)
}

// SetResolution updates the internal game resolution
func (cm *configManager) SetResolution(x, y int) {
	ClientConfig.resolution.x = x
	ClientConfig.resolution.y = y
}

// SetResizable controls whether the game window can be resized
func (cm *configManager) SetResizable(isResizable bool) {
	if isResizable {
		ebiten.SetWindowResizingMode(ebiten.WindowResizingModeEnabled)
		return
	}
	ebiten.SetWindowResizingMode(ebiten.WindowResizingModeDisabled)
}

// SetFullScreen toggles fullscreen mode and updates stored window dimensions
func (cm *configManager) SetFullScreen(fullscreen bool) {
	ebiten.SetFullscreen(fullscreen)
	width, height := ebiten.WindowSize()
	ClientConfig.windowSize.x = width
	ClientConfig.windowSize.y = height
}

// SetMinimumLoadTime sets the minimum loading time in ticks
func (cm *configManager) SetMinimumLoadTime(ticks int) {
	ClientConfig.minimumLoadTime = ticks
}

// SetEnforceMinOnActive controls whether minimum load time is enforced when active
func (cm *configManager) SetEnforceMinOnActive(val bool) {
	ClientConfig.enforceMinOnActive = val
}

// SetTPS updates the game's ticks per second
func (cm *configManager) SetTPS(ticks int) {
	ebiten.SetTPS(ticks)
	ClientConfig.tps = ticks
}

// BindDebugKey sets the key used to trigger debug functionality
func (cm *configManager) BindDebugKey(key ebiten.Key) {
	ClientConfig.debugKey = key
}

// BindDebugKey sets the key used to trigger debug functionality
func (cm *configManager) SetDebugMode(val bool) {
	ClientConfig.DebugVisual = val
}

// BindDebugKey sets the key used to trigger debug functionality
func (cm *configManager) SetCameraBorderSize(thickness int) {
	ClientConfig.cameraBorderSize = thickness
}
