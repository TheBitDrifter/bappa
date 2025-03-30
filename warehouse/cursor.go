package warehouse

import (
	"iter"
	"sync/atomic"
)

// Ensure Cursor implements iCursor interface
var _ iCursor = &Cursor{}

// iCursor defines the interface for iterating over entities in storage
type iCursor interface {
	Next() iter.Seq[*Cursor]
	// Deprecated
	OldNext() bool
}

type iterBitLock32 struct {
	counter atomic.Uint32
}

// Single instance of the bit lock generator
var iterBitLock = &iterBitLock32{}

// Next returns the next unique bit lock value in a thread-safe manner
func (ibl *iterBitLock32) Next() uint32 {
	// Use atomic operations to safely increment and return
	return (ibl.counter.Add(1)) % 257
}

// Cursor provides iteration over filtered entities in storage
type Cursor struct {
	bitLock          uint32
	query            QueryNode
	storage          Storage
	currentArchetype ArchetypeImpl
	storageIndex     int
	entityIndex      int
	remaining        int

	initialized     bool
	matchedStorages []ArchetypeImpl
}

// newCursor creates a new cursor for the given query and storage
func newCursor(query QueryNode, storage Storage) *Cursor {
	return &Cursor{
		query:   query,
		storage: storage,
	}
}

// Deprecated
// OldNext advances to the next entity and returns whether one exists
func (c *Cursor) OldNext() bool {
	if c.entityIndex < c.remaining {
		c.entityIndex++
		return true
	}
	return c.advance()
}

// advance moves to the next available archetype with entities
func (c *Cursor) advance() bool {
	if !c.initialized {
		c.Initialize()
	}

	for c.storageIndex < len(c.matchedStorages) {
		c.currentArchetype = c.matchedStorages[c.storageIndex]
		c.remaining = c.currentArchetype.table.Length()
		if c.entityIndex < c.remaining {
			c.entityIndex++
			return true
		}
		c.storageIndex++
		c.entityIndex = 0
	}

	c.Reset()
	return false
}

// Next advances the cursor and returns it
//
// Whats especially useful about using an iterator pattern here (instead of the deprecated loop version)
// is the automatic cleanup via the magic yield func — It's pretty lit! It also didn't exist when I first started
// looking into this project, so that's pretty cool
func (c *Cursor) Next() iter.Seq[*Cursor] {
	return func(yield func(*Cursor) bool) {
		c.Initialize()

		for c.storageIndex < len(c.matchedStorages) {
			c.currentArchetype = c.matchedStorages[c.storageIndex]
			c.remaining = c.currentArchetype.table.Length()

			for c.entityIndex < c.remaining {
				c.entityIndex++
				if !yield(c) {
					c.Reset()
					return
				}
			}

			c.entityIndex = 0
			c.storageIndex++
		}

		c.Reset()
	}
}

// Initialize sets up the cursor by finding matching archetypes
func (c *Cursor) Initialize() {
	if c.initialized {
		return
	}

	c.bitLock = iterBitLock.Next()
	c.storage.AddLock(c.bitLock)
	c.matchedStorages = make([]ArchetypeImpl, 0)

	// Find all matching archetypes
	for _, arch := range c.storage.Archetypes() {
		if c.query.Evaluate(arch, c.storage) {
			c.matchedStorages = append(c.matchedStorages, arch)
		}
	}

	if len(c.matchedStorages) > 0 {
		c.storageIndex = 0
		c.currentArchetype = c.matchedStorages[0]
		c.remaining = c.currentArchetype.table.Length()
	}

	c.initialized = true
}

// Reset clears cursor state and releases the storage lock
func (c *Cursor) Reset() {
	c.storageIndex = 0
	c.entityIndex = 0
	c.remaining = 0
	c.matchedStorages = nil
	c.initialized = false
	c.storage.RemoveLock(c.bitLock)
	c.bitLock = 0
}

// CurrentEntity returns the entity at the current cursor position
func (c *Cursor) CurrentEntity() (Entity, error) {
	entry, err := c.currentArchetype.table.Entry(c.entityIndex - 1)
	if err != nil {
		return nil, err
	}
	entityID := entry.ID()
	return c.storage.Entity(int(entityID))
}

// EntityAtOffset returns an entity at the specified offset from current position
func (c *Cursor) EntityAtOffset(offset int) (Entity, error) {
	entry, err := c.currentArchetype.table.Entry(c.entityIndex - 1 + offset)
	if err != nil {
		return nil, err
	}
	entityID := entry.ID()
	return c.storage.Entity(int(entityID))
}

// EntityIndex returns the current entity index within the current archetype
func (c *Cursor) EntityIndex() int {
	return c.entityIndex
}

// RemainingInArchetype returns the number of entities left in the current archetype
func (c *Cursor) RemainingInArchetype() int {
	return c.remaining - c.entityIndex
}

// TotalMatched returns the total number of entities matching the query
func (c *Cursor) TotalMatched() int {
	if !c.initialized {
		c.Initialize()
	}

	total := 0
	for _, arch := range c.matchedStorages {
		total += arch.table.Length()
	}

	c.Reset()
	return total
}
