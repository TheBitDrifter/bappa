package combat

import (
	"github.com/TheBitDrifter/bappa/blueprint/vector"
	"github.com/TheBitDrifter/bappa/warehouse"
)

var nextIR = 0

type Hurt struct {
	StartTick int
	Direction vector.Two
}

type invincibleReason int

type Invincible struct {
	StartTick int
	Reason    invincibleReason
}

type Health struct {
	Value int
}

type Defeat struct {
	StartTick int
}

type components struct {
	Attack     warehouse.AccessibleComponent[Attack]
	HurtBox    warehouse.AccessibleComponent[HurtBox]
	HurtBoxes  warehouse.AccessibleComponent[HurtBoxes]
	Invincible warehouse.AccessibleComponent[Invincible]
	Hurt       warehouse.AccessibleComponent[Hurt]
	Health     warehouse.AccessibleComponent[Health]
	Defeat     warehouse.AccessibleComponent[Defeat]
}

var Components = components{
	Attack:     warehouse.FactoryNewComponent[Attack](),
	HurtBox:    warehouse.FactoryNewComponent[HurtBox](),
	HurtBoxes:  warehouse.FactoryNewComponent[HurtBoxes](),
	Invincible: warehouse.FactoryNewComponent[Invincible](),
	Hurt:       warehouse.FactoryNewComponent[Hurt](),
	Health:     warehouse.FactoryNewComponent[Health](),
	Defeat:     warehouse.FactoryNewComponent[Defeat](),
}

func NewIR() invincibleReason {
	nextIR++
	return invincibleReason(nextIR)
}
