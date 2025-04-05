package warehouse

import (
	"sync"
)

type TypeRegistry struct {
	mu         sync.RWMutex
	nameToComp map[string]Component
	compToName map[Component]string
}

var GlobalTypeRegistry = NewTypeRegistry()

// NewTypeRegistry creates a new type registry
func NewTypeRegistry() *TypeRegistry {
	return &TypeRegistry{
		nameToComp: make(map[string]Component),
		compToName: make(map[Component]string),
	}
}

func (r *TypeRegistry) RegisterComp(comp Component) {
	typeName := comp.Type().String()
	r.nameToComp[typeName] = comp
	r.compToName[comp] = typeName
}

// LookupType retrieves a type by name
func (r *TypeRegistry) LookupComp(name string) (Component, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	t, ok := r.nameToComp[name]
	return t, ok
}

func (r *TypeRegistry) LookupName(comp Component) (string, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	name, ok := r.compToName[comp]
	return name, ok
}
