package rendersystems

import (
	"image/color"
	"math"

	"github.com/TheBitDrifter/blueprint"
	blueprintspatial "github.com/TheBitDrifter/blueprint/spatial"
	"github.com/TheBitDrifter/blueprint/vector"
	"github.com/TheBitDrifter/coldbrew"
	"github.com/TheBitDrifter/tteokbokki/spatial"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
)

// DebugRenderer visualizes physics and collisions for debugging purposes
// It operates independently from actual collision processing systems
// A messy expensive system...a temporary solution for now
type DebugRenderer struct{}

// ShapeInfo stores the shape and position data for rendering
type ShapeInfo struct {
	Shape    blueprintspatial.Shape
	Position vector.Two
}

// Render draws debug information when debug visualization is enabled
// Displays shapes and their collisions for each active camera
func (sys *DebugRenderer) Render(cli coldbrew.Client, screen coldbrew.Screen) {
	if !coldbrew.ClientConfig.DebugVisual || coldbrew.IsProd() {
		return
	}
	for _, cam := range cli.Cameras() {
		if !cam.Active() {
			continue
		}
		scene := cli.CameraSceneTracker()[cam].Scene

		if scene == nil || !scene.Ready() {
			continue
		}

		tracker := cli.CameraSceneTracker()[cam]
		currTick := cli.CurrentTick()
		lastChanged := tracker.Tick
		minLoadTime := coldbrew.ClientConfig.MinimumLoadTime()
		sceneRecentlyChanged := currTick-lastChanged < minLoadTime
		alwaysShowLoader := minLoadTime != 0
		if sceneRecentlyChanged && alwaysShowLoader {
			scene = cli.LoadingScenes()[0]
		}
		shapes := make([]ShapeInfo, 0)
		shapeCursor := scene.NewCursor(blueprint.Queries.Shape)
		for range shapeCursor.Next() {
			shape := blueprintspatial.Components.Shape.GetFromCursor(shapeCursor)
			pos := blueprintspatial.Components.Position.GetFromCursor(shapeCursor)
			if shape != nil && pos != nil {
				shapes = append(shapes, ShapeInfo{
					Shape:    *shape,
					Position: vector.Two{X: pos.X, Y: pos.Y},
				})
			}
		}

		_, localPos := cam.Positions()
		sys.renderShapes(cam.Surface(), shapes, *localPos)
		sys.renderCollisions(cam.Surface(), shapes, *localPos)
		cam.PresentToScreen(screen, coldbrew.ClientConfig.CameraBorderSize())

	}
}

// renderShapes draws all shape outlines with appropriate colors
// Green for normal shapes and semi-transparent green for skin shapes
func (sys *DebugRenderer) renderShapes(camSurface *ebiten.Image, shapes []ShapeInfo, camPos vector.Two) {
	baseColor := color.RGBA{0, 255, 0, 255}
	skinColor := color.RGBA{0, 255, 0, 128} // Semi-transparent for skins

	for _, shape := range shapes {
		// Render polygon vertices and edges
		if len(shape.Shape.Polygon.WorldVertices) > 0 {
			verts := shape.Shape.Polygon.WorldVertices
			for j := 0; j < len(verts); j++ {
				next := (j + 1) % len(verts)
				x1 := verts[j].X - camPos.X
				y1 := verts[j].Y - camPos.Y
				x2 := verts[next].X - camPos.X
				y2 := verts[next].Y - camPos.Y
				ebitenutil.DrawLine(camSurface, x1, y1, x2, y2, baseColor)
				ebitenutil.DrawCircle(camSurface, x1, y1, 2, color.RGBA{255, 0, 0, 255})
			}
		}

		// Render AAB
		if shape.Shape.WorldAAB.Height != 0 {
			halfWidth := shape.Shape.WorldAAB.Width / 2
			halfHeight := shape.Shape.WorldAAB.Height / 2
			x := shape.Position.X - camPos.X
			y := shape.Position.Y - camPos.Y
			ebitenutil.DrawLine(camSurface, x-halfWidth, y-halfHeight, x+halfWidth, y-halfHeight, baseColor)
			ebitenutil.DrawLine(camSurface, x+halfWidth, y-halfHeight, x+halfWidth, y+halfHeight, baseColor)
			ebitenutil.DrawLine(camSurface, x+halfWidth, y+halfHeight, x-halfWidth, y+halfHeight, baseColor)
			ebitenutil.DrawLine(camSurface, x-halfWidth, y+halfHeight, x-halfWidth, y-halfHeight, baseColor)
		}

		// Render Skin AAB
		if shape.Shape.Skin.AAB.Height != 0 {
			halfWidth := shape.Shape.Skin.AAB.Width / 2
			halfHeight := shape.Shape.Skin.AAB.Height / 2
			x := shape.Position.X - camPos.X
			y := shape.Position.Y - camPos.Y
			ebitenutil.DrawLine(camSurface, x-halfWidth, y-halfHeight, x+halfWidth, y-halfHeight, skinColor)
			ebitenutil.DrawLine(camSurface, x+halfWidth, y-halfHeight, x+halfWidth, y+halfHeight, skinColor)
			ebitenutil.DrawLine(camSurface, x+halfWidth, y+halfHeight, x-halfWidth, y+halfHeight, skinColor)
			ebitenutil.DrawLine(camSurface, x-halfWidth, y+halfHeight, x-halfWidth, y-halfHeight, skinColor)
		}

		// Render Skin Circle
		if shape.Shape.Skin.Circle.Radius != 0 {
			x := shape.Position.X - camPos.X
			y := shape.Position.Y - camPos.Y
			radius := shape.Shape.Skin.Circle.Radius

			// Draw the circle outline
			const segments = 32
			for i := 0; i < segments; i++ {
				angle1 := 2 * math.Pi * float64(i) / segments
				angle2 := 2 * math.Pi * float64(i+1) / segments

				x1 := x + radius*math.Cos(angle1)
				y1 := y + radius*math.Sin(angle1)
				x2 := x + radius*math.Cos(angle2)
				y2 := y + radius*math.Sin(angle2)

				ebitenutil.DrawLine(camSurface, x1, y1, x2, y2, skinColor)
			}

			// Draw center point
			ebitenutil.DrawCircle(camSurface, x, y, 2, skinColor)
		}
	}
}

// renderCollisions detects and visualizes collisions between shapes
// Draws connecting lines between colliding shapes and highlights them in red
func (sys *DebugRenderer) renderCollisions(screen *ebiten.Image, shapes []ShapeInfo, camPos vector.Two) {
	for i := 0; i < len(shapes); i++ {
		for j := i + 1; j < len(shapes); j++ {
			// TODO: Probably should render the collision itself
			hasCollision, _ := spatial.Detector.Check(
				shapes[i].Shape,
				shapes[j].Shape,
				&shapes[i].Position,
				&shapes[j].Position,
			)

			if hasCollision {
				// Draw collision line
				ebitenutil.DrawLine(
					screen,
					shapes[i].Position.X-camPos.X,
					shapes[i].Position.Y-camPos.Y,
					shapes[j].Position.X-camPos.X,
					shapes[j].Position.Y-camPos.Y,
					color.RGBA{255, 255, 255, 128},
				)

				// Draw colliding shapes in red
				drawShapeWithColor(screen, shapes[i], camPos, color.RGBA{255, 0, 0, 255})
				drawShapeWithColor(screen, shapes[j], camPos, color.RGBA{255, 0, 0, 255})
			}
		}
	}
}

// drawShapeWithColor renders a shape with the specified color
// Used to highlight colliding shapes
func drawShapeWithColor(screen *ebiten.Image, shape ShapeInfo, camPos vector.Two, color color.RGBA) {
	if len(shape.Shape.Polygon.WorldVertices) > 0 {
		verts := shape.Shape.Polygon.WorldVertices
		for j := 0; j < len(verts); j++ {
			next := (j + 1) % len(verts)
			x1 := verts[j].X - camPos.X
			y1 := verts[j].Y - camPos.Y
			x2 := verts[next].X - camPos.X
			y2 := verts[next].Y - camPos.Y
			ebitenutil.DrawLine(screen, x1, y1, x2, y2, color)
		}
	}

	if shape.Shape.WorldAAB.Height != 0 {
		halfWidth := shape.Shape.WorldAAB.Width / 2
		halfHeight := shape.Shape.WorldAAB.Height / 2
		x := shape.Position.X - camPos.X
		y := shape.Position.Y - camPos.Y

		ebitenutil.DrawLine(screen, x-halfWidth, y-halfHeight, x+halfWidth, y-halfHeight, color)
		ebitenutil.DrawLine(screen, x+halfWidth, y-halfHeight, x+halfWidth, y+halfHeight, color)
		ebitenutil.DrawLine(screen, x+halfWidth, y+halfHeight, x-halfWidth, y+halfHeight, color)
		ebitenutil.DrawLine(screen, x-halfWidth, y+halfHeight, x-halfWidth, y-halfHeight, color)
	}
}
