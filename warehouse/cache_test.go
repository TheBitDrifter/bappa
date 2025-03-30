package warehouse

import (
	"testing"
)

// TestCacheBasicOperations tests the basic operations of the SimpleCache
func TestCacheBasicOperations(t *testing.T) {
	// Create a cache with a fixed capacity
	const capacity = 10
	cache := FactoryNewCache[string](capacity)

	// Register some items
	items := []string{"item1", "item2", "item3", "item4", "item5"}
	indices := make([]int, len(items))

	for i, item := range items {
		index, err := cache.Register(item, item)
		if err != nil {
			t.Errorf("Failed to register item %s: %v", item, err)
		}
		indices[i] = index

		// Verify index starts at 1 and increments
		if index != i+1 {
			t.Errorf("Index for item %s is %d, expected %d", item, index, i+1)
		}
	}

	// Get indices
	for i, item := range items {
		index, found := cache.GetIndex(item)
		if !found {
			t.Errorf("Item %s not found in cache", item)
		}
		if index != indices[i] {
			t.Errorf("Index for item %s is %d, expected %d", item, index, indices[i])
		}
	}

	// Get items by index
	for i, item := range items {
		cachedItem := cache.GetItem(indices[i])
		if cachedItem != item {
			t.Errorf("Item at index %d is %s, expected %s", indices[i], cachedItem, item)
		}
	}

	// Get items by uint32 index
	for i, item := range items {
		cachedItem := cache.GetItem32(uint32(indices[i]))
		if cachedItem != item {
			t.Errorf("Item at index %d is %s, expected %s", indices[i], cachedItem, item)
		}
	}

	// Test for non-existent item
	_, found := cache.GetIndex("nonexistent")
	if found {
		t.Errorf("Found non-existent item in cache")
	}
}

// TestCacheCapacity tests the cache capacity limits
func TestCacheCapacity(t *testing.T) {
	// Create a cache with a small capacity
	const capacity = 5
	cache := FactoryNewCache[int](capacity)

	// Register up to capacity
	for i := 1; i <= capacity; i++ {
		key := "item" + string(rune(i+'0'))
		_, err := cache.Register(key, i)
		if err != nil {
			t.Errorf("Failed to register item %s: %v", key, err)
		}
	}

	// Try to register one more (should fail)
	_, err := cache.Register("overflow", 100)
	if err == nil {
		t.Errorf("Expected error when exceeding cache capacity, but got none")
	}
}

// TestCacheClear tests the cache clear functionality
func TestCacheClear(t *testing.T) {
	// Create a cache and cast to SimpleCache to access Clear method
	cache := FactoryNewCache[string](10).(*SimpleCache[string])

	// Register some items
	items := []string{"item1", "item2", "item3"}
	for _, item := range items {
		_, err := cache.Register(item, item)
		if err != nil {
			t.Errorf("Failed to register item %s: %v", item, err)
		}
	}

	// Clear the cache
	cache.Clear()

	// Verify items are gone
	for _, item := range items {
		_, found := cache.GetIndex(item)
		if found {
			t.Errorf("Item %s still found after cache clear", item)
		}
	}

	// Verify we can add items again
	for _, item := range items {
		_, err := cache.Register(item, item)
		if err != nil {
			t.Errorf("Failed to register item %s after clear: %v", item, err)
		}
	}
}

// TestCacheWithComplexTypes tests the cache with more complex data types
func TestCacheWithComplexTypes(t *testing.T) {
	// Create a cache for position structs
	cache := FactoryNewCache[Position](10)

	// Register some positions
	positions := []Position{
		{X: 1.0, Y: 2.0},
		{X: 3.0, Y: 4.0},
		{X: 5.0, Y: 6.0},
	}

	keys := []string{"pos1", "pos2", "pos3"}

	// Register positions
	for i, pos := range positions {
		_, err := cache.Register(keys[i], pos)
		if err != nil {
			t.Errorf("Failed to register position %v: %v", pos, err)
		}
	}

	// Retrieve positions
	for i, key := range keys {
		index, found := cache.GetIndex(key)
		if !found {
			t.Errorf("Position with key %s not found", key)
			continue
		}

		pos := cache.GetItem(index)
		if pos.X != positions[i].X || pos.Y != positions[i].Y {
			t.Errorf("Position at index %d is %v, expected %v", index, pos, positions[i])
		}
	}
}

// TestCacheConcurrentAccess tests concurrent access to the cache
// Note: This is just a basic concurrent access test. More sophisticated tests might use the race detector.
func TestCacheConcurrentAccess(t *testing.T) {
	// Create a cache
	cache := FactoryNewCache[int](100)

	// Register an initial item
	initialIndex, err := cache.Register("item", 42)
	if err != nil {
		t.Fatalf("Failed to register initial item: %v", err)
	}

	// Create done channel
	done := make(chan struct{})

	// Start a goroutine that reads from the cache
	go func() {
		defer close(done)
		for i := 0; i < 1000; i++ {
			// Read the initial item
			item := cache.GetItem(initialIndex)
			if item != 42 {
				t.Errorf("Expected item value 42, got %d", item)
				return
			}
		}
	}()

	// In the main goroutine, add more items
	for i := 0; i < 50; i++ {
		key := "new_item" + string(rune(i+'0'))
		_, err := cache.Register(key, i)
		if err != nil {
			// Error might be expected if capacity is reached
			break
		}
	}

	// Wait for reader goroutine to finish
	<-done
}
