package blueprint

import (
	blueprintclient "github.com/TheBitDrifter/blueprint/client"
	blueprintinput "github.com/TheBitDrifter/blueprint/input"
	blueprintmotion "github.com/TheBitDrifter/blueprint/motion"
	blueprintspatial "github.com/TheBitDrifter/blueprint/spatial"
	"github.com/TheBitDrifter/warehouse"
)

type defaultQueries struct {
	ParallaxBackground warehouse.Query
	CameraIndex        warehouse.Query
	InputBuffer        warehouse.Query
	Position           warehouse.Query
	Dynamics           warehouse.Query
	Shape              warehouse.Query
	SpriteBundle       warehouse.Query
	SoundBundle        warehouse.Query
}

var Queries defaultQueries = defaultQueries{}

var _ = func() error {
	Queries.ParallaxBackground = warehouse.Factory.NewQuery()
	Queries.ParallaxBackground.And(blueprintclient.Components.ParallaxBackground)

	Queries.CameraIndex = warehouse.Factory.NewQuery()
	Queries.CameraIndex.And(blueprintclient.Components.CameraIndex)

	Queries.InputBuffer = warehouse.Factory.NewQuery()
	Queries.InputBuffer.And(blueprintinput.Components.InputBuffer)

	Queries.Position = warehouse.Factory.NewQuery()
	Queries.Position.And(blueprintspatial.Components.Position)

	Queries.Dynamics = warehouse.Factory.NewQuery()
	Queries.Dynamics.And(blueprintmotion.Components.Dynamics)

	Queries.Shape = warehouse.Factory.NewQuery()
	Queries.Shape.And(blueprintspatial.Components.Shape)

	Queries.SpriteBundle = warehouse.Factory.NewQuery()
	Queries.SpriteBundle.And(blueprintclient.Components.SpriteBundle)

	Queries.SoundBundle = warehouse.Factory.NewQuery()
	Queries.SoundBundle.And(blueprintclient.Components.SoundBundle)
	return nil
}()
