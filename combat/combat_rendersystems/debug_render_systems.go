package combat_rendersystems

import (
	"image/color"
	"log"

	"github.com/TheBitDrifter/bappa/coldbrew"
	"github.com/TheBitDrifter/bappa/combat"
	"github.com/TheBitDrifter/bappa/tteokbokki/spatial"

	"github.com/TheBitDrifter/bappa/warehouse"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

type HitBoxRenderSystem struct{}

func (HitBoxRenderSystem) Render(scene coldbrew.Scene, screen coldbrew.Screen, cli coldbrew.LocalClient) {
	if !coldbrew.ClientConfig.DebugVisual {
		return
	}
	query := warehouse.Factory.NewQuery().And(combat.Components.Attack, spatial.Components.Position)
	cursor := scene.NewCursor(query)
	tick := scene.CurrentTick()
	baseColor := color.RGBA{255, 255, 0, 255}

	for _, cam := range cli.ActiveCamerasFor(scene) {
		if !cli.Ready(cam) {
			continue
		}
		for range cursor.Next() {
			attack := combat.Components.Attack.GetFromCursor(cursor)

			if attack.Length == 0 {
				log.Println(attack.Name, "has no length for hitboxes")
				continue
			}
			index := ((tick - attack.StartTick) / attack.Speed) % attack.Length

			boxes := attack.Boxes[index]

			for k, b := range boxes {
				if b.LocalAAB.Height == 0 {
					continue
				}
				halfWidth := float32(b.WorldAAB.Width / 2)
				halfHeight := float32(b.WorldAAB.Height / 2)

				pos := spatial.Components.Position.GetFromCursor(cursor)

				_, local := cam.Positions()
				x := float32(pos.X - local.X + attack.BoxesPositionOffsets[index][k].X)
				y := float32(pos.Y - local.Y + +attack.BoxesPositionOffsets[index][k].Y)

				dir, hasDir := spatial.Components.Direction.GetFromCursorSafe(cursor)
				if attack.LRDirection.Valid() {
					dir = &attack.LRDirection
					hasDir = true
				}

				if hasDir && dir.IsLeft() {
					x = float32(pos.X - local.X - attack.BoxesPositionOffsets[index][k].X)
				}

				// Top line
				vector.StrokeLine(cam.Surface(), x-halfWidth, y-halfHeight, x+halfWidth, y-halfHeight, 1, baseColor, false)
				// Right line
				vector.StrokeLine(cam.Surface(), x+halfWidth, y-halfHeight, x+halfWidth, y+halfHeight, 1, baseColor, false)
				// Bottom line
				vector.StrokeLine(cam.Surface(), x+halfWidth, y+halfHeight, x-halfWidth, y+halfHeight, 1, baseColor, false)
				// Left line
				vector.StrokeLine(cam.Surface(), x-halfWidth, y+halfHeight, x-halfWidth, y-halfHeight, 1, baseColor, false)

			}

		}

		cam.PresentToScreen(screen, coldbrew.ClientConfig.CameraBorderSize())

	}
}

type HurtBoxRenderSystem struct{}

func (HurtBoxRenderSystem) Render(scene coldbrew.Scene, screen coldbrew.Screen, cli coldbrew.LocalClient) {
	if !coldbrew.ClientConfig.DebugVisual {
		return
	}
	query := warehouse.Factory.NewQuery().And(combat.Components.HurtBox, spatial.Components.Position)
	cursor := scene.NewCursor(query)
	baseColor := color.RGBA{128, 180, 40, 255}

	for _, cam := range cli.ActiveCamerasFor(scene) {
		if !cli.Ready(cam) {
			continue
		}
		for range cursor.Next() {

			b := combat.Components.HurtBox.GetFromCursor(cursor)

			halfWidth := float32(b.WorldAAB.Width / 2)
			halfHeight := float32(b.WorldAAB.Height / 2)

			pos := spatial.Components.Position.GetFromCursor(cursor)

			_, local := cam.Positions()
			x := float32(pos.X - local.X)
			y := float32(pos.Y - local.Y)

			// Top line
			vector.StrokeLine(cam.Surface(), x-halfWidth, y-halfHeight, x+halfWidth, y-halfHeight, 1, baseColor, false)
			// Right line
			vector.StrokeLine(cam.Surface(), x+halfWidth, y-halfHeight, x+halfWidth, y+halfHeight, 1, baseColor, false)
			// Bottom line
			vector.StrokeLine(cam.Surface(), x+halfWidth, y+halfHeight, x-halfWidth, y+halfHeight, 1, baseColor, false)
			// Left line
			vector.StrokeLine(cam.Surface(), x-halfWidth, y+halfHeight, x-halfWidth, y-halfHeight, 1, baseColor, false)

		}

		query = warehouse.Factory.NewQuery().And(combat.Components.HurtBoxes, spatial.Components.Position)
		cursor = scene.NewCursor(query)

		for range cursor.Next() {

			hurtBoxes := combat.Components.HurtBoxes.GetFromCursor(cursor)
			for b := range hurtBoxes.Active() {
				halfWidth := float32(b.WorldAAB.Width / 2)
				halfHeight := float32(b.WorldAAB.Height / 2)

				direction, directionOK := spatial.Components.Direction.GetFromCursorSafe(cursor)
				adj := b.RelativePos
				if directionOK {
					adj.X = adj.X * direction.AsFloat()
				}

				p := spatial.Components.Position.GetFromCursor(cursor)
				pos := p.Add(adj)

				_, local := cam.Positions()
				x := float32(pos.X - local.X)
				y := float32(pos.Y - local.Y)

				// Top line
				vector.StrokeLine(cam.Surface(), x-halfWidth, y-halfHeight, x+halfWidth, y-halfHeight, 1, baseColor, false)
				// Right line
				vector.StrokeLine(cam.Surface(), x+halfWidth, y-halfHeight, x+halfWidth, y+halfHeight, 1, baseColor, false)
				// Bottom line
				vector.StrokeLine(cam.Surface(), x+halfWidth, y+halfHeight, x-halfWidth, y+halfHeight, 1, baseColor, false)
				// Left line
				vector.StrokeLine(cam.Surface(), x-halfWidth, y+halfHeight, x-halfWidth, y-halfHeight, 1, baseColor, false)

			}
		}

		cam.PresentToScreen(screen, coldbrew.ClientConfig.CameraBorderSize())
	}
}
