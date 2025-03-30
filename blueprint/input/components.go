package input

import "github.com/TheBitDrifter/bappa/warehouse"

type defaultComponents struct {
	InputBuffer warehouse.AccessibleComponent[InputBuffer]
}

var Components = defaultComponents{
	InputBuffer: warehouse.FactoryNewComponent[InputBuffer](),
}
