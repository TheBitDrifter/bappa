
# Overview

The table package implements a high-performance Entity Component System [(ECS)](https://github.com/SanderMertens/ecs-faq) focused on efficient data organization and processing.

**Note:** for more in depth documentation please see: [bappa-docs](https://dl43t3h5ccph3.cloudfront.net/docs/table/overview/)

## Key Features

- Schema-based component organization
- Safe and unsafe memory access modes
- Efficient entity recycling
- Cache-friendly memory layout
- Type-safe component access

## Use Cases

- Game Development (physics engines, behavior systems, particle effects, etc)
- Simulations (ecosystem modeling, vehicle traffic, etc)
- Real-time Systems (trading systems, sensor processing, etc)

## Example Usage

```go
// Define components
type Position struct { X, Y float64 }
type Velocity struct { X, Y float64 }

// Create table
schema := table.Factory.NewSchema()
entryIndex := table.Factory.NewEntryIndex()
posType := table.FactoryNewElementType[Position]()
velType := table.FactoryNewElementType[Velocity]()

tbl, _ := table.Factory.NewTable(schema, entryIndex, posType, velType)

// Add entities
tbl.NewEntries(1000)

// Access components
posAccessor := table.FactoryNewAccessor[Position](posType)
velAccessor := table.FactoryNewAccessor[Velocity](velType)

// Process entities
for i := 0; i < tbl.Length(); i++ {
    pos := posAccessor.Get(i, table)
    vel := velAccessor.Get(i, table)
    pos.X += vel.X * dt
    pos.Y += vel.Y * dt
}
```

## Configuration

```go
// Build tags for optimization
//go:build unsafe         // Enable unsafe pointer operations
//go:build schema_enabled // Enable flexible schemas
//go:build m256          // Configure mask size up to 1024
```

## Performance Characteristics

- O(1) component access
- Cache-coherent iteration
- Low memory fragmentation

## When to Use

- Need component-based architecture
- Performance critical systems
- Large numbers of similar entities
- Frequent batch processing
- Cache-sensitive operations
