package motion

import (
	"github.com/TheBitDrifter/warehouse"
)

type components struct {
	Dynamics warehouse.AccessibleComponent[Dynamics]
}

var Components = components{
	Dynamics: warehouse.FactoryNewComponent[Dynamics](),
}
