package drip

import (
	"sync"

	"github.com/TheBitDrifter/bappa/blueprint"
	"github.com/TheBitDrifter/bappa/warehouse"
)

type Scene interface {
	blueprint.Scene
	Name() string
	CoreSystems() []blueprint.CoreSystem
	IncrementTick()
}

type sceneImpl struct {
	name         string
	width        int
	height       int
	plan         blueprint.Plan
	coreSystems  []blueprint.CoreSystem
	storage      warehouse.Storage
	planExecuted bool

	sceneTick   int
	sceneMutex  sync.RWMutex
	serverMutex *sync.RWMutex
}

func (s *sceneImpl) Width() int {
	return s.width
}

func (s *sceneImpl) Height() int {
	return s.height
}

func (s *sceneImpl) Storage() warehouse.Storage {
	return s.storage
}

func (s *sceneImpl) NewCursor(queryNode warehouse.QueryNode) *warehouse.Cursor {
	return warehouse.Factory.NewCursor(queryNode, s.storage)
}

func (s *sceneImpl) CurrentTick() int {
	s.sceneMutex.RLock()
	defer s.sceneMutex.RUnlock()
	return s.sceneTick
}

// IncrementTick updates the tick safely
func (s *sceneImpl) IncrementTick() {
	s.sceneMutex.Lock()
	s.sceneTick++
	s.sceneMutex.Unlock()
}

// Name returns the scene's name
func (s *sceneImpl) Name() string {
	return s.name
}

// CoreSystems returns all core systems for this scene
func (s *sceneImpl) CoreSystems() []blueprint.CoreSystem {
	return s.coreSystems
}
