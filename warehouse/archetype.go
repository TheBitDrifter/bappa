package warehouse

import (
	"reflect"

	"github.com/TheBitDrifter/bappa/table"
	"github.com/TheBitDrifter/bark"
	"github.com/TheBitDrifter/mask"
)

// archetypeID is a unique identifier for an archetype
type archetypeID uint32

// Archetype represents a collection of entities with the same component types
type Archetype interface {
	// ID returns the unique identifier of the ArchetypeImpl
	ID() uint32
	// Table returns the underlying data table for the ArchetypeImpl
	Table() table.Table

	// Generate creates entities with the specified components
	Generate(count int, fromComponents ...any) error

	GenerateAndReturnEntity(count int, fromComponents ...any) ([]Entity, error)

	Mask() mask.Mask
}

// ArchetypeImpl is the concrete implementation of the Archetype interface
type ArchetypeImpl struct {
	id         archetypeID
	table      table.Table
	storage    *storage
	components []Component
}

// newArchetypeImpl creates a new archetype with the given components
func newArchetype(
	sto *storage, entryIndex table.EntryIndex, id archetypeID, components ...Component,
) (ArchetypeImpl, error) {
	elementTypes := make([]table.ElementType, len(components))
	for i, comp := range components {
		elementTypes[i] = comp
	}
	tbl, err := table.NewTableBuilder().
		WithSchema(sto.schema).
		WithEntryIndex(entryIndex).
		WithElementTypes(elementTypes...).
		WithEvents(Config.tableEvents).
		WithInitialCapacity(MemConfig.DefaultTableCapacity).
		Build()
	if err != nil {
		return ArchetypeImpl{}, err
	}
	return ArchetypeImpl{
		storage:    sto,
		components: components,
		table:      tbl,
		id:         id,
	}, nil
}

// ID returns the unique identifier of the ArchetypeImpl
func (a ArchetypeImpl) ID() uint32 {
	return uint32(a.id)
}

// Table returns the underlying data table for the ArchetypeImpl
func (a ArchetypeImpl) Table() table.Table {
	return a.table
}

// Generate creates the specified number of entities with optional component values
func (a ArchetypeImpl) Generate(count int, fromComponents ...any) error {
	entities, err := a.storage.NewEntities(count, a.components...)
	if err != nil {
		return err
	}
	// Create mapping from component type to table row for efficient lookups
	reflectTypeToRow := make(map[reflect.Type]table.Row)
	for _, row := range a.table.Rows() {
		reflectTypeToRow[row.Type().Elem()] = row
	}
	// Get logger for this component
	log := bark.For("ArchetypeImpl")

	// Assign component values to each entity
	for _, en := range entities {
		for _, component := range fromComponents {
			compType := reflect.TypeOf(component)
			row, exists := reflectTypeToRow[compType]
			if !exists {
				log.Debug("skipping component not in ArchetypeImpl",
					"component_type", compType.String(),
					"ArchetypeImpl_id", a.id,
					"entity_index", en.Index())
				continue
			}
			compValue := reflect.ValueOf(component)
			reflect.Value(row).Index(en.Index()).Set(compValue)
		}
	}
	return nil
}

// Generate creates the specified number of entities with optional component values
func (a ArchetypeImpl) GenerateAndReturnEntity(count int, fromComponents ...any) ([]Entity, error) {
	entities, err := a.storage.NewEntities(count, a.components...)
	if err != nil {
		return nil, err
	}
	reflectTypeToRow := make(map[reflect.Type]table.Row)
	for _, row := range a.table.Rows() {
		reflectTypeToRow[row.Type().Elem()] = row
	}
	log := bark.For("ArchetypeImpl")

	for _, en := range entities {
		for _, component := range fromComponents {
			compType := reflect.TypeOf(component)
			row, exists := reflectTypeToRow[compType]
			if !exists {
				log.Debug("skipping component not in ArchetypeImpl",
					"component_type", compType.String(),
					"ArchetypeImpl_id", a.id,
					"entity_index", en.Index())
				continue
			}
			compValue := reflect.ValueOf(component)
			reflect.Value(row).Index(en.Index()).Set(compValue)
		}
	}
	return entities, nil
}

func (a ArchetypeImpl) Mask() mask.Mask {
	return a.table.(mask.Maskable).Mask()
}
