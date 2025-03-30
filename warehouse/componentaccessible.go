package warehouse

import "github.com/TheBitDrifter/table"

// AccessibleComponent extends a base Component with table-based accessibility
// It provides methods to retrieve components using different access patterns
type AccessibleComponent[T any] struct {
	Component
	table.Accessor[T] // concrete.
}

// GetFromCursor retrieves a component value for the entity at the cursor position
func (c AccessibleComponent[T]) GetFromCursor(cursor *Cursor) *T {
	return c.Get(
		cursor.entityIndex-1,
		cursor.currentArchetype.table,
	)
}

// GetFromCursorSafe safely retrieves a component value, checking if the component exists
// Returns a boolean indicating success and the component pointer if found
func (c AccessibleComponent[T]) GetFromCursorSafe(cursor *Cursor) (bool, *T) {
	ok := c.Accessor.Check(cursor.currentArchetype.table)
	if ok {
		return true, c.GetFromCursor(cursor)
	}
	return false, nil
}

// CheckCursor determines if the component exists in the archetype at the cursor position
func (c AccessibleComponent[T]) CheckCursor(cursor *Cursor) bool {
	return c.Accessor.Check(cursor.currentArchetype.table)
}

// GetFromEntity retrieves a component value for the specified entity
func (c AccessibleComponent[T]) GetFromEntity(entity Entity) *T {
	return c.Get(entity.Index(), entity.Table())
}
