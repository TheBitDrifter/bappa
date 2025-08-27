package warehouse

var MemConfig = &memConfig{}

type memConfig struct {
	InitialEntityCapacity int
	DefaultTableCapacity  int
}

func (*memConfig) Preallocate() {
	if MemConfig.InitialEntityCapacity > 0 {
		globalEntities = make([]entity, 0, MemConfig.InitialEntityCapacity)
		globalEntryIndex.Preallocate(MemConfig.InitialEntityCapacity)
	}
}

func (c *memConfig) Set(initialEntityCapacity, defaultTableCapacity int) {
	if MemConfig.InitialEntityCapacity > 0 {
		c.InitialEntityCapacity = initialEntityCapacity
		c.DefaultTableCapacity = defaultTableCapacity
	}
}
