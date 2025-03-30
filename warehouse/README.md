# Warehouse

Warehouse is a high-performance Entity-Component-System (ECS) framework for Go, designed for games and simulations that require efficient entity management and querying.
While its primary purpose is to be the underlying ECS for the [Bappa Framework](https://dl43t3h5ccph3.cloudfront.net/), it functions as a standalone ECS too.

## Features

- **Component-based architecture**: Build entities by composing reusable components
- **Archetype-based storage**: Entities with the same component types are stored together for optimal cache utilization
- **Powerful query system**: Find entities with specific component combinations using AND, OR, and NOT operations
- **Fast iteration**: Optimized for performance with support for Go's iterator pattern

## Installation

```bash
go get github.com/TheBitDrifter/warehouse
```

## Quick Start

```go
package main

import (
 "github.com/TheBitDrifter/table"
 "github.com/TheBitDrifter/warehouse"
)

// Define component types as structs
type Position struct {
 X, Y float64
}

type Velocity struct {
 X, Y float64
}

func main() {
 // Create a schema and storage
 schema := table.Factory.NewSchema()
 storage := warehouse.Factory.NewStorage(schema)

 // Create component accessors
 position := warehouse.FactoryNewComponent[Position]()
 velocity := warehouse.FactoryNewComponent[Velocity]()

 // Create entities with components
 entities, _ := storage.NewEntities(100, position, velocity)

 // Set values for the first entity
 pos := position.GetFromEntity(entities[0])
 vel := velocity.GetFromEntity(entities[0])
 pos.X, pos.Y = 10.0, 20.0
 vel.X, vel.Y = 1.0, 2.0

 // Query for entities with both position and velocity
 query := warehouse.Factory.NewQuery()
 queryNode := query.And(position, velocity)
 cursor := warehouse.Factory.NewCursor(queryNode, storage)

 // Process all matching entities
 for range cursor.Next() {
  pos := position.GetFromCursor(cursor)
  vel := velocity.GetFromCursor(cursor)
  
  // Update position based on velocity
  pos.X += vel.X
  pos.Y += vel.Y
 }
}
```

## Core Concepts

### Entities

Entities are game objects represented by a unique ID. They have no behavior of their own but gain functionality through attached components.

### Components

Components are simple data structs that define attributes and state. They should follow single-responsibility principles and focus on related data.

### Archetypes

Archetypes are collections of entities that share the same component types. This storage pattern optimizes memory layout for cache-friendly access.

### Queries

Queries allow for finding entities with specific component combinations using logical operations (AND, OR, NOT).

### Cursors

Cursors provide efficient iteration over query results for processing matched entities.

## License

MIT License - see the [LICENSE](LICENSE) file for details.
