package ldtk

import (
	"log"

	blueprintspatial "github.com/TheBitDrifter/blueprint/spatial"
	"github.com/TheBitDrifter/warehouse"
)

// Rectangle represents a merged rectangular area for IntGrid optimization
type Rectangle struct {
	X      float64
	Y      float64
	Width  float64
	Height float64
}

// LoadIntGrid loads collision entities based on IntGrid values
// Takes archetypes where index+1 corresponds to the IntGrid value
func (p *LDtkProject) LoadIntGrid(levelName string, sto warehouse.Storage, archetypes ...warehouse.Archetype) error {
	level, exists := p.parsedLevels[levelName]
	if !exists {
		log.Printf("Level '%s' not found", levelName)
		return nil
	}

	// Process each IntGrid layer
	for layerID, grid := range level.IntGridRawData {
		// Find layer info to get cell size
		var cellSize int
		for _, layer := range level.LayerInstances {
			if layer.Identifier == layerID {
				cellSize = layer.GridSize
				break
			}
		}

		if cellSize == 0 {
			log.Printf("Couldn't find grid size for layer '%s'", layerID)
			continue
		}

		// Process each grid value type that we have an archetype for
		for gridValue := 1; gridValue <= len(archetypes); gridValue++ {
			archetypeIndex := gridValue - 1
			if archetypeIndex >= len(archetypes) {
				continue
			}

			archetype := archetypes[archetypeIndex]

			// Find optimized rectangles for this grid value
			rectangles := mergeRectangles(grid, gridValue, cellSize)

			// Create entities for the merged rectangles
			for _, rect := range rectangles {
				// Calculate center position
				centerX := rect.X + rect.Width/2
				centerY := rect.Y + rect.Height/2

				err := archetype.Generate(1,
					blueprintspatial.NewPosition(centerX, centerY),
					blueprintspatial.NewRectangle(rect.Width, rect.Height),
				)
				if err != nil {
					return err
				}
			}
		}
	}

	return nil
}

// mergeRectangles finds optimized rectangles for a specific grid value
func mergeRectangles(grid [][]int, gridValue, cellSize int) []Rectangle {
	var rectangles []Rectangle

	// Keep track of which cells we've already processed
	height := len(grid)
	if height == 0 {
		return rectangles
	}
	width := len(grid[0])

	visited := make([][]bool, height)
	for i := range visited {
		visited[i] = make([]bool, width)
	}

	// Scan the grid
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			// Skip if cell doesn't match or is already processed
			if grid[y][x] != gridValue || visited[y][x] {
				continue
			}

			// Find largest rectangle starting at this cell
			rectWidth, rectHeight := 1, 1

			// Expand width (horizontal merging)
			for x+rectWidth < width && grid[y][x+rectWidth] == gridValue && !visited[y][x+rectWidth] {
				rectWidth++
			}

			// Try to expand height (vertical merging)
			canExpandVertically := true
			for canExpandVertically && y+rectHeight < height {
				// Check if the next row can be merged
				for i := 0; i < rectWidth; i++ {
					if x+i >= width || grid[y+rectHeight][x+i] != gridValue || visited[y+rectHeight][x+i] {
						canExpandVertically = false
						break
					}
				}

				if canExpandVertically {
					rectHeight++
				}
			}

			// Mark all cells in this rectangle as visited
			for dy := 0; dy < rectHeight; dy++ {
				for dx := 0; dx < rectWidth; dx++ {
					visited[y+dy][x+dx] = true
				}
			}

			// Create a rectangle
			rectangle := Rectangle{
				X:      float64(x * cellSize),
				Y:      float64(y * cellSize),
				Width:  float64(rectWidth * cellSize),
				Height: float64(rectHeight * cellSize),
			}

			rectangles = append(rectangles, rectangle)
		}
	}

	return rectangles
}
