package warehouse

import (
	"testing"

	"github.com/TheBitDrifter/table"
)

// TestArchetypeCreation tests the creation and reuse of archetypes
func TestArchetypeCreation(t *testing.T) {
	// Create component instances once
	posComp := FactoryNewComponent[Position]()
	velComp := FactoryNewComponent[Velocity]()
	healthComp := FactoryNewComponent[Health]()

	tests := []struct {
		name                string
		firstComponents     []Component
		secondComponents    []Component
		expectSameArchetype bool
	}{
		{
			name:                "Identical components",
			firstComponents:     []Component{posComp, velComp},
			secondComponents:    []Component{posComp, velComp},
			expectSameArchetype: true,
		},
		{
			name:                "Different order",
			firstComponents:     []Component{posComp, velComp},
			secondComponents:    []Component{velComp, posComp},
			expectSameArchetype: true, // Archetypes should be based on component sets, not order
		},
		{
			name:                "Different components",
			firstComponents:     []Component{posComp},
			secondComponents:    []Component{velComp},
			expectSameArchetype: false,
		},
		{
			name:                "Subset components",
			firstComponents:     []Component{posComp, velComp},
			secondComponents:    []Component{posComp},
			expectSameArchetype: false,
		},
		{
			name:                "Superset components",
			firstComponents:     []Component{posComp},
			secondComponents:    []Component{posComp, velComp, healthComp},
			expectSameArchetype: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			schema := table.Factory.NewSchema()
			storage := Factory.NewStorage(schema)

			// Create first archetype
			archetype1, err := storage.NewOrExistingArchetype(tt.firstComponents...)
			if err != nil {
				t.Fatalf("Failed to create first archetype: %v", err)
			}

			// Create second archetype
			archetype2, err := storage.NewOrExistingArchetype(tt.secondComponents...)
			if err != nil {
				t.Fatalf("Failed to create second archetype: %v", err)
			}

			// Check if archetypes are the same
			sameArchetype := archetype1.ID() == archetype2.ID()
			if sameArchetype != tt.expectSameArchetype {
				t.Errorf("Archetypes same: %v, expected: %v", sameArchetype, tt.expectSameArchetype)
			}
		})
	}
}

// TestEntityDestruction tests destroying entities
func TestEntityDestruction(t *testing.T) {
	schema := table.Factory.NewSchema()
	storage := Factory.NewStorage(schema)

	// Create a component to use
	posComp := FactoryNewComponent[Position]()

	// Create some entities
	entities, err := storage.NewEntities(10, posComp)
	if err != nil {
		t.Fatalf("Failed to create entities: %v", err)
	}

	// Destroy half of them
	err = storage.DestroyEntities(entities[0], entities[2], entities[4], entities[6], entities[8])
	if err != nil {
		t.Fatalf("Failed to destroy entities: %v", err)
	}

	// Create a query to count remaining entities
	query := Factory.NewQuery()
	queryNode := query.And(posComp)
	cursor := Factory.NewCursor(queryNode, storage)

	// Count entities
	count := 0
	for range cursor.Next() {
		count++
	}

	// Verify count
	if count != 5 {
		t.Errorf("Entity count after destruction: %d, want 5", count)
	}
}

// TestStorageLocking tests the storage locking mechanism
func TestStorageLocking(t *testing.T) {
	tests := []struct {
		name      string
		lockBits  []uint32
		unlockIdx int    // Index of bit to unlock for midway test
		checks    []bool // Expected lock state at each check
	}{
		{
			name:      "Single lock",
			lockBits:  []uint32{1},
			unlockIdx: 0,
			checks:    []bool{true, false},
		},
		{
			name:      "Multiple locks",
			lockBits:  []uint32{1, 2, 3},
			unlockIdx: 1,
			checks:    []bool{true, true, false}, // Still locked after removing one lock
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			schema := table.Factory.NewSchema()
			storage := Factory.NewStorage(schema)
			posComp := FactoryNewComponent[Position]()

			// Apply all locks
			for _, bit := range tt.lockBits {
				storage.AddLock(bit)
			}

			// Check initial lock state
			if storage.Locked() != tt.checks[0] {
				t.Errorf("Initial lock state: %v, want %v", storage.Locked(), tt.checks[0])
			}

			// Try to create entities while locked (should be queued)
			err := storage.EnqueueNewEntities(5, posComp)
			if err != nil {
				t.Fatalf("EnqueueNewEntities failed: %v", err)
			}

			// Remove one lock
			storage.RemoveLock(tt.lockBits[tt.unlockIdx])

			// Check mid-operation lock state
			if storage.Locked() != tt.checks[1] {
				t.Errorf("Mid-operation lock state: %v, want %v", storage.Locked(), tt.checks[1])
			}

			// Remove all remaining locks
			for i, bit := range tt.lockBits {
				if i != tt.unlockIdx {
					storage.RemoveLock(bit)
				}
			}

			// Check final lock state
			if storage.Locked() != tt.checks[len(tt.checks)-1] {
				t.Errorf("Final lock state: %v, want %v", storage.Locked(), tt.checks[len(tt.checks)-1])
			}

			// Verify entities were created after unlocking
			query := Factory.NewQuery()
			queryNode := query.And(posComp)
			cursor := Factory.NewCursor(queryNode, storage)

			count := 0
			for range cursor.Next() {
				count++
			}

			// Entities should be created now that locks are removed
			if count != 5 {
				t.Errorf("Entity count after unlocking: %d, want 5", count)
			}
		})
	}
}

// TestEntityTransfer tests transferring entities between storages
func TestEntityTransfer(t *testing.T) {
	// Create two storages
	schema1 := table.Factory.NewSchema()
	storage1 := Factory.NewStorage(schema1)

	schema2 := table.Factory.NewSchema()
	storage2 := Factory.NewStorage(schema2)

	// Create components
	posComp := FactoryNewComponent[Position]()
	velComp := FactoryNewComponent[Velocity]()

	// Create entities in storage1
	posEntities, err := storage1.NewEntities(5, posComp)
	if err != nil {
		t.Fatalf("Failed to create position entities: %v", err)
	}

	posVelEntities, err := storage1.NewEntities(5, posComp, velComp)
	if err != nil {
		t.Fatalf("Failed to create position+velocity entities: %v", err)
	}

	// Transfer some entities to storage2
	err = storage1.TransferEntities(storage2, posEntities[0], posEntities[1], posVelEntities[0])
	if err != nil {
		t.Fatalf("Failed to transfer entities: %v", err)
	}

	// Verify storage1 count
	query1 := Factory.NewQuery()
	queryNode1 := query1.And(posComp)
	cursor1 := Factory.NewCursor(queryNode1, storage1)

	count1 := 0
	for range cursor1.Next() {
		count1++
	}

	if count1 != 7 {
		t.Errorf("Entity count in storage1: %d, want 7", count1)
	}

	// Verify storage2 count
	query2 := Factory.NewQuery()
	queryNode2 := query2.And(posComp)
	cursor2 := Factory.NewCursor(queryNode2, storage2)

	count2 := 0
	for range cursor2.Next() {
		count2++
	}

	if count2 != 3 {
		t.Errorf("Entity count in storage2: %d, want 3", count2)
	}

	// Verify that the transferred entities have the correct storage
	for _, entity := range []Entity{posEntities[0], posEntities[1], posVelEntities[0]} {
		if entity.Storage() != storage2 {
			t.Errorf("Entity has incorrect storage after transfer")
		}
	}
}

// TestComponentAccessAfterTransfer tests component access after entity transfer
func TestComponentAccessAfterTransfer(t *testing.T) {
	// Create two storages
	schema1 := table.Factory.NewSchema()
	storage1 := Factory.NewStorage(schema1)

	schema2 := table.Factory.NewSchema()
	storage2 := Factory.NewStorage(schema2)

	// Create components
	posComp := FactoryNewComponent[Position]()
	velComp := FactoryNewComponent[Velocity]()

	// Create entity with position in storage1
	entities, err := storage1.NewEntities(1, posComp)
	if err != nil {
		t.Fatalf("Failed to create entity: %v", err)
	}
	entity := entities[0]

	// Add velocity with value
	vel := Velocity{X: 1.0, Y: 2.0}
	err = entity.AddComponentWithValue(velComp, vel)
	if err != nil {
		t.Fatalf("Failed to add velocity: %v", err)
	}

	// Set position value
	pos := Position{X: 10.0, Y: 20.0}
	posPtr := posComp.GetFromEntity(entity)
	*posPtr = pos

	// Transfer entity to storage2
	err = storage1.TransferEntities(storage2, entity)
	if err != nil {
		t.Fatalf("Failed to transfer entity: %v", err)
	}

	// Verify entity has new storage
	if entity.Storage() != storage2 {
		t.Errorf("Entity has incorrect storage after transfer")
	}

	// Get and check values after transfer
	posPtr = posComp.GetFromEntity(entity)
	velPtr := velComp.GetFromEntity(entity)

	if posPtr.X != pos.X || posPtr.Y != pos.Y {
		t.Errorf("Position after transfer = {%v, %v}, want {%v, %v}",
			posPtr.X, posPtr.Y, pos.X, pos.Y)
	}

	if velPtr.X != vel.X || velPtr.Y != vel.Y {
		t.Errorf("Velocity after transfer = {%v, %v}, want {%v, %v}",
			velPtr.X, velPtr.Y, vel.X, vel.Y)
	}

	// Modify values in the new storage
	posPtr.X = 30.0
	posPtr.Y = 40.0

	// Verify changes persisted
	posPtr2 := posComp.GetFromEntity(entity)
	if posPtr2.X != 30.0 || posPtr2.Y != 40.0 {
		t.Errorf("Updated position after transfer = {%v, %v}, want {30.0, 40.0}",
			posPtr2.X, posPtr2.Y)
	}
}
