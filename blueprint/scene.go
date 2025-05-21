package blueprint

import (
	"github.com/TheBitDrifter/bappa/warehouse"
)

type Scene interface {
	NewCursor(warehouse.QueryNode) *warehouse.Cursor
	Height() int
	Width() int
	CurrentTick() int

	Storage() warehouse.Storage
	Name() string
}
