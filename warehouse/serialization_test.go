package warehouse

import (
	"os"
	"testing"

	"github.com/TheBitDrifter/bappa/table"
)

// Test types for serialization
type TestPosition struct {
	X, Y float64
}

type TestVelocity struct {
	X, Y float64
}

// TestPosition2 is an alias type to TestPosition
type TestPosition2 TestPosition

// TestRelationship demonstrates entity references
type TestRelationship struct {
	ChildID table.EntryID
}

// TestSerializationScenarios runs multiple serialization scenarios
func TestSerializationScenarios(t *testing.T) {
	scenarios := []struct {
		name string
		fn   func(t *testing.T)
	}{
		{
			name: "Scenario 1: Creating entities only",
			fn:   testScenarioCreatingEntities,
		},
		{
			name: "Scenario 2: Creating entities then deleting some",
			fn:   testScenarioCreatingThenDeleting,
		},
		{
			name: "Scenario 3: Creating, deleting, then creating more",
			fn:   testScenarioCreateDeleteCreate,
		},
	}

	for _, scenario := range scenarios {
		t.Run(scenario.name, scenario.fn)
	}
}

func testScenarioCreatingEntities(t *testing.T) {
	ResetAll()

	posComp := FactoryNewComponent[TestPosition]()
	velComp := FactoryNewComponent[TestVelocity]()

	schema := table.Factory.NewSchema()
	storage := Factory.NewStorage(schema)

	entities, err := storage.NewEntities(5, posComp)
	if err != nil {
		t.Fatalf("Failed to create entities: %v", err)
	}

	for i, entity := range entities {
		pos := posComp.GetFromEntity(entity)
		pos.X = float64(i * 10)
		pos.Y = float64(i * 5)

		if i < 3 {
			err := entity.AddComponent(velComp)
			if err != nil {
				t.Fatalf("Failed to add velocity component: %v", err)
			}
			vel := velComp.GetFromEntity(entity)
			vel.X = float64(i) * 0.5
			vel.Y = float64(i) * 0.25
		}
	}

	world, err := SerializeStorage(storage, 0)
	if err != nil {
		t.Fatalf("Failed to serialize storage: %v", err)
	}

	if len(world.Entities) != 5 {
		t.Errorf("Expected 5 serialized entities, got %d", len(world.Entities))
	}

	ResetAll()

	newStorage := Factory.NewStorage(table.Factory.NewSchema())
	newStorage, err = DeserializeStorage(newStorage, world)
	if err != nil {
		t.Fatalf("Failed to deserialize storage: %v", err)
	}

	totalEntities := 0
	for _, arch := range newStorage.Archetypes() {
		totalEntities += arch.table.Length()
	}

	if totalEntities != 5 {
		t.Errorf("Deserialized storage has %d entities, expected 5", totalEntities)
	}

	verifyEntitiesWithPositionAndVelocity(t, posComp, velComp, 5, 3, newStorage)

	tempFile := "test_storage_create_only.json"
	err = SaveStorage(storage, tempFile, 0)
	if err != nil {
		t.Fatalf("Failed to save storage: %v", err)
	}
	defer os.Remove(tempFile)

	jsonWorld, err := LoadStorage(tempFile)
	if err != nil {
		t.Fatalf("Failed to load storage: %v", err)
	}

	if len(jsonWorld.Entities) != 5 {
		t.Errorf("JSON loaded %d entities, expected 5", len(jsonWorld.Entities))
	}
}

func testScenarioCreatingThenDeleting(t *testing.T) {
	ResetAll()

	posComp := FactoryNewComponent[TestPosition]()
	velComp := FactoryNewComponent[TestVelocity]()

	schema := table.Factory.NewSchema()
	storage := Factory.NewStorage(schema)

	entities, err := storage.NewEntities(7, posComp)
	if err != nil {
		t.Fatalf("Failed to create entities: %v", err)
	}

	for i, entity := range entities {
		pos := posComp.GetFromEntity(entity)
		pos.X = float64(i * 10)
		pos.Y = float64(i * 5)

		if i < 4 {
			err := entity.AddComponent(velComp)
			if err != nil {
				t.Fatalf("Failed to add velocity component: %v", err)
			}

			vel := velComp.GetFromEntity(entity)
			vel.X = float64(i) * 0.5
			vel.Y = float64(i) * 0.25
		}
	}

	err = storage.DestroyEntities(entities[2], entities[5])
	if err != nil {
		t.Fatalf("Failed to destroy entities: %v", err)
	}

	world, err := SerializeStorage(storage, 0)
	if err != nil {
		t.Fatalf("Failed to serialize storage: %v", err)
	}

	if len(world.Entities) != 5 {
		t.Errorf("Expected 5 serialized entities, got %d", len(world.Entities))
	}

	ResetAll()

	newStorage := Factory.NewStorage(table.Factory.NewSchema())
	newStorage, err = DeserializeStorage(newStorage, world)
	if err != nil {
		t.Fatalf("Failed to deserialize storage: %v", err)
	}

	totalEntities := 0
	for _, arch := range newStorage.Archetypes() {
		totalEntities += arch.table.Length()
	}

	if totalEntities != 5 {
		t.Errorf("Deserialized storage has %d entities, expected 5", totalEntities)
	}

	verifyEntitiesWithPositionAndVelocity(t, posComp, velComp, 5, 3, newStorage)

	tempFile := "test_storage_create_delete.json"
	err = SaveStorage(storage, tempFile, 0)
	if err != nil {
		t.Fatalf("Failed to save storage: %v", err)
	}
	defer os.Remove(tempFile)
}

func testScenarioCreateDeleteCreate(t *testing.T) {
	ResetAll()

	posComp := FactoryNewComponent[TestPosition]()
	velComp := FactoryNewComponent[TestVelocity]()

	schema := table.Factory.NewSchema()
	storage := Factory.NewStorage(schema)

	entities, err := storage.NewEntities(5, posComp)
	if err != nil {
		t.Fatalf("Failed to create entities: %v", err)
	}

	for i, entity := range entities {
		pos := posComp.GetFromEntity(entity)
		pos.X = float64(i * 10)
		pos.Y = float64(i * 5)

		// Add velocity to first 3 entities
		if i < 3 {
			err := entity.AddComponent(velComp)
			if err != nil {
				t.Fatalf("Failed to add velocity component: %v", err)
			}

			vel := velComp.GetFromEntity(entity)
			vel.X = float64(i) * 0.5
			vel.Y = float64(i) * 0.25
		}
	}

	err = storage.DestroyEntities(entities[1], entities[3])
	if err != nil {
		t.Fatalf("Failed to destroy entities: %v", err)
	}

	newEntities, err := storage.NewEntities(3, posComp)
	if err != nil {
		t.Fatalf("Failed to create new entities: %v", err)
	}

	for i, entity := range newEntities {
		pos := posComp.GetFromEntity(entity)
		pos.X = float64((i + 5) * 10)
		pos.Y = float64((i + 5) * 5)

		// Add velocity to first 2 new entities
		if i < 2 {
			err := entity.AddComponent(velComp)
			if err != nil {
				t.Fatalf("Failed to add velocity component: %v", err)
			}

			vel := velComp.GetFromEntity(entity)
			vel.X = float64(i+5) * 0.5
			vel.Y = float64(i+5) * 0.25
		}
	}

	world, err := SerializeStorage(storage, 0)
	if err != nil {
		t.Fatalf("Failed to serialize storage: %v", err)
	}

	// Verify serialized entity count (3 original + 3 new - 2 deleted = 6)
	if len(world.Entities) != 6 {
		t.Errorf("Expected 6 serialized entities, got %d", len(world.Entities))
	}

	ResetAll()

	newStorage := Factory.NewStorage(table.Factory.NewSchema())
	newStorage, err = DeserializeStorage(newStorage, world)
	if err != nil {
		t.Fatalf("Failed to deserialize storage: %v", err)
	}

	// Verify entity count
	totalEntities := 0
	for _, arch := range newStorage.Archetypes() {
		totalEntities += arch.table.Length()
	}

	if totalEntities != 6 {
		t.Errorf("Deserialized storage has %d entities, expected 6", totalEntities)
	}

	verifyEntitiesWithPositionAndVelocity(t, posComp, velComp, 6, 4, newStorage)

	tempFile := "test_storage_create_delete_create.json"
	err = SaveStorage(storage, tempFile, 0)
	if err != nil {
		t.Fatalf("Failed to save storage: %v", err)
	}
	defer os.Remove(tempFile)
}

func verifyEntitiesWithPositionAndVelocity(t *testing.T, posComp AccessibleComponent[TestPosition], velComp AccessibleComponent[TestVelocity], expectedPosCount, expectedVelCount int, storage Storage) {
	// Count entities with position and velocity
	posQuery := Factory.NewQuery()
	posQueryNode := posQuery.And(posComp)
	posCursor := Factory.NewCursor(posQueryNode, storage)

	velQuery := Factory.NewQuery()
	velQueryNode := velQuery.And(velComp)
	velCursor := Factory.NewCursor(velQueryNode, storage)

	posCount := 0
	for range posCursor.Next() {
		entity, err := posCursor.CurrentEntity()
		if err != nil {
			t.Fatalf("Failed to get entity: %v", err)
		}

		pos := posComp.GetFromEntity(entity)
		idx := entity.Index()

		expectedX := pos.X
		expectedY := pos.Y

		if expectedX/10*5 != expectedY {
			t.Errorf("Entity %d: position data inconsistent, X = %v, Y = %v",
				idx, expectedX, expectedY)
		}

		posCount++
	}

	velCount := 0
	for range velCursor.Next() {
		entity, err := velCursor.CurrentEntity()
		if err != nil {
			t.Fatalf("Failed to get entity: %v", err)
		}

		vel := velComp.GetFromEntity(entity)
		pos := posComp.GetFromEntity(entity)

		if vel.X != pos.X*0.05 || vel.Y != pos.Y*0.05 {
			t.Errorf("Entity %d: velocity data inconsistent with position, Pos = {%v, %v}, Vel = {%v, %v}",
				entity.Index(), pos.X, pos.Y, vel.X, vel.Y)
		}

		velCount++
	}

	if posCount != expectedPosCount {
		t.Errorf("Found %d entities with position, expected %d", posCount, expectedPosCount)
	}

	if velCount != expectedVelCount {
		t.Errorf("Found %d entities with velocity, expected %d", velCount, expectedVelCount)
	}
}

// Tests handling of type aliases during serialization
func TestTypeAliasHandling(t *testing.T) {
	ResetAll()
	posComp := FactoryNewComponent[TestPosition]()
	pos2Comp := FactoryNewComponent[TestPosition2]()

	schema := table.Factory.NewSchema()
	storage := Factory.NewStorage(schema)

	entities1, err := storage.NewEntities(2, posComp)
	if err != nil {
		t.Fatalf("Failed to create entities: %v", err)
	}

	entities2, err := storage.NewEntities(2, pos2Comp)
	if err != nil {
		t.Fatalf("Failed to create entities: %v", err)
	}

	for i, entity := range entities1 {
		pos := posComp.GetFromEntity(entity)
		pos.X = float64(i * 10)
		pos.Y = float64(i * 5)
	}

	for i, entity := range entities2 {
		pos := pos2Comp.GetFromEntity(entity)
		pos.X = float64(100 + i*10)
		pos.Y = float64(100 + i*5)
	}

	world, err := SerializeStorage(storage, 0)
	if err != nil {
		t.Fatalf("Failed to serialize storage: %v", err)
	}

	ResetAll()

	newStorage := Factory.NewStorage(table.Factory.NewSchema())
	newStorage, err = DeserializeStorage(newStorage, world)
	if err != nil {
		t.Fatalf("Failed to deserialize storage: %v", err)
	}

	// Verify both component types exist
	posQuery := Factory.NewQuery()
	posQueryNode := posQuery.And(posComp)
	posCursor := Factory.NewCursor(posQueryNode, newStorage)

	pos2Query := Factory.NewQuery()
	pos2QueryNode := pos2Query.And(pos2Comp)
	pos2Cursor := Factory.NewCursor(pos2QueryNode, newStorage)

	// Count entities with TestPosition
	posCount := 0
	for range posCursor.Next() {
		posCount++
	}

	// Count entities with TestPosition2
	pos2Count := 0
	for range pos2Cursor.Next() {
		pos2Count++
	}

	if posCount != 2 {
		t.Errorf("Expected 2 entities with TestPosition, got %d", posCount)
	}

	if pos2Count != 2 {
		t.Errorf("Expected 2 entities with TestPosition2, got %d", pos2Count)
	}
}

// Tests proper handling of entity relationships during serialization
func TestEntityRelationships(t *testing.T) {
	ResetAll()
	// Register component types
	posComp := FactoryNewComponent[TestPosition]()
	relComp := FactoryNewComponent[TestRelationship]()

	// Create storage and entities
	schema := table.Factory.NewSchema()
	storage := Factory.NewStorage(schema)

	// Create parent and child entities
	parentEntities, err := storage.NewEntities(2, posComp, relComp)
	if err != nil {
		t.Fatalf("Failed to create parent entities: %v", err)
	}

	childEntities, err := storage.NewEntities(3, posComp)
	if err != nil {
		t.Fatalf("Failed to create child entities: %v", err)
	}

	// Set up parent-child relationships
	relationships := []struct {
		ParentID table.EntryID
		ChildID  table.EntryID
	}{
		{parentEntities[0].ID(), childEntities[0].ID()},
		{parentEntities[1].ID(), childEntities[2].ID()},
	}

	for i, rel := range relationships {
		relData := relComp.GetFromEntity(parentEntities[i])
		relData.ChildID = rel.ChildID
	}

	// Serialize the storage
	world, err := SerializeStorage(storage, 0)
	if err != nil {
		t.Fatalf("Failed to serialize storage: %v", err)
	}

	ResetAll()

	newStorage := Factory.NewStorage(table.Factory.NewSchema())
	newStorage, err = DeserializeStorage(newStorage, world)
	if err != nil {
		t.Fatalf("Failed to deserialize storage: %v", err)
	}

	// Check each relationship
	for _, rel := range relationships {
		parentID := int(rel.ParentID)
		childID := rel.ChildID

		parentEn, err := newStorage.Entity(parentID)
		if err != nil {
			t.Errorf("Failed to get parent entity %d: %v", parentID, err)
			continue
		}

		relFromEn := relComp.GetFromEntity(parentEn)

		if relFromEn.ChildID != childID {
			t.Errorf("Parent entity %d has child ID %d, expected %d",
				parentID, relFromEn.ChildID, childID)
		}
	}
}
