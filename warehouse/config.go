package warehouse

import "github.com/TheBitDrifter/table"

// Config holds global configuration for the table system
var Config config = config{}

type config struct {
	tableEvents table.TableEvents
}

// SetTableEvents configures the table event callbacks
func (c *config) SetTableEvents(te table.TableEvents) {
	c.tableEvents = te
}
