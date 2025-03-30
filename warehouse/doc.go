/*
Package warehouse provides an Entity-Component-System (ECS) framework for games and simulations.

Warehouse offers a performant approach to managing game entities through component-based design.
It's built on an archetype-based storage system that keeps entities with the same component types
together for optimal cache utilization.

Core Concepts:

  - Entity: A unique identifier that represents a game object.
  - Component: A data container that defines entity attributes.
  - Archetype: A collection of entities sharing the same component types.
  - Query: A way to find entities with specific component combinations.

Basic Usage:

	// Create storage with schema
	schema := table.Factory.NewSchema()
	storage := warehouse.Factory.NewStorage(schema)

	// Define components
	position := warehouse.FactoryNewComponent[Position]()
	velocity := warehouse.FactoryNewComponent[Velocity]()

	// Create entities
	entities, _ := storage.NewEntities(100, position, velocity)

	// Query entities and process them
	query := warehouse.Factory.NewQuery()
	queryNode := query.And(position, velocity)
	cursor := warehouse.Factory.NewCursor(queryNode, storage)

	for range cursor.Next() {
		pos := position.GetFromCursor(cursor)
		vel := velocity.GetFromCursor(cursor)
		pos.X += vel.X
		pos.Y += vel.Y
	}

Warehouse is the underlying ECS for the Bappa Framework but also works as a standalone library.
*/
package warehouse
