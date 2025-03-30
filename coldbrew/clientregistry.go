package coldbrew

// SystemManager manages registration of various system types
type SystemManager interface {
	RegisterGlobalRenderSystem(...GlobalRenderSystem)
	RegisterGlobalClientSystem(...GlobalClientSystem)
}

type systemManager struct {
	globalRenderers     []GlobalRenderSystem
	globalClientSystems []GlobalClientSystem
}

// RegisterGlobalRenderSystem adds one or more render systems to the global renderers
func (sm *systemManager) RegisterGlobalRenderSystem(renderers ...GlobalRenderSystem) {
	for _, r := range renderers {
		sm.globalRenderers = append(sm.globalRenderers, r)
	}
}

// RegisterGlobalClientSystem adds one or more client systems to the global client systems
func (sm *systemManager) RegisterGlobalClientSystem(systems ...GlobalClientSystem) {
	for _, s := range systems {
		sm.globalClientSystems = append(sm.globalClientSystems, s)
	}
}
