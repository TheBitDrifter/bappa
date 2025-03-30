package coldbrew

import (
	"errors"
	"iter"
	"sync"

	"github.com/TheBitDrifter/bark"
	"github.com/TheBitDrifter/blueprint"
	blueprintclient "github.com/TheBitDrifter/blueprint/client"
	"github.com/TheBitDrifter/table"
	"github.com/TheBitDrifter/warehouse"
)

var _ SceneManager = &sceneManager{}

// SceneManager handles scene lifecycle, transitions, and state management
type SceneManager interface {
	// Scene access

	// IsActive checks if the given scene is currently active
	IsActive(Scene) bool
	// LoadingScenes returns all scenes currently being loaded
	LoadingScenes() []Scene
	// Cache returns the scene cache
	Cache() warehouse.Cache[Scene]
	// Scene lifecycle

	ActiveScenes() iter.Seq[Scene]
	ActiveScene(int) Scene
	SceneCount() int
	// RegisterScene creates and registers a new scene with the provided configuration
	RegisterScene(string, int, int, blueprint.Plan, []RenderSystem, []ClientSystem, []blueprint.CoreSystem) error
	// ChangeScene transitions to a target scene, transferring specified entities
	ChangeScene(target Scene, entities ...warehouse.Entity) error
	// ActivateScene activates a target scene while keeping the origin scene active
	ActivateScene(target Scene, entities ...warehouse.Entity) error
	// DeactivateScene removes a scene from the active scenes list
	DeactivateScene(target Scene)
}

type sceneManager struct {
	mu            sync.RWMutex
	activeScenes  []Scene
	loadingScenes []Scene
	cache         warehouse.Cache[Scene]
}

// newSceneManager creates a scene manager with specified cache size
func newSceneManager(maxScenesCached int) *sceneManager {
	return &sceneManager{
		cache: warehouse.FactoryNewCache[Scene](maxScenesCached),
	}
}

func (m *sceneManager) ActiveScenes() iter.Seq[Scene] {
	return func(yield func(Scene) bool) {
		m.mu.RLock()
		defer m.mu.RUnlock()

		for _, s := range m.activeScenes {
			if !yield(s) {
				return
			}
		}
	}
}

// Scene access methods
func (m *sceneManager) IsActive(check Scene) bool {
	m.mu.RLock()
	defer m.mu.RUnlock()

	for _, match := range m.activeScenes {
		if check == match {
			return true
		}
	}
	return false
}

func (m *sceneManager) LoadingScenes() []Scene {
	m.mu.RLock()
	defer m.mu.RUnlock()

	// Create a copy to avoid race conditions on the returned slice
	result := make([]Scene, len(m.loadingScenes))
	copy(result, m.loadingScenes)
	return result
}

func (m *sceneManager) Cache() warehouse.Cache[Scene] {
	m.mu.RLock()
	defer m.mu.RUnlock()

	return m.cache
}

// ActivateScene adds a target scene to active scenes and transfers specified entities from origin
func (m *sceneManager) ActivateScene(target Scene, entities ...warehouse.Entity) error {
	targetStorage := target.Storage()
	for _, en := range entities {
		originStorage := en.Storage()
		if err := originStorage.TransferEntities(targetStorage, en); err != nil {
			return bark.AddTrace(err)
		}
	}

	for scene := range m.ActiveScenes() {
		if scene == target {
			target.SetSelectedTick()
			return nil
		}
	}
	target.SetActivatedTick()
	target.SetSelectedTick()
	m.activeScenes = append(m.activeScenes, target)
	return nil
}

// ChangeScene replaces the current active scene with the target scene
// Only works when exactly one scene is active
func (m *sceneManager) ChangeScene(target Scene, entities ...warehouse.Entity) error {
	if len(m.activeScenes) > 1 {
		return bark.AddTrace(
			errors.New("cannot use change scene api when multiple scenes are active — use activate scene api instead"),
		)
	}
	if len(m.activeScenes) == 0 {
		return bark.AddTrace(errors.New("no scenes are active"))
	}
	origin := m.activeScenes[0]
	originStorage := origin.Storage()
	targetStorage := target.Storage()
	for _, en := range entities {
		if err := originStorage.TransferEntities(targetStorage, en); err != nil {
			return bark.AddTrace(err)
		}
	}
	target.SetSelectedTick()
	m.activeScenes[0] = target
	return nil
}

// DeactivateScene removes the target scene from the active scenes list
func (m *sceneManager) DeactivateScene(target Scene) {
	i := 0
	for scene := range m.ActiveScenes() {
		if scene == target {
			lastIdx := len(m.activeScenes) - 1
			m.activeScenes[i] = m.activeScenes[lastIdx]
			m.activeScenes = m.activeScenes[:lastIdx]

			// When the scene is deactivated it may have its assets cleared for a cache resolution
			// Thus the atomic indexes are no longer safe
			// We must fall back on keys
			cursor := scene.NewCursor(blueprint.Queries.SpriteBundle)
			for range cursor.Next() {
				bundle := blueprintclient.Components.SpriteBundle.GetFromCursor(cursor)
				for i := range bundle.Blueprints {
					bundle.Blueprints[i].Location.Index.Store(0)
				}
			}
			cursor = scene.NewCursor(blueprint.Queries.SoundBundle)
			for range cursor.Next() {
				bundle := blueprintclient.Components.SoundBundle.GetFromCursor(cursor)
				for i := range bundle.Blueprints {
					bundle.Blueprints[i].Location.Index.Store(0)
				}
			}

			target.SetLoaded(false)
			break
		}

		i++
	}
}

// RegisterScene creates and registers a new scene with the provided configuration
// If no scenes are active, the new scene becomes active
func (m *sceneManager) RegisterScene(
	name string,
	width, height int,
	plan blueprint.Plan,
	renderSystems []RenderSystem,
	clientSystems []ClientSystem,
	coreSystems []blueprint.CoreSystem,
) error {
	newScene, err := m.newScene(name, width, height, plan, renderSystems, clientSystems, coreSystems)
	if err != nil {
		return bark.AddTrace(err)
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	m.cache.Register(name, newScene)
	if len(m.activeScenes) == 0 {
		m.activeScenes = append(m.activeScenes, newScene)
	}
	return nil
}

// newScene creates a new scene with the provided configuration
func (m *sceneManager) newScene(
	name string,
	width, height int,
	plan blueprint.Plan,
	renderSystems []RenderSystem,
	clientSystems []ClientSystem,
	coreSystems []blueprint.CoreSystem,
) (Scene, error) {
	schema := table.Factory.NewSchema()
	storage := warehouse.Factory.NewStorage(schema)
	newScene := &scene{
		plan:    plan,
		storage: storage,
		name:    name,
		height:  height,
		width:   width,
		systems: struct {
			renderers []RenderSystem
			client    []ClientSystem
			core      []blueprint.CoreSystem
		}{
			renderers: renderSystems,
			core:      coreSystems,
			client:    clientSystems,
		},
	}
	return newScene, nil
}

// ActiveScene returns the scene at the specified index
func (m *sceneManager) ActiveScene(index int) Scene {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if index < 0 || index >= len(m.activeScenes) {
		return nil // Or return a zero value, or panic, depending on your error handling strategy
	}

	return m.activeScenes[index]
}

// SceneCount returns the total number of active scenes
func (m *sceneManager) SceneCount() int {
	m.mu.RLock()
	defer m.mu.RUnlock()

	return len(m.activeScenes)
}
