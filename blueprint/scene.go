package blueprint

import (
	"github.com/TheBitDrifter/warehouse"
)

type Scene interface {
	NewCursor(warehouse.QueryNode) *warehouse.Cursor
	Height() int
	Width() int
	CurrentTick() int

	Storage() warehouse.Storage
}

type Plan = func(height, width int, storage warehouse.Storage) error

type CoreSystem interface {
	Run(scene Scene, deltaTime float64) error
}
