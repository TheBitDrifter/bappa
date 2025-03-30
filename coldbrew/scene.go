package coldbrew

import (
	"errors"
	"sync/atomic"

	"github.com/TheBitDrifter/blueprint"
	"github.com/TheBitDrifter/table"
	"github.com/TheBitDrifter/warehouse"
)

var _ blueprint.Scene = &scene{} // Ensure scene implements blueprint.Scene

// Scene manages game scene state and systems
// It handles loading, storage, world dimensions, and local game systems
type Scene interface {
	// Core information
	Name() string
	Height() int
	Width() int
	CurrentTick() int
	LastSelectedTick() int
	SetSelectedTick()
	LastActivatedTick() int
	SetActivatedTick()
	TicksSinceActivated() int
	TicksSinceSelected() int
	// Storage and queries
	Storage() warehouse.Storage
	NewCursor(warehouse.QueryNode) *warehouse.Cursor
	// Loading state
	IsLoaded() bool
	SetLoaded(bool)
	IsLoading() bool
	SetLoading(bool)
	TryStartLoading() bool
	Ready() bool
	// Systems and execution
	CoreSystems() []blueprint.CoreSystem
	Renderers() []RenderSystem
	ClientSystems() []ClientSystem
	ExecutePlan() (alreadyExecuted bool, err error)
	Reset() error
}

type scene struct {
	name              string
	loaded            atomic.Bool       // Tracks if scene resources are loaded
	loading           atomic.Bool       // Tracks if scene is currently loading
	height, width     int               // Dimensions of the scene
	storage           warehouse.Storage // Entity component storage
	plan              blueprint.Plan    // Initialization plan
	planExecuted      bool              // Tracks if initialization plan has run
	lastSelectedTick  int               // Last tick when scene was selected
	lastActivatedTick int               // Last tick when scene was activated
	systems           struct {
		renderers []RenderSystem         // Systems for rendering
		client    []ClientSystem         // Systems for client-side logic
		core      []blueprint.CoreSystem // Core game systems
	}
}

// Core information methods
func (s *scene) CurrentTick() int         { return tick }
func (s *scene) LastSelectedTick() int    { return s.lastSelectedTick }
func (s *scene) SetSelectedTick()         { s.lastSelectedTick = tick }
func (s *scene) LastActivatedTick() int   { return s.lastActivatedTick }
func (s *scene) SetActivatedTick()        { s.lastActivatedTick = tick }
func (s *scene) TicksSinceSelected() int  { return tick - s.lastActivatedTick }
func (s *scene) TicksSinceActivated() int { return tick - s.lastActivatedTick }
func (s *scene) Height() int              { return s.height }
func (s *scene) Name() string             { return s.name }
func (s *scene) Width() int               { return s.width }

// Storage and query methods
func (s *scene) NewCursor(query warehouse.QueryNode) *warehouse.Cursor {
	return warehouse.Factory.NewCursor(query, s.storage)
}
func (s *scene) Storage() warehouse.Storage { return s.storage }

// Loading state methods
func (s *scene) IsLoaded() bool        { return s.loaded.Load() }
func (s *scene) IsLoading() bool       { return s.loading.Load() }
func (s *scene) SetLoaded(value bool)  { s.loaded.Store(value) }
func (s *scene) SetLoading(value bool) { s.loading.Store(value) }
func (s *scene) TryStartLoading() bool { return s.loading.CompareAndSwap(false, true) }
func (s *scene) Ready() bool {
	return !s.IsLoading() && s.IsLoaded()
}

// Systems and execution methods
func (s *scene) ExecutePlan() (bool, error) {
	if !s.planExecuted {
		err := s.plan(s.height, s.width, s.storage)
		if err != nil {
			return false, err
		}
		s.planExecuted = true
		return false, nil
	}
	return true, nil
}
func (s *scene) Renderers() []RenderSystem           { return s.systems.renderers }
func (s *scene) CoreSystems() []blueprint.CoreSystem { return s.systems.core }
func (s *scene) ClientSystems() []ClientSystem       { return s.systems.client }

// Reset resets the scene by creating new storage and clearing execution state
func (s *scene) Reset() error {
	schema := table.Factory.NewSchema()
	if s.storage.Locked() {
		return errors.New("storage is locked")
	}
	s.storage = warehouse.Factory.NewStorage(schema)
	s.planExecuted = false
	return nil
}
