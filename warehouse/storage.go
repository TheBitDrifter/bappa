package warehouse

import (
	"errors"
	"fmt"
	"sync"

	"github.com/TheBitDrifter/bappa/table"
	"github.com/TheBitDrifter/mask"
)

// Ensure storage implements Storage interface
var _ Storage = &storage{}

var (
	globalEntryIndex = table.Factory.NewEntryIndex()
	globalEntities   = make([]entity, 0)
)

// Storage defines the interface for entity storage and manipulation
type Storage interface {
	Entity(id int) (Entity, error)
	NewEntities(int, ...Component) ([]Entity, error)
	NewOrExistingArchetype(components ...Component) (Archetype, error)
	EnqueueNewEntities(int, ...Component) error
	DestroyEntities(...Entity) error
	EnqueueDestroyEntities(...Entity) error
	RowIndexFor(Component) uint32
	Locked() bool
	AddLock(bit uint32)
	RemoveLock(bit uint32)
	Register(...Component)
	tableFor(...Component) (table.Table, error)

	TransferEntities(target Storage, entities ...Entity) error
	Enqueue(EntityOperation)
	Archetypes() []ArchetypeImpl
	TotalEntities() int
	Entities() []Entity

	ForceSerializedEntityWithID(SerializedEntity, int) (Entity, error)
	ForceSerializedEntity(SerializedEntity) (Entity, error)
	ForceSerializedEntityExclude(se SerializedEntity, excludeComps ...Component) (Entity, error)
}

// storage implements the Storage interface
type storage struct {
	locks          mask.Mask256
	lockmu         sync.RWMutex
	schema         table.Schema
	archetypes     *archetypes
	operationQueue EntityOperationsQueue
}

// archetypes manages archetype collections and identification
type archetypes struct {
	nextID           archetypeID
	asSlice          []ArchetypeImpl
	idsGroupedByMask map[mask.Mask]archetypeID
}

// newStorage creates a new Storage implementation with the given schema
func newStorage(schema table.Schema) Storage {
	archetypes := &archetypes{
		nextID:           1,
		idsGroupedByMask: make(map[mask.Mask]archetypeID),
	}
	storage := &storage{
		archetypes:     archetypes,
		schema:         schema,
		operationQueue: &entityOperationsQueue{},
	}
	return storage
}

// Entity retrieves an entity by ID
func (sto *storage) Entity(id int) (Entity, error) {
	index := id - 1
	if index < 0 || index >= len(globalEntities) {
		return nil, errors.New("invalid index")
	}
	return &globalEntities[index], nil
}

// NewOrExistingArchetype gets an existing archetype matching the component signature or creates a new one
func (sto *storage) NewOrExistingArchetype(components ...Component) (Archetype, error) {
	var entityMask mask.Mask
	for _, component := range components {
		sto.schema.Register(component)
		bit := sto.schema.RowIndexFor(component)
		entityMask.Mark(bit)
	}
	id, archetypeFound := sto.archetypes.idsGroupedByMask[entityMask]
	if archetypeFound {
		return sto.archetypes.asSlice[id-1], nil
	}

	created, err := newArchetype(sto, globalEntryIndex, sto.archetypes.nextID, components...)
	if err != nil {
		return nil, err
	}
	sto.archetypes.asSlice = append(sto.archetypes.asSlice, created)
	sto.archetypes.idsGroupedByMask[entityMask] = created.id
	sto.archetypes.nextID++
	return &created, nil
}

// NewEntities creates n new entities with the specified components
func (sto *storage) NewEntities(n int, components ...Component) ([]Entity, error) {
	if sto.Locked() {
		return nil, errors.New("storage is locked")
	}

	// Prepare component mask and find/create archetype
	var entityMask mask.Mask
	for _, component := range components {
		sto.schema.Register(component)
		bit := sto.schema.RowIndexFor(component)
		entityMask.Mark(bit)
	}

	var entityArchetype Archetype
	id, archetypeFound := sto.archetypes.idsGroupedByMask[entityMask]
	if archetypeFound {
		entityArchetype = sto.archetypes.asSlice[id-1]
	} else {
		created, err := sto.NewOrExistingArchetype(components...)
		entityArchetype = created
		if err != nil {
			return nil, err
		}
	}

	// Create entries with or without recycling
	entries, err := entityArchetype.Table().NewEntries(n)
	if err != nil {
		return nil, err
	}

	// Resize globalEntities as needed and create entity objects
	entities := make([]Entity, n)

	// Find maximum ID for preallocation
	maxID := uint32(0)
	for _, entry := range entries {
		if uint32(entry.ID()) > maxID {
			maxID = uint32(entry.ID())
		}
	}

	// Expand globalEntities if needed
	if int(maxID) > len(globalEntities) {
		newCap := max(int(maxID), 2*len(globalEntities))
		if cap(globalEntities) < newCap {
			newEntities := make([]entity, len(globalEntities), newCap)
			copy(newEntities, globalEntities)
			globalEntities = newEntities
		}
		globalEntities = globalEntities[:int(maxID)]
	}

	// Create entities and add them to the right positions
	for i, entry := range entries {
		entryID := entry.ID()
		en := &entity{
			Entry:      entry,
			sto:        sto,
			id:         entryID,
			components: components,
		}
		entities[i] = en

		// Place the entity at the correct position in globalEntities
		idx := int(entryID) - 1
		for idx >= len(globalEntities) {
			globalEntities = append(globalEntities, entity{})
		}
		globalEntities[idx] = *en
	}
	return entities, nil
}

// RowIndexFor returns the bit index for a component in the schema
func (sto *storage) RowIndexFor(c Component) uint32 {
	return sto.schema.RowIndexFor(c)
}

// Locked checks if the storage is currently locked
func (sto *storage) Locked() bool {
	sto.lockmu.Lock()
	defer sto.lockmu.Unlock()
	return !sto.locks.IsEmpty()
}

// AddLock adds a bit lock to prevent entity modifications
func (sto *storage) AddLock(bit uint32) {
	sto.lockmu.Lock()
	defer sto.lockmu.Unlock()
	sto.locks.Mark(bit)
}

// RemoveLock releases a specific bit lock and processes queued operations if fully unlocked
func (sto *storage) RemoveLock(bit uint32) {
	sto.lockmu.Lock()
	defer sto.lockmu.Unlock()

	sto.locks.Unmark(bit)

	// Only process operations if no locks remain
	if sto.locks.IsEmpty() {
		// Release the lock before processing queue to avoid deadlocks
		// since processing may involve acquiring the lock again
		sto.lockmu.Unlock()

		err := sto.operationQueue.ProcessAll(sto)
		if err != nil {
			// Handle the error appropriately for your application
			panic(fmt.Errorf("error processing queued operations: %w", err))
		}

		// Re-acquire the lock
		sto.lockmu.Lock()
	}
}

// EnqueueNewEntities either creates entities immediately or queues creation if storage is locked
func (s *storage) EnqueueNewEntities(count int, components ...Component) error {
	if !s.Locked() {
		_, err := s.NewEntities(count, components...)
		if err != nil {
			return fmt.Errorf("failed to create entities directly: %w", err)
		}
		return nil
	}
	s.operationQueue.Enqueue(
		NewEntityOperation{
			count:      count,
			components: components,
		},
	)
	return nil
}

// DestroyEntities removes entities from storage
func (s *storage) DestroyEntities(entities ...Entity) error {
	if s.Locked() {
		return errors.New("storage is locked")
	}
	for _, en := range entities {
		if en == nil || !en.Valid() {
			continue
		}

		table := en.Table()
		_, err := table.DeleteEntries(en.Index())
		if err != nil {
			return err
		}

	}

	return nil
}

// EnqueueDestroyEntities either destroys entities immediately or queues destruction if storage is locked
func (s *storage) EnqueueDestroyEntities(entities ...Entity) error {
	if !s.Locked() {
		return s.DestroyEntities(entities...)
	}
	for _, en := range entities {
		s.operationQueue.Enqueue(
			DestroyEntityOperation{
				entity:   en,
				recycled: en.Recycled(),
			})
	}
	return nil
}

// TransferEntities moves entities from this storage to the target storage
func (s *storage) TransferEntities(target Storage, entities ...Entity) error {
	if s.Locked() {
		return errors.New("storage is locked")
	}
	for _, en := range entities {
		comps := en.Components()
		target.Register(comps...)
		targetTbl, err := target.tableFor(comps...)
		if err != nil {
			return err
		}

		err = en.Table().TransferEntries(targetTbl, en.Index())
		if err != nil {
			return err
		}
		en.SetStorage(target)
	}
	return nil
}

// Register adds components to the storage schema
func (s *storage) Register(comps ...Component) {
	ets := make([]table.ElementType, len(comps))
	for i, c := range comps {
		ets[i] = c
	}
	s.schema.Register(ets...)
}

// Enqueue adds an operation to the queue
func (s *storage) Enqueue(op EntityOperation) {
	s.operationQueue.Enqueue(op)
}

// Archetypes returns all archetypes in this storage
func (s *storage) Archetypes() []ArchetypeImpl {
	return s.archetypes.asSlice
}

func (s *storage) TotalEntities() int {
	total := 0
	for _, archetype := range s.archetypes.asSlice {
		total += archetype.table.Length()
	}
	return total
}

// tableFor gets or creates a table for the given component set
func (s *storage) tableFor(comps ...Component) (table.Table, error) {
	archeMask := mask.Mask{}
	for _, c := range comps {
		bit := s.RowIndexFor(c)
		archeMask.Mark(bit)
	}

	id, ok := s.archetypes.idsGroupedByMask[archeMask]
	decrement := 1
	if !ok {
		decrement++
		created, err := newArchetype(s, globalEntryIndex, s.archetypes.nextID, comps...)
		if err != nil {
			return nil, err
		}
		s.archetypes.asSlice = append(s.archetypes.asSlice, created)
		s.archetypes.nextID++
		id = s.archetypes.nextID
	}
	arche := s.archetypes.asSlice[id-archetypeID(decrement)]
	return arche.table, nil
}

func (s *storage) Entities() []Entity {
	result := []Entity{}

	for _, a := range s.archetypes.asSlice {
		tbl := a.table
		n := tbl.Length()

		for i := 0; i < n; i++ {
			entry, err := tbl.Entry(i)
			if err == nil {
				result = append(result, &globalEntities[entry.ID()-1])
			}
		}
	}
	return result
}

func (s *storage) ForceSerializedEntityWithID(se SerializedEntity, id int) (Entity, error) {
	comps := se.GetComponents()
	index := id - 1

	entityExistsGlobally := id > 0 && index < len(globalEntities) && globalEntities[index].Valid()

	var entityPtr *entity

	if !entityExistsGlobally {
		createdEntity, err := s.forceNewEntity(se) // Creates in globalEntities and table
		if err != nil {
			return nil, fmt.Errorf("failed to force new entity %d: %w", id, err)
		}
		return createdEntity, nil
	}

	entityPtr = &globalEntities[index]

	if entityPtr.sto != s {
		sourceStorage := entityPtr.sto
		if sourceStorage == nil {
			return nil, fmt.Errorf("entity %d exists globally but has nil storage", id)
		}
		err := sourceStorage.TransferEntities(s, entityPtr)
		if err != nil {
			return nil, fmt.Errorf("failed to transfer entity %d from other storage: %w", id, err)
		}

		refreshedEntry, err := globalEntryIndex.Entry(index)
		if err != nil {
			return nil, err
		}
		entityPtr.Entry = refreshedEntry
	}

	targetArchetype, err := s.NewOrExistingArchetype(comps...)
	if err != nil {
		return nil, fmt.Errorf("failed to get target archetype for entity %d: %w", id, err)
	}
	currentMask := entityPtr.Table().(mask.Maskable).Mask()
	targetMask := targetArchetype.Table().(mask.Maskable).Mask()

	if targetMask != currentMask {
		err := entityPtr.Table().TransferEntries(targetArchetype.Table(), entityPtr.Index())
		if err != nil {
			return nil, fmt.Errorf("failed to transfer entity %d to target archetype: %w", id, err)
		}
		refreshedEntry, err := globalEntryIndex.Entry(index)
		if err != nil {
			return nil, err
		}
		entityPtr.Entry = refreshedEntry
	}

	entityPtr.components = comps

	return entityPtr, nil
}

func (s *storage) ForceSerializedEntity(se SerializedEntity) (Entity, error) {
	return s.ForceSerializedEntityWithID(se, int(se.ID))
}

func (s *storage) forceNewEntity(se SerializedEntity) (Entity, error) {
	id := int(se.ID)
	comps := se.GetComponents()
	recycled := se.Recycled

	// Calculate index in the slice (id-1 since arrays are 0-indexed)
	index := id - 1

	// Check how many new slots we need to add
	amountNeeded := index + 1 - len(globalEntities)

	// If we need to expand the slice
	if amountNeeded > 0 {
		// Create a new slice with enough capacity
		newEntities := make([]entity, amountNeeded)
		// Append the new space to the global entities
		globalEntities = append(globalEntities, newEntities...)
	}

	arche, err := s.NewOrExistingArchetype(comps...)
	if err != nil {
		return nil, err
	}
	err = arche.Table().ForceNewEntry(id, recycled)
	if err != nil {
		return nil, err
	}

	// Initialize the entity at the specified index
	globalEntities[index] = entity{
		id:         table.EntryID(id),
		components: comps,
		Entry:      globalEntryIndex.Entries()[id-1],
		sto:        s,
	}

	return &globalEntities[index], nil
}

func (s *storage) ForceSerializedEntityExclude(se SerializedEntity, xComps ...Component) (Entity, error) {
	id := int(se.ID)
	seComps := se.GetComponents()
	index := id - 1

	entityExistsGlobally := id > 0 && index < len(globalEntities) && globalEntities[index].Valid()

	var entityPtr *entity

	if !entityExistsGlobally {
		createdEntity, err := s.forceNewEntity(se)
		if err != nil {
			return nil, fmt.Errorf("failed to force new entity %d: %w", id, err)
		}
		return createdEntity, nil
	}

	entityPtr = &globalEntities[index]

	if entityPtr.sto != s {
		sourceStorage := entityPtr.sto

		if sourceStorage == nil {
			return nil, fmt.Errorf("entity %d exists globally but has nil storage", id)
		}

		err := sourceStorage.TransferEntities(s, entityPtr)
		if err != nil {
			return nil, fmt.Errorf("failed to transfer entity %d from other storage: %w", id, err)
		}

		refreshedEntry, err := globalEntryIndex.Entry(index)
		if err != nil {
			return nil, err
		}
		entityPtr.Entry = refreshedEntry
	}

	existingXComps := mergeUniqueComponents(entityPtr.components, xComps)
	targetComps := mergeUniqueComponents(existingXComps, seComps)

	targetArchetype, err := s.NewOrExistingArchetype(targetComps...)
	if err != nil {
		return nil, fmt.Errorf("failed to get target archetype for entity %d: %w", id, err)
	}
	currentMask := entityPtr.Table().(mask.Maskable).Mask()
	targetMask := targetArchetype.Table().(mask.Maskable).Mask()

	if targetMask != currentMask {
		err := entityPtr.Table().TransferEntries(targetArchetype.Table(), entityPtr.Index())
		if err != nil {
			return nil, fmt.Errorf("failed to transfer entity %d to target archetype: %w", id, err)
		}
		refreshedEntry, err := globalEntryIndex.Entry(index)
		if err != nil {
			return nil, err
		}
		entityPtr.Entry = refreshedEntry
	}

	entityPtr.components = targetComps

	return entityPtr, nil
}

func mergeUniqueComponents(slice1, slice2 []Component) []Component {
	seen := make(map[int]bool)
	result := []Component{}

	for _, comp := range slice1 {
		id := int(comp.ID())

		if !seen[id] {
			seen[id] = true
			result = append(result, comp)
		}
	}
	for _, comp := range slice2 {
		id := int(comp.ID())
		if !seen[id] {
			seen[id] = true
			result = append(result, comp)
		}
	}
	return result
}
