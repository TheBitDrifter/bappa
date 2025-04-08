package warehouse

import (
	"testing"

	"github.com/TheBitDrifter/bappa/table"
)

// TestUpdateSerialization tests the non-destructive update-based serialization behavior
func TestUpdateSerialization(t *testing.T) {
	// Run each test in isolation with a clean state
	t.Run("UpdateExistingEntityValues", func(t *testing.T) {
		ResetAll() // Reset before each test
		testUpdateExistingEntityValues(t)
	})
	t.Run("EntityIDPreservation", func(t *testing.T) {
		ResetAll()
		testEntityIDPreservation(t)
	})
	t.Run("AddComponentsToExistingEntities", func(t *testing.T) {
		ResetAll()
		testAddComponentsToExistingEntities(t)
	})
	t.Run("RemoveComponentsFromExistingEntities", func(t *testing.T) {
		ResetAll()
		testRemoveComponentsFromExistingEntities(t)
	})
	t.Run("MixedOperations", func(t *testing.T) {
		ResetAll()
		testMixedOperations(t)
	})
	t.Run("EntityReferences", func(t *testing.T) {
		ResetAll()
		testEntityReferences(t)
	})
	t.Run("LargeGapsInEntityIDs", func(t *testing.T) {
		ResetAll()
		testLargeGapsInEntityIDs(t)
	})
	t.Run("EntityPurging", func(t *testing.T) {
		ResetAll()
		testEntityPurging(t)
	})
	t.Run("ConcurrentStorageManagement", func(t *testing.T) {
		ResetAll()
		testConcurrentStorageManagement(t)
	})
}

// testUpdateExistingEntityValues tests updating existing entities with new component values
func testUpdateExistingEntityValues(t *testing.T) {
	// Create component types
	posComp := FactoryNewComponent[TestPosition]()

	// Create initial storage with entities
	schema := table.Factory.NewSchema()
	storage := Factory.NewStorage(schema)

	// Create entities with initial values
	entities, err := storage.NewEntities(3, posComp)
	if err != nil {
		t.Fatalf("Failed to create entities: %v", err)
	}

	// Set initial positions
	for i, entity := range entities {
		pos := posComp.GetFromEntity(entity)
		pos.X = float64(i * 10)
		pos.Y = float64(i * 5)
	}

	// Verify initial entity count
	initialCount := storage.TotalEntities()
	if initialCount != 3 {
		t.Fatalf("Expected 3 initial entities, got %d", initialCount)
	}

	// Serialize the storage
	serialized, err := SerializeStorage(storage, 0)
	if err != nil {
		t.Fatalf("Failed to serialize storage: %v", err)
	}

	// Verify serialized entity count
	if len(serialized.Entities) != 3 {
		t.Fatalf("Expected 3 serialized entities, got %d", len(serialized.Entities))
	}

	// Modify the serialized data to have new values
	for i := range serialized.Entities {
		typeName := posComp.Type().String()
		posData, ok := serialized.Entities[i].Data[typeName].(TestPosition)
		if !ok {
			t.Fatalf("Failed to cast position data to TestPosition")
		}
		serialized.Entities[i].Data[typeName] = TestPosition{
			X: posData.X * 2, // Double the X value
			Y: posData.Y * 2, // Double the Y value
		}
	}

	// Deserialize back to the same storage (update-based approach)
	updatedStorage, err := DeserializeStorage(storage, serialized)
	if err != nil {
		t.Fatalf("Failed to deserialize storage: %v", err)
	}

	// Verify entity count is still the same
	finalCount := updatedStorage.TotalEntities()
	if finalCount != 3 {
		t.Errorf("Expected 3 entities after deserialization, got %d", finalCount)
	}

	// Verify the entities were updated with new values
	for i, entity := range entities {
		pos := posComp.GetFromEntity(entity)
		expectedX := float64(i*10) * 2 // Original * 2
		expectedY := float64(i*5) * 2  // Original * 2

		if pos.X != expectedX || pos.Y != expectedY {
			t.Errorf("Entity %d: position not updated correctly, got {%v, %v}, expected {%v, %v}",
				i, pos.X, pos.Y, expectedX, expectedY)
		}
	}
}

// testEntityIDPreservation tests that entity IDs are preserved across serialization cycles
func testEntityIDPreservation(t *testing.T) {
	// Create component types
	posComp := FactoryNewComponent[TestPosition]()

	// Create initial storage with entities
	schema := table.Factory.NewSchema()
	storage := Factory.NewStorage(schema)

	// Create entities with specific IDs
	entities, err := storage.NewEntities(5, posComp)
	if err != nil {
		t.Fatalf("Failed to create entities: %v", err)
	}

	// Record original entity IDs
	originalIDs := make([]table.EntryID, len(entities))
	for i, entity := range entities {
		originalIDs[i] = entity.ID()
	}

	// Serialize then deserialize
	serialized, err := SerializeStorage(storage, 0)
	if err != nil {
		t.Fatalf("Failed to serialize storage: %v", err)
	}

	// Deserialize to the same storage
	updatedStorage, err := DeserializeStorage(storage, serialized)
	if err != nil {
		t.Fatalf("Failed to deserialize storage: %v", err)
	}

	// Get all entities from updated storage
	allEntities := updatedStorage.Entities()

	// Create map of entity IDs for easy lookup
	entityMap := make(map[table.EntryID]bool)
	for _, entity := range allEntities {
		entityMap[entity.ID()] = true
	}

	// Verify all original IDs still exist
	for i, id := range originalIDs {
		if !entityMap[id] {
			t.Errorf("Entity ID %d (at index %d) not preserved after deserialization", id, i)
		}
	}
}

// testAddComponentsToExistingEntities tests adding components to existing entities
func testAddComponentsToExistingEntities(t *testing.T) {
	// Create component types
	posComp := FactoryNewComponent[TestPosition]()
	velComp := FactoryNewComponent[TestVelocity]()

	// Create initial storage with entities having only position
	schema := table.Factory.NewSchema()
	storage := Factory.NewStorage(schema)

	entities, err := storage.NewEntities(3, posComp)
	if err != nil {
		t.Fatalf("Failed to create entities: %v", err)
	}

	// Set initial positions
	for i, entity := range entities {
		pos := posComp.GetFromEntity(entity)
		pos.X = float64(i * 10)
		pos.Y = float64(i * 5)
	}

	// Verify initial state
	initialCount := storage.TotalEntities()
	if initialCount != 3 {
		t.Fatalf("Expected 3 initial entities, got %d", initialCount)
	}

	// Check initial components
	query := Factory.NewQuery()
	velQuery := query.And(velComp)
	velCursor := Factory.NewCursor(velQuery, storage)

	velCount := 0
	for range velCursor.Next() {
		velCount++
	}

	if velCount != 0 {
		t.Fatalf("Expected 0 entities with velocity initially, got %d", velCount)
	}

	// Serialize the storage
	serialized, err := SerializeStorage(storage, 0)
	if err != nil {
		t.Fatalf("Failed to serialize storage: %v", err)
	}

	// Modify the serialized data to add velocity component
	for i := range serialized.Entities {
		// Add velocity component type
		typeName := velComp.Type().String()
		serialized.Entities[i].Components = append(
			serialized.Entities[i].Components,
			typeName,
		)

		// Add velocity component data
		serialized.Entities[i].Data[typeName] = TestVelocity{
			X: 1.0,
			Y: 2.0,
		}
	}

	// Deserialize back to the same storage
	updatedStorage, err := DeserializeStorage(storage, serialized)
	if err != nil {
		t.Fatalf("Failed to deserialize storage: %v", err)
	}

	// Verify entity count remains the same
	finalCount := updatedStorage.TotalEntities()
	if finalCount != 3 {
		t.Errorf("Expected 3 entities after deserialization, got %d", finalCount)
	}

	// Create a query for entities with both position and velocity
	bothQuery := Factory.NewQuery()
	bothQueryNode := bothQuery.And(posComp, velComp)
	bothCursor := Factory.NewCursor(bothQueryNode, updatedStorage)

	// Count entities with both components
	entitiesWithBoth := 0
	for range bothCursor.Next() {
		entitiesWithBoth++
	}

	// Verify all entities now have both components
	if entitiesWithBoth != 3 {
		t.Errorf("Expected 3 entities with both position and velocity, got %d", entitiesWithBoth)
	}

	// Verify component values
	bothCursor = Factory.NewCursor(bothQueryNode, updatedStorage)
	for range bothCursor.Next() {
		entity, err := bothCursor.CurrentEntity()
		if err != nil {
			t.Fatalf("Failed to get entity: %v", err)
		}

		vel := velComp.GetFromEntity(entity)
		if vel.X != 1.0 || vel.Y != 2.0 {
			t.Errorf("Entity %d: velocity not set correctly, got {%v, %v}, expected {1.0, 2.0}",
				entity.ID(), vel.X, vel.Y)
		}
	}
}

// testRemoveComponentsFromExistingEntities tests removing components from existing entities
func testRemoveComponentsFromExistingEntities(t *testing.T) {
	// Create component types
	posComp := FactoryNewComponent[TestPosition]()
	velComp := FactoryNewComponent[TestVelocity]()

	// Create initial storage with entities having both position and velocity
	schema := table.Factory.NewSchema()
	storage := Factory.NewStorage(schema)

	entities, err := storage.NewEntities(3, posComp, velComp)
	if err != nil {
		t.Fatalf("Failed to create entities: %v", err)
	}

	// Set initial component values
	for i, entity := range entities {
		pos := posComp.GetFromEntity(entity)
		pos.X = float64(i * 10)
		pos.Y = float64(i * 5)

		vel := velComp.GetFromEntity(entity)
		vel.X = 1.0
		vel.Y = 2.0
	}

	// Verify initial state
	initialCount := storage.TotalEntities()
	if initialCount != 3 {
		t.Fatalf("Expected 3 initial entities, got %d", initialCount)
	}

	// Check initial components
	bothQuery := Factory.NewQuery()
	bothQueryNode := bothQuery.And(posComp, velComp)
	bothCursor := Factory.NewCursor(bothQueryNode, storage)

	bothCount := 0
	for range bothCursor.Next() {
		bothCount++
	}

	if bothCount != 3 {
		t.Fatalf("Expected 3 entities with both components initially, got %d", bothCount)
	}

	// Serialize the storage
	serialized, err := SerializeStorage(storage, 0)
	if err != nil {
		t.Fatalf("Failed to serialize storage: %v", err)
	}

	// Modify the serialized data to remove velocity component
	for i := range serialized.Entities {
		// Keep only the position component
		typeName := posComp.Type().String()
		serialized.Entities[i].Components = []string{typeName}

		// Remove velocity data
		delete(serialized.Entities[i].Data, velComp.Type().String())
	}

	// Deserialize back to the same storage
	updatedStorage, err := DeserializeStorage(storage, serialized)
	if err != nil {
		t.Fatalf("Failed to deserialize storage: %v", err)
	}

	// Verify entity count remains the same
	finalCount := updatedStorage.TotalEntities()
	if finalCount != 3 {
		t.Errorf("Expected 3 entities after deserialization, got %d", finalCount)
	}

	// Create queries for entities with position and entities with velocity
	posQuery := Factory.NewQuery()
	posNode := posQuery.And(posComp)
	posCursor := Factory.NewCursor(posNode, updatedStorage)

	velQuery := Factory.NewQuery()
	velNode := velQuery.And(velComp)
	velCursor := Factory.NewCursor(velNode, updatedStorage)

	// Count entities with each component
	posCount := 0
	for range posCursor.Next() {
		posCount++
	}

	velCount := 0
	for range velCursor.Next() {
		velCount++
	}

	// Verify all entities have position but none have velocity
	if posCount != 3 {
		t.Errorf("Expected 3 entities with position, got %d", posCount)
	}

	if velCount != 0 {
		t.Errorf("Expected 0 entities with velocity, got %d", velCount)
	}
}

// testMixedOperations tests a mix of creating, updating, and removing entities
func testMixedOperations(t *testing.T) {
	// Create component types
	posComp := FactoryNewComponent[TestPosition]()
	velComp := FactoryNewComponent[TestVelocity]()

	// Create initial storage with some entities
	schema := table.Factory.NewSchema()
	storage := Factory.NewStorage(schema)

	// Create 5 entities with position
	entities, err := storage.NewEntities(5, posComp)
	if err != nil {
		t.Fatalf("Failed to create entities: %v", err)
	}

	// Add velocity to first 2 entities
	for i := 0; i < 2; i++ {
		err := entities[i].AddComponentWithValue(velComp, TestVelocity{X: 1.0, Y: 2.0})
		if err != nil {
			t.Fatalf("Failed to add velocity component: %v", err)
		}
	}

	// Set positions for all entities
	for i, entity := range entities {
		pos := posComp.GetFromEntity(entity)
		pos.X = float64(i * 10)
		pos.Y = float64(i * 5)
	}

	// Verify initial state
	initialCount := storage.TotalEntities()
	if initialCount != 5 {
		t.Fatalf("Expected 5 initial entities, got %d", initialCount)
	}

	// Serialize the storage
	serialized, err := SerializeStorage(storage, 0)
	if err != nil {
		t.Fatalf("Failed to serialize storage: %v", err)
	}

	// Verify serialized entity count
	if len(serialized.Entities) != 5 {
		t.Fatalf("Expected 5 serialized entities, got %d", len(serialized.Entities))
	}

	// Update entity[1]'s position
	entity1ID := entities[1].ID()
	for i, se := range serialized.Entities {
		if se.ID == entity1ID {
			serialized.Entities[i].Data[posComp.Type().String()] = TestPosition{X: 100.0, Y: 200.0}
			break
		}
	}

	// Remove entity[2] (will be purged)
	entity2ID := entities[2].ID()
	newEntities := make([]SerializedEntity, 0)
	for _, se := range serialized.Entities {
		if se.ID != entity2ID {
			newEntities = append(newEntities, se)
		}
	}
	serialized.Entities = newEntities

	// Add velocity to entity[3]
	entity3ID := entities[3].ID()
	for i, se := range serialized.Entities {
		if se.ID == entity3ID {
			serialized.Entities[i].Components = append(
				serialized.Entities[i].Components,
				velComp.Type().String(),
			)
			serialized.Entities[i].Data[velComp.Type().String()] = TestVelocity{X: 3.0, Y: 4.0}
			break
		}
	}

	// Add a completely new entity with ID 6
	newEntityID := table.EntryID(6) // Assuming IDs 1-5 are taken
	newEntity := SerializedEntity{
		ID:         newEntityID,
		Recycled:   0,
		Components: []string{posComp.Type().String(), velComp.Type().String()},
		Data: map[string]any{
			posComp.Type().String(): TestPosition{X: 50.0, Y: 60.0},
			velComp.Type().String(): TestVelocity{X: 5.0, Y: 6.0},
		},
	}
	serialized.Entities = append(serialized.Entities, newEntity)

	// Verify the modified serialized data has 5 entities
	// (4 original - 1 removed + 1 new = 5)
	if len(serialized.Entities) != 5 {
		t.Fatalf("Expected 5 entities in modified serialized data, got %d", len(serialized.Entities))
	}

	// Deserialize back to the same storage
	updatedStorage, err := DeserializeStorage(storage, serialized)
	if err != nil {
		t.Fatalf("Failed to deserialize storage: %v", err)
	}

	// Verify total entity count (should be 5: entities[0,1,3,4] + new entity)
	finalCount := updatedStorage.TotalEntities()
	if finalCount != 5 {
		t.Errorf("Expected 5 entities after mixed operations, got %d", finalCount)
	}

	// Verify entity[0] remains unchanged
	entity0, err := updatedStorage.Entity(int(entities[0].ID()))
	if err != nil {
		t.Fatalf("Failed to get entity[0]: %v", err)
	}
	pos0 := posComp.GetFromEntity(entity0)
	if pos0.X != 0.0 || pos0.Y != 0.0 {
		t.Errorf("Entity[0] position changed, got {%v, %v}, expected {0.0, 0.0}", pos0.X, pos0.Y)
	}

	// Verify entity[1]'s position was updated
	entity1, err := updatedStorage.Entity(int(entities[1].ID()))
	if err != nil {
		t.Fatalf("Failed to get entity[1]: %v", err)
	}
	pos1 := posComp.GetFromEntity(entity1)
	if pos1.X != 100.0 || pos1.Y != 200.0 {
		t.Errorf("Entity[1] position not updated, got {%v, %v}, expected {100.0, 200.0}", pos1.X, pos1.Y)
	}

	// Verify entity[2] was purged (should not exist or be invalid)
	entity2, err := updatedStorage.Entity(int(entities[2].ID()))
	if err == nil && entity2.Valid() {
		t.Errorf("Entity[2] still exists, should have been purged")
	}

	// Verify entity[3] has velocity component added
	entity3, err := updatedStorage.Entity(int(entities[3].ID()))
	if err != nil {
		t.Fatalf("Failed to get entity[3]: %v", err)
	}

	// Check if entity3 has velocity component
	vel3 := velComp.GetFromEntity(entity3)
	if vel3 == nil {
		t.Errorf("Entity[3] should have velocity component added")
	} else if vel3.X != 3.0 || vel3.Y != 4.0 {
		t.Errorf("Entity[3] velocity not set correctly, got {%v, %v}, expected {3.0, 4.0}", vel3.X, vel3.Y)
	}

	// Verify new entity was created
	newEntityList := updatedStorage.Entities()
	newEntityFound := false
	for _, entity := range newEntityList {
		if entity.ID() == newEntityID {
			newEntityFound = true
			pos := posComp.GetFromEntity(entity)
			vel := velComp.GetFromEntity(entity)

			if pos.X != 50.0 || pos.Y != 60.0 {
				t.Errorf("New entity position not set correctly, got {%v, %v}, expected {50.0, 60.0}", pos.X, pos.Y)
			}

			if vel.X != 5.0 || vel.Y != 6.0 {
				t.Errorf("New entity velocity not set correctly, got {%v, %v}, expected {5.0, 6.0}", vel.X, vel.Y)
			}

			break
		}
	}

	if !newEntityFound {
		t.Errorf("New entity with ID %d not found", newEntityID)
	}
}

// testEntityReferences tests preservation of entity references
func testEntityReferences(t *testing.T) {
	// Create component types
	posComp := FactoryNewComponent[TestPosition]()
	relComp := FactoryNewComponent[TestRelationship]()

	// Create initial storage
	schema := table.Factory.NewSchema()
	storage := Factory.NewStorage(schema)

	// Create parent entities with position and relationship components
	parentEntities, err := storage.NewEntities(2, posComp, relComp)
	if err != nil {
		t.Fatalf("Failed to create parent entities: %v", err)
	}

	// Create child entities with just position
	childEntities, err := storage.NewEntities(2, posComp)
	if err != nil {
		t.Fatalf("Failed to create child entities: %v", err)
	}

	// Set up parent-child relationships
	for i, parent := range parentEntities {
		// Set parent position
		parentPos := posComp.GetFromEntity(parent)
		parentPos.X = float64((i + 1) * 10)
		parentPos.Y = float64((i + 1) * 20)

		// Set child position
		childPos := posComp.GetFromEntity(childEntities[i])
		childPos.X = float64((i + 1) * 5)
		childPos.Y = float64((i + 1) * 15)

		// Set relationship
		rel := relComp.GetFromEntity(parent)
		rel.ChildID = childEntities[i].ID()
	}

	// Serialize the storage
	serialized, err := SerializeStorage(storage, 0)
	if err != nil {
		t.Fatalf("Failed to serialize storage: %v", err)
	}

	// Modify child entity positions in serialized data
	for i, se := range serialized.Entities {
		// Find child entities
		for _, childEntity := range childEntities {
			if se.ID == childEntity.ID() {
				// Change child position
				childIndex := -1
				for j, child := range childEntities {
					if child.ID() == se.ID {
						childIndex = j
						break
					}
				}

				if childIndex >= 0 {
					// Multiply position by 10
					posData := se.Data[posComp.Type().String()].(TestPosition)
					serialized.Entities[i].Data[posComp.Type().String()] = TestPosition{
						X: posData.X * 10,
						Y: posData.Y * 10,
					}
				}
				break
			}
		}
	}

	// Deserialize back to the same storage
	updatedStorage, err := DeserializeStorage(storage, serialized)
	if err != nil {
		t.Fatalf("Failed to deserialize storage: %v", err)
	}

	// Verify entity references are preserved and child positions updated
	for i, parent := range parentEntities {
		// Get updated parent
		updatedParent, err := updatedStorage.Entity(int(parent.ID()))
		if err != nil {
			t.Fatalf("Failed to get updated parent entity %d: %v", parent.ID(), err)
		}

		// Get parent relationship
		rel := relComp.GetFromEntity(updatedParent)

		// Verify relationship preserved
		if rel.ChildID != childEntities[i].ID() {
			t.Errorf("Parent %d: relationship not preserved, got child ID %d, expected %d",
				parent.ID(), rel.ChildID, childEntities[i].ID())
			continue
		}

		// Get child entity
		childEntity, err := updatedStorage.Entity(int(rel.ChildID))
		if err != nil {
			t.Fatalf("Failed to get child entity %d: %v", rel.ChildID, err)
		}

		// Verify child position was updated
		childPos := posComp.GetFromEntity(childEntity)
		expectedX := float64((i + 1) * 5 * 10)  // Original * 10
		expectedY := float64((i + 1) * 15 * 10) // Original * 10

		if childPos.X != expectedX || childPos.Y != expectedY {
			t.Errorf("Child entity %d: position not updated correctly, got {%v, %v}, expected {%v, %v}",
				childEntity.ID(), childPos.X, childPos.Y, expectedX, expectedY)
		}
	}
}

// testLargeGapsInEntityIDs tests handling cases with large gaps in entity IDs
func testLargeGapsInEntityIDs(t *testing.T) {
	// Create component types
	posComp := FactoryNewComponent[TestPosition]()

	// Create initial storage
	schema := table.Factory.NewSchema()
	storage := Factory.NewStorage(schema)

	// Create serialized entities with large ID gaps
	// Map of ID -> Position values
	idToPosition := map[table.EntryID]TestPosition{
		1:    {X: 10.0, Y: 20.0},
		100:  {X: 100.0, Y: 200.0},
		1000: {X: 1000.0, Y: 2000.0},
	}

	// Create serialized data structure
	serialized := &SerializedStorage{
		Version:  "1.0",
		Entities: []SerializedEntity{},
	}

	// Add entities to serialized data
	for id, pos := range idToPosition {
		serialized.Entities = append(serialized.Entities, SerializedEntity{
			ID:         id,
			Recycled:   0,
			Components: []string{posComp.Type().String()},
			Data: map[string]any{
				posComp.Type().String(): pos,
			},
		})
	}

	// Deserialize to storage
	updatedStorage, err := DeserializeStorage(storage, serialized)
	if err != nil {
		t.Fatalf("Failed to deserialize storage with large ID gaps: %v", err)
	}

	// Verify all entities were created
	entityCount := updatedStorage.TotalEntities()
	expectedCount := len(idToPosition)
	if entityCount != expectedCount {
		t.Errorf("Expected %d entities, got %d", expectedCount, entityCount)
	}

	// Verify each entity exists with correct ID and data
	for id, expectedPos := range idToPosition {
		entity, err := updatedStorage.Entity(int(id))
		if err != nil {
			t.Errorf("Entity with ID %d not found: %v", id, err)
			continue
		}

		pos := posComp.GetFromEntity(entity)
		if pos.X != expectedPos.X || pos.Y != expectedPos.Y {
			t.Errorf("Entity %d: incorrect position, got {%v, %v}, expected {%v, %v}",
				id, pos.X, pos.Y, expectedPos.X, expectedPos.Y)
		}
	}

	// Check that intermediate entities are not created (they should be nil or invalid)
	for _, id := range []int{2, 50, 500} {
		entity, err := updatedStorage.Entity(id)
		if err == nil && entity.Valid() {
			t.Errorf("Unexpected entity with ID %d found", id)
		}
	}
}

// testEntityPurging tests that entities in storage not present in deserialized data are purged
func testEntityPurging(t *testing.T) {
	// Create component types
	posComp := FactoryNewComponent[TestPosition]()

	// Create initial storage with entities
	schema := table.Factory.NewSchema()
	storage := Factory.NewStorage(schema)

	// Create 5 entities
	entities, err := storage.NewEntities(5, posComp)
	if err != nil {
		t.Fatalf("Failed to create entities: %v", err)
	}

	// Set positions
	for i, entity := range entities {
		pos := posComp.GetFromEntity(entity)
		pos.X = float64(i * 10)
		pos.Y = float64(i * 5)
	}

	// Record the IDs we want to keep
	keepIDs := make([]table.EntryID, 0, 3)
	keepIndices := []int{0, 2, 4}
	for _, idx := range keepIndices {
		keepIDs = append(keepIDs, entities[idx].ID())
	}

	// Serialize the storage
	serialized, err := SerializeStorage(storage, 0)
	if err != nil {
		t.Fatalf("Failed to serialize storage: %v", err)
	}

	// Count initial serialized entities
	initialSerializedCount := len(serialized.Entities)
	if initialSerializedCount != 5 {
		t.Fatalf("Expected 5 entities in initial serialized data, got %d", initialSerializedCount)
	}

	// Modify serialized data to keep only entities 0, 2, and 4
	// Entities 1 and 3 should be purged during deserialization
	keptEntities := make([]SerializedEntity, 0)
	for _, se := range serialized.Entities {
		for _, id := range keepIDs {
			if se.ID == id {
				keptEntities = append(keptEntities, se)
				break
			}
		}
	}
	serialized.Entities = keptEntities

	// Verify we have only 3 entities in serialized data
	if len(serialized.Entities) != 3 {
		t.Fatalf("Expected 3 entities in modified serialized data, got %d", len(serialized.Entities))
	}

	// Deserialize back to the same storage
	_, err = DeserializeStorage(storage, serialized)
	if err != nil {
		t.Fatalf("Failed to deserialize storage: %v", err)
	}

	// Verify entity count is now 3
	currentCount := storage.TotalEntities()
	if currentCount != 3 {
		t.Errorf("Expected 3 entities after purging, got %d", currentCount)
	}

	// Check which entities exist
	for i, entity := range entities {
		shouldExist := i == 0 || i == 2 || i == 4

		// Try to get the entity
		retrievedEntity, err := storage.Entity(int(entity.ID()))
		entityExists := (err == nil && retrievedEntity.Valid())

		if shouldExist && !entityExists {
			t.Errorf("Entity %d (ID %d) should exist but doesn't", i, entity.ID())
		} else if !shouldExist && entityExists {
			t.Errorf("Entity %d (ID %d) should have been purged but still exists", i, entity.ID())
		}

		// If entity exists, check its position is preserved
		if entityExists {
			pos := posComp.GetFromEntity(retrievedEntity)
			expectedX := float64(i * 10)
			expectedY := float64(i * 5)

			if pos.X != expectedX || pos.Y != expectedY {
				t.Errorf("Entity %d position changed: got {%v, %v}, expected {%v, %v}",
					i, pos.X, pos.Y, expectedX, expectedY)
			}
		}
	}

	// Create a new entity in the storage
	newEntities, err := storage.NewEntities(1, posComp)
	if err != nil {
		t.Fatalf("Failed to create new entity after purging: %v", err)
	}

	// Set position
	newPos := posComp.GetFromEntity(newEntities[0])
	newPos.X = 999.0
	newPos.Y = 888.0

	// Verify new total count
	newCount := storage.TotalEntities()
	expectedNewCount := 4 // 3 kept + 1 new
	if newCount != expectedNewCount {
		t.Errorf("Expected %d entities after adding new entity, got %d", expectedNewCount, newCount)
	}

	// Serialize again
	serialized2, err := SerializeStorage(storage, 0)
	if err != nil {
		t.Fatalf("Failed to serialize storage after adding new entity: %v", err)
	}

	// Verify new entity is included in serialized data
	newEntityFound := false
	for _, se := range serialized2.Entities {
		if se.ID == newEntities[0].ID() {
			newEntityFound = true
			break
		}
	}

	if !newEntityFound {
		t.Errorf("New entity ID %d not found in serialized data", newEntities[0].ID())
	}

	// Verify total count is correct
	expectedSerializedCount := 4 // 3 kept + 1 new
	if len(serialized2.Entities) != expectedSerializedCount {
		t.Errorf("Expected %d entities in serialized data, got %d",
			expectedSerializedCount, len(serialized2.Entities))
	}
}

// testConcurrentStorageManagement tests deserializing into one storage while entities exist in another
func testConcurrentStorageManagement(t *testing.T) {
	// Create component types
	posComp := FactoryNewComponent[TestPosition]()
	velComp := FactoryNewComponent[TestVelocity]()

	// Create two separate storages
	schema1 := table.Factory.NewSchema()
	storage1 := Factory.NewStorage(schema1)

	schema2 := table.Factory.NewSchema()
	storage2 := Factory.NewStorage(schema2)

	// Create 3 entities in storage1
	entities1, err := storage1.NewEntities(3, posComp)
	if err != nil {
		t.Fatalf("Failed to create entities in storage1: %v", err)
	}

	// Set positions in storage1
	for i, entity := range entities1 {
		pos := posComp.GetFromEntity(entity)
		pos.X = float64(i * 10)
		pos.Y = float64(i * 5)
	}

	// Verify storage1 has 3 entities
	if storage1.TotalEntities() != 3 {
		t.Fatalf("Expected 3 entities in storage1, got %d", storage1.TotalEntities())
	}

	// Serialize storage1
	serialized1, err := SerializeStorage(storage1, 0)
	if err != nil {
		t.Fatalf("Failed to serialize storage1: %v", err)
	}

	if len(serialized1.Entities) != 3 {
		t.Fatalf("Expected 3 entities in serialized data, got %d", len(serialized1.Entities))
	}

	// Deserialize into storage2
	storage2, err = DeserializeStorage(storage2, serialized1)
	if err != nil {
		t.Fatalf("Failed to deserialize storage1 into storage2: %v", err)
	}

	// Verify storage2 has 3 entities
	if storage2.TotalEntities() != 3 {
		t.Errorf("Expected 3 entities in storage2 after deserialization, got %d", storage2.TotalEntities())
	}

	// Check that entities from storage1 are now in storage2 with correct positions
	for i, entityID := range []int{1, 2, 3} {
		entity2, err := storage2.Entity(entityID)
		if err != nil {
			t.Errorf("Entity %d not found in storage2: %v", entityID, err)
			continue
		}

		if !entity2.Valid() {
			t.Errorf("Entity %d is not valid in storage2", entityID)
			continue
		}

		// Check if entity is properly associated with storage2
		if entity2.Storage() != storage2 {
			t.Errorf("Entity %d is not associated with storage2", entityID)
		}

		// Check position values
		pos := posComp.GetFromEntity(entity2)
		expectedX := float64((i) * 10)
		expectedY := float64((i) * 5)

		if pos.X != expectedX || pos.Y != expectedY {
			t.Errorf("Entity %d position is incorrect: got {%v, %v}, expected {%v, %v}",
				entityID, pos.X, pos.Y, expectedX, expectedY)
		}
	}

	// Add velocity to first entity in storage2
	firstEntity, err := storage2.Entity(1)
	if err != nil {
		t.Fatalf("Failed to get first entity from storage2: %v", err)
	}

	err = firstEntity.AddComponentWithValue(velComp, TestVelocity{X: 1.5, Y: 2.5})
	if err != nil {
		t.Fatalf("Failed to add velocity to first entity: %v", err)
	}

	// Serialize storage2
	serialized2, err := SerializeStorage(storage2, 0)
	if err != nil {
		t.Fatalf("Failed to serialize storage2: %v", err)
	}

	// Create a third storage and deserialize
	schema3 := table.Factory.NewSchema()
	storage3 := Factory.NewStorage(schema3)

	storage3, err = DeserializeStorage(storage3, serialized2)
	if err != nil {
		t.Fatalf("Failed to deserialize into storage3: %v", err)
	}

	// Verify storage3 has 3 entities
	if storage3.TotalEntities() != 3 {
		t.Errorf("Expected 3 entities in storage3, got %d", storage3.TotalEntities())
	}

	// Check that first entity has velocity component
	firstEntityInStorage3, err := storage3.Entity(1)
	if err != nil {
		t.Fatalf("Failed to get first entity from storage3: %v", err)
	}

	// Check if velocity component exists on the entity
	hasVelocity := false
	for _, comp := range firstEntityInStorage3.Components() {
		if comp.ID() == velComp.ID() {
			hasVelocity = true
			break
		}
	}

	if !hasVelocity {
		t.Errorf("First entity in storage3 should have velocity component")
	} else {
		// Verify velocity values
		vel := velComp.GetFromEntity(firstEntityInStorage3)
		if vel.X != 1.5 || vel.Y != 2.5 {
			t.Errorf("First entity velocity incorrect: got {%v, %v}, expected {1.5, 2.5}",
				vel.X, vel.Y)
		}
	}

	// Verify other 2 entities have position only
	for i := 2; i <= 3; i++ {
		entity, err := storage3.Entity(i)
		if err != nil {
			t.Errorf("Entity %d not found in storage3: %v", i, err)
			continue
		}

		componentCount := len(entity.Components())
		if componentCount != 1 {
			t.Errorf("Entity %d should have 1 component, got %d", i, componentCount)
		}
	}
}
