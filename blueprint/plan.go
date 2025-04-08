package blueprint

import "github.com/TheBitDrifter/bappa/warehouse"

type Plan = func(width, height int, storage warehouse.Storage) error
