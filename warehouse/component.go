package warehouse

import (
	"github.com/TheBitDrifter/table"
)

// Component represents a data container that can be attached to entities.
// Components define the attributes and properties of entities without containing behavior.
//
// Each component type should be a struct with the data fields needed for that attribute.
type Component interface {
	table.ElementType
}
