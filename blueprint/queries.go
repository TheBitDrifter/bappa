package blueprint

import (
	"github.com/TheBitDrifter/bappa/blueprint/client"
	"github.com/TheBitDrifter/bappa/blueprint/input"
	"github.com/TheBitDrifter/bappa/tteokbokki/motion"
	"github.com/TheBitDrifter/bappa/tteokbokki/spatial"
	"github.com/TheBitDrifter/bappa/warehouse"
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
	Queries.ParallaxBackground.And(client.Components.ParallaxBackground)

	Queries.CameraIndex = warehouse.Factory.NewQuery()
	Queries.CameraIndex.And(client.Components.CameraIndex)

	Queries.InputBuffer = warehouse.Factory.NewQuery()
	Queries.InputBuffer.And(input.Components.InputBuffer)

	Queries.Position = warehouse.Factory.NewQuery()
	Queries.Position.And(spatial.Components.Position)

	Queries.Dynamics = warehouse.Factory.NewQuery()
	Queries.Dynamics.And(motion.Components.Dynamics)

	Queries.Shape = warehouse.Factory.NewQuery()
	Queries.Shape.And(spatial.Components.Shape)

	Queries.SpriteBundle = warehouse.Factory.NewQuery()
	Queries.SpriteBundle.And(client.Components.SpriteBundle)

	Queries.SoundBundle = warehouse.Factory.NewQuery()
	Queries.SoundBundle.And(client.Components.SoundBundle)
	return nil
}()
