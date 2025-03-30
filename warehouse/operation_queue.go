package warehouse

// EntityOperation represents an operation that can be applied to a storage
type EntityOperation interface {
	Apply(Storage) error
}

// entityOperationsQueue holds a list of operations to be processed
type entityOperationsQueue struct {
	operations []EntityOperation
}

// EntityOperationsQueue provides an interface for queuing and processing operations
type EntityOperationsQueue interface {
	Enqueue(EntityOperation)
	ProcessAll(Storage) error
}

// ProcessAll applies all queued operations to the provided storage
// and clears the queue afterward
func (queue *entityOperationsQueue) ProcessAll(sto Storage) error {
	// If storage is locked, keep operations in queue for later processing
	if sto.Locked() {
		return nil // Return without error, but don't clear queue
	}
	for _, op := range queue.operations {
		err := op.Apply(sto)
		if err != nil {
			return err
		}
	}
	queue.operations = []EntityOperation{}
	return nil
}

// Enqueue adds an operation to the queue
func (queue *entityOperationsQueue) Enqueue(op EntityOperation) {
	queue.operations = append(queue.operations, op)
}

// NewEntityOperation creates multiple entities with the same components
type NewEntityOperation struct {
	count      int
	components []Component
}

// Apply creates entities with the specified components
func (op NewEntityOperation) Apply(sto Storage) error {
	entityArchetype, err := sto.NewOrExistingArchetype(op.components...)
	if err != nil {
		return err
	}
	err = entityArchetype.Generate(op.count)
	if err != nil {
		return err
	}
	return nil
}

// DestroyEntityOperation removes an entity from storage
type DestroyEntityOperation struct {
	entity   Entity
	recycled int
}

// Apply destroys the entity if it's valid and has the expected recycled value
func (op DestroyEntityOperation) Apply(sto Storage) error {
	if !op.entity.Valid() {
		return nil
	}
	if op.entity.Recycled() != op.recycled {
		return nil
	}
	err := sto.DestroyEntities(op.entity)
	if err != nil {
		return err
	}
	return nil
}

// TransferEntityOperation moves an entity from one storage to another
type TransferEntityOperation struct {
	target   Storage
	entity   Entity
	recycled int
}

// Apply transfers the entity if it's valid and has the expected recycled value
func (op TransferEntityOperation) Apply(sto Storage) error {
	if !op.entity.Valid() {
		return nil
	}
	if op.entity.Recycled() != op.recycled {
		return nil
	}
	err := sto.TransferEntities(op.target, op.entity)
	if err != nil {
		return err
	}
	return nil
}

// AddComponentOperation adds a component to an entity
type AddComponentOperation struct {
	entity    Entity
	recycled  int
	component Component
	value     any
	storage   Storage
}

// Apply adds the component to the entity if conditions are met
func (op AddComponentOperation) Apply(sto Storage) error {
	if !op.entity.Valid() {
		return nil
	}
	if op.entity.Recycled() != op.recycled {
		return nil
	}
	if op.storage != op.entity.Storage() {
		return nil
	}
	if op.value != nil {
		err := op.entity.AddComponentWithValue(op.component, op.value)
		if err != nil {
			return err
		}
		return nil
	}
	err := op.entity.AddComponent(op.component)
	if err != nil {
		return err
	}
	return nil
}

// RemoveComponentOperation removes a component from an entity
type RemoveComponentOperation struct {
	entity    Entity
	recycled  int
	component Component
	storage   Storage
}

// Apply removes the component from the entity if conditions are met
func (op RemoveComponentOperation) Apply(sto Storage) error {
	if !op.entity.Valid() {
		return nil
	}
	if op.entity.Recycled() != op.recycled {
		return nil
	}
	if op.storage != sto {
		return nil
	}
	err := op.entity.RemoveComponent(op.component)
	if err != nil {
		return err
	}
	return nil
}
