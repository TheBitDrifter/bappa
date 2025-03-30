package blueprintspatial

import (
	"github.com/TheBitDrifter/warehouse"
)

type components struct {
	Position         warehouse.AccessibleComponent[Position]
	PreviousPosition warehouse.AccessibleComponent[Position]
	Rotation         warehouse.AccessibleComponent[Rotation]
	Scale            warehouse.AccessibleComponent[Scale]
	Shape            warehouse.AccessibleComponent[Shape]
	Direction        warehouse.AccessibleComponent[Direction]
}

var Components = components{
	Position:         warehouse.FactoryNewComponent[Position](),
	PreviousPosition: warehouse.FactoryNewComponent[Position](),
	Rotation:         warehouse.FactoryNewComponent[Rotation](),
	Scale:            warehouse.FactoryNewComponent[Scale](),
	Shape:            warehouse.FactoryNewComponent[Shape](),
	Direction:        warehouse.FactoryNewComponent[Direction](),
}
