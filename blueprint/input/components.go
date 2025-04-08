package input

import "github.com/TheBitDrifter/bappa/warehouse"

type defaultComponents struct {
	ActionBuffer warehouse.AccessibleComponent[ActionBuffer]
}

var Components = defaultComponents{
	ActionBuffer: warehouse.FactoryNewComponent[ActionBuffer](),
}
