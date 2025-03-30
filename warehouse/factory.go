package warehouse

import "github.com/TheBitDrifter/table"

// factory implements the factory pattern for warehouse components.
type factory struct{}

// Factory is the global factory instance for creating warehouse components.
var Factory factory

// NewStorage creates a new Storage instance with the given schema.
func (f factory) NewStorage(schema table.Schema) Storage {
	return newStorage(schema)
}

// NewQuery creates a new Query instance.
func (f factory) NewQuery() Query {
	return newQuery()
}

// NewCursor creates a new Cursor with the specified query and storage.
func (f factory) NewCursor(query QueryNode, storage Storage) *Cursor {
	return newCursor(query, storage)
}

// FactoryNewComponent creates a new AccessibleComponent for type T.
func FactoryNewComponent[T any]() AccessibleComponent[T] {
	iden := table.FactoryNewElementType[T]()
	return AccessibleComponent[T]{
		Component: iden,
		Accessor:  table.FactoryNewAccessor[T](iden),
	}
}

// FactoryNewCache creates a new Cache with the specified capacity.
func FactoryNewCache[T any](cap int) Cache[T] {
	return &SimpleCache[T]{
		itemIndices: make(map[string]int),
		maxCapacity: cap,
	}
}
