package warehouse

import (
	"errors"
	"fmt"
	"reflect"
	"sort"
	"strings"

	"github.com/TheBitDrifter/bark"
	"github.com/TheBitDrifter/table"
)

// Verify entity implements Entity interface
var _ Entity = &entity{}

// Entity represents a game object that's composed of components.
// Entities are essentially IDs that tie related components together.
//
// An entity's behavior is determined by which components are attached to it.
// Adding or removing components changes the entity's archetype (its component signature).
type Entity interface {
	table.Entry

	// AddComponent attaches a component to this entity
	AddComponent(Component) error

	// AddComponentWithValue attaches a component with an initial value
	AddComponentWithValue(Component, any) error

	// RemoveComponent detaches a component from this entity
	RemoveComponent(Component) error

	// EnqueueAddComponent schedules a component addition when storage is locked
	EnqueueAddComponent(Component) error

	// EnqueueAddComponentWithValue schedules a component addition with value when storage is locked
	EnqueueAddComponentWithValue(Component, any) error

	// EnqueueRemoveComponent schedules a component removal when storage is locked
	EnqueueRemoveComponent(Component) error

	// Components returns all components attached to this entity
	Components() []Component

	// ComponentsAsString returns a string representation of component names
	ComponentsAsString() string

	// Valid returns whether this entity has a valid ID
	Valid() bool

	// Storage returns the storage this entity belongs to
	Storage() Storage

	// SetStorage changes the storage this entity belongs to
	SetStorage(Storage)
}

// EntityDestroyCallback is called when an entity is destroyed
type EntityDestroyCallback func(Entity)

// entity implements the Entity interface
type entity struct {
	table.Entry
	id         table.EntryID
	sto        Storage
	components []Component
}

// ID returns the entity's unique identifier
func (e *entity) ID() table.EntryID {
	return e.id
}

// Index returns the entity's index in its table
func (e *entity) Index() int {
	return e.entry().Index()
}

// Recycled returns the entity's recycled count
func (e *entity) Recycled() int {
	return e.entry().Recycled()
}

// Table returns the table this entity belongs to
func (e *entity) Table() table.Table {
	return e.entry().Table()
}

// Storage returns the storage this entity belongs to
func (e *entity) Storage() Storage {
	return e.sto
}

// AddComponent adds a component to the entity, moving it to a new archetype if needed
func (e *entity) AddComponent(c Component) error {
	if e.sto.Locked() {
		return errors.New("storage is locked")
	}

	originTable := e.Table()
	if originTable.Contains(c) {
		return nil
	}

	// Check if the component already exists in the entity's component list
	for _, comp := range e.components {
		if comp.ID() == c.ID() {
			return nil // Component already exists, nothing to do
		}
	}

	e.components = append(e.components, c)
	destArchetype, err := e.sto.NewOrExistingArchetype(e.components...)
	if err != nil {
		return err
	}
	if err := originTable.TransferEntries(destArchetype.Table(), e.Index()); err != nil {
		return err
	}
	return nil
}

// AddComponentWithValue adds a component with an initial value
func (e *entity) AddComponentWithValue(c Component, value any) error {
	if e.sto.Locked() {
		return errors.New("storage is locked")
	}

	originTable := e.Table()
	if originTable.Contains(c) {
		return nil
	}

	// Check if the component already exists in the entity's component list
	for _, comp := range e.components {
		if comp.ID() == c.ID() {
			return nil // Component already exists, nothing to do
		}
	}

	e.components = append(e.components, c)
	destArchetype, err := e.sto.NewOrExistingArchetype(e.components...)
	if err != nil {
		return err
	}
	if err := originTable.TransferEntries(destArchetype.Table(), e.Index()); err != nil {
		return err
	}

	valueType := reflect.TypeOf(value)
	for _, row := range destArchetype.Table().Rows() {
		if row.Type().Elem() == valueType {
			reflect.Value(row).Index(e.Index()).Set(reflect.ValueOf(value))
			return nil
		}
	}
	return fmt.Errorf("invalid value type %v for component %v", valueType, c.Type())
}

// RemoveComponent removes a component from the entity, moving it to a new archetype
func (e *entity) RemoveComponent(c Component) error {
	if e.sto.Locked() {
		return errors.New("storage is locked")
	}
	originTable := e.Table()
	if !originTable.Contains(c) {
		return nil
	}
	newComps := []Component{}
	for _, comp := range e.components {
		if comp.ID() != c.ID() {
			newComps = append(newComps, comp)
		}
	}
	e.components = newComps
	destArchetype, err := e.sto.NewOrExistingArchetype(newComps...)
	if err != nil {
		return fmt.Errorf("failed to get/create archetype: %w", err)
	}
	if err := originTable.TransferEntries(destArchetype.Table(), e.Index()); err != nil {
		return fmt.Errorf("failed to transfer entity: %w", err)
	}
	return nil
}

// EnqueueAddComponent queues a component addition or executes immediately if storage isn't locked
func (e *entity) EnqueueAddComponent(c Component) error {
	if !e.sto.Locked() {
		return e.AddComponent(c)
	}
	e.sto.Enqueue(AddComponentOperation{
		entity:    e,
		recycled:  e.Recycled(),
		component: c,
		storage:   e.sto,
	})
	return nil
}

// EnqueueAddComponentWithValue queues a component addition with value or executes immediately
func (e *entity) EnqueueAddComponentWithValue(c Component, val any) error {
	if !e.sto.Locked() {
		return e.AddComponentWithValue(c, val)
	}
	e.sto.Enqueue(AddComponentOperation{
		entity:    e,
		recycled:  e.Recycled(),
		component: c,
		value:     val,
		storage:   e.sto,
	})
	return nil
}

// EnqueueRemoveComponent queues a component removal or executes immediately if storage isn't locked
func (e *entity) EnqueueRemoveComponent(c Component) error {
	if !e.sto.Locked() {
		return e.RemoveComponent(c)
	}
	e.sto.Enqueue(RemoveComponentOperation{
		entity:    e,
		recycled:  e.Recycled(),
		component: c,
		storage:   e.sto,
	})
	return nil
}

// entry returns the table entry for this entity
func (e *entity) entry() table.Entry {
	en, err := globalEntryIndex.Entry(int(e.id - 1))
	if err != nil {
		panic(bark.AddTrace(err))
	}
	return en
}

// Components returns all components attached to this entity
func (e *entity) Components() []Component {
	return e.components
}

// ComponentsAsString returns a sorted, formatted string of component names
func (e *entity) ComponentsAsString() string {
	if len(e.components) == 0 {
		return "[]"
	}

	var components []string
	for _, c := range e.components {
		typeName := reflect.TypeOf(c).String()
		typeName = strings.TrimPrefix(typeName, "*")
		parts := strings.Split(typeName, ".")
		name := parts[len(parts)-1]
		name = strings.TrimSuffix(name, "]")

		components = append(components, name)
	}

	sort.Strings(components)

	return "[" + strings.Join(components, ", ") + "]"
}

// Valid returns whether this entity has a valid ID
func (e entity) Valid() bool {
	return e.id != 0
}

// SetStorage sets the storage for this entity
func (e *entity) SetStorage(sto Storage) {
	e.sto = sto
}
