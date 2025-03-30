package warehouse

import (
	"fmt"
	"sync"
	"sync/atomic"
)

// Ensure SimpleCache implements the Cache interface
var _ Cache[any] = &SimpleCache[any]{}

// Cache defines a generic thread-safe cache interface
type Cache[T any] interface {
	// GetIndex retrieves the index of an item by its key
	GetIndex(string) (int, bool)
	// GetItem retrieves an item by its index
	GetItem(int) T
	// GetItem32 retrieves an item by its uint32 index
	GetItem32(uint32) T
	// Register adds a new item to the cache with the given key
	Register(string, T) (int, error)
}

// CacheLocation represents the position of an item in a cache
type CacheLocation struct {
	Key   string
	Index atomic.Uint32
}

// SimpleCache implements the Cache interface with a slice-backed storage
type SimpleCache[T any] struct {
	mu          sync.RWMutex
	items       []T
	itemIndices map[string]int
	maxCapacity int
}

// GetIndex retrieves the index of an item by its key
func (c *SimpleCache[T]) GetIndex(key string) (int, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	index, ok := c.itemIndices[key]
	return index, ok
}

// GetItem retrieves an item by its index
func (c *SimpleCache[T]) GetItem(index int) T {
	c.mu.RLock()
	defer c.mu.RUnlock()
	item := c.items[index-1]
	return item
}

// GetItem32 retrieves an item by its uint32 index
func (c *SimpleCache[T]) GetItem32(index uint32) T {
	c.mu.RLock()
	defer c.mu.RUnlock()
	item := c.items[index-1]
	return item
}

// Register adds a new item to the cache with the given key
// Returns the index of the newly added item or an error if the cache is full
func (c *SimpleCache[T]) Register(key string, item T) (int, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if len(c.itemIndices) >= c.maxCapacity {
		return -1, fmt.Errorf("cache at maximum capacity (%d)", c.maxCapacity)
	}
	idx := len(c.items) + 1
	c.itemIndices[key] = idx
	c.items = append(c.items, item)
	return idx, nil
}

// Clear removes all items from the cache
func (c *SimpleCache[T]) Clear() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.items = make([]T, 0, c.maxCapacity)
	c.itemIndices = make(map[string]int)
}
