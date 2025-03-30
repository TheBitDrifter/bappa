package warehouse_test

import (
	"fmt"

	"github.com/TheBitDrifter/table"
	"github.com/TheBitDrifter/warehouse"
)

// Position is a simple component for 2D coordinates
type Position struct {
	X float64
	Y float64
}

// Velocity is a simple component for 2D movement
type Velocity struct {
	X float64
	Y float64
}

// Name is a simple component for entity identification
type Name struct {
	Value string
}

// Example shows basic warehouse usage with entity creation and queries
func Example_basic() {
	// Create storage
	schema := table.Factory.NewSchema()
	storage := warehouse.Factory.NewStorage(schema)

	// Define components
	position := warehouse.FactoryNewComponent[Position]()
	velocity := warehouse.FactoryNewComponent[Velocity]()
	name := warehouse.FactoryNewComponent[Name]()

	// Create entities
	storage.NewEntities(5, position)
	storage.NewEntities(3, position, velocity)

	// Create one named entity
	entities, _ := storage.NewEntities(1, position, velocity, name)
	nameComp := name.GetFromEntity(entities[0])
	nameComp.Value = "Player"

	// Set position and velocity
	pos := position.GetFromEntity(entities[0])
	vel := velocity.GetFromEntity(entities[0])
	pos.X, pos.Y = 10.0, 20.0
	vel.X, vel.Y = 1.0, 2.0

	// Query for all entities with position and velocity
	query := warehouse.Factory.NewQuery()
	queryNode := query.And(position, velocity)
	cursor := warehouse.Factory.NewCursor(queryNode, storage)

	// Count matching entities
	matchCount := 0
	for range cursor.Next() {
		matchCount++
	}
	fmt.Printf("Found %d entities with position and velocity\n", matchCount)

	// Query for just the named entity
	query = warehouse.Factory.NewQuery()
	queryNode = query.And(name)
	cursor = warehouse.Factory.NewCursor(queryNode, storage)

	// Process the named entity
	for range cursor.Next() {
		pos := position.GetFromCursor(cursor)
		vel := velocity.GetFromCursor(cursor)
		nme := name.GetFromCursor(cursor)

		// Update position based on velocity
		pos.X += vel.X
		pos.Y += vel.Y

		fmt.Printf("Updated %s to position (%.1f, %.1f)\n", nme.Value, pos.X, pos.Y)
	}

	// Output:
	// Found 4 entities with position and velocity
	// Updated Player to position (11.0, 22.0)
}

// Example_queries shows how to use different query operations
func Example_queries() {
	// Create storage
	schema := table.Factory.NewSchema()
	storage := warehouse.Factory.NewStorage(schema)

	// Define components
	position := warehouse.FactoryNewComponent[Position]()
	velocity := warehouse.FactoryNewComponent[Velocity]()
	name := warehouse.FactoryNewComponent[Name]()

	// Create different entity types
	storage.NewEntities(3, position)
	storage.NewEntities(3, position, velocity)
	storage.NewEntities(3, position, name)
	storage.NewEntities(3, position, velocity, name)

	// AND query: entities with position AND velocity
	query := warehouse.Factory.NewQuery()
	andQuery := query.And(position, velocity)

	cursor := warehouse.Factory.NewCursor(andQuery, storage)
	fmt.Printf("AND query matched %d entities\n", cursor.TotalMatched())

	// OR query: entities with velocity OR name
	orQuery := query.Or(velocity, name)

	cursor = warehouse.Factory.NewCursor(orQuery, storage)
	fmt.Printf("OR query matched %d entities\n", cursor.TotalMatched())

	// NOT query: entities with position but NOT velocity
	notQuery := query.And(position)
	notQuery = query.Not(velocity)

	cursor = warehouse.Factory.NewCursor(notQuery, storage)
	fmt.Printf("NOT query matched %d entities\n", cursor.TotalMatched())

	// Output:
	// AND query matched 6 entities
	// OR query matched 9 entities
	// NOT query matched 6 entities
}
