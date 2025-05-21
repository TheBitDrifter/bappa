package ldtk

import (
	"log"

	"github.com/TheBitDrifter/bappa/tteokbokki/spatial"
	"github.com/TheBitDrifter/bappa/warehouse"
)

// Rectangle represents a merged rectangular area for IntGrid optimization
type Rectangle struct {
	X      float64
	Y      float64
	Width  float64
	Height float64
}

// IntGridLayerConfig holds the component composition and default values for an IntGrid layer
type IntGridLayerConfig struct {
	Composition   []warehouse.Component
	DefaultValues []interface{}
	Padding       float64
}

// LoadIntGrid loads collision entities based on IntGrid values
// It takes a map where the key is the IntGrid value and the value is the configuration for that layer
func (p *LDtkProject) LoadIntGridFromConfig(levelName string, sto warehouse.Storage, intGridLayerConfigs map[int]IntGridLayerConfig) error {
	level, exists := p.parsedLevels[levelName]
	if !exists {
		log.Printf("Level '%s' not found", levelName)
		return nil
	}

	// Process each IntGrid layer
	for layerID, grid := range level.IntGridRawData {
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

		for gridValue, config := range intGridLayerConfigs {
			archetype, err := sto.NewOrExistingArchetype(config.Composition...)
			if err != nil {
				return err
			}

			rectangles := mergeRectangles(grid, gridValue, cellSize, config.Padding)

			// Create entities for the merged rectangles
			for _, rect := range rectangles {
				centerX := rect.X + rect.Width/2
				centerY := rect.Y + rect.Height/2

				// Combine default values with calculated spatial components
				componentsToGenerate := []any{
					spatial.NewPosition(centerX, centerY),
					spatial.NewRectangle(rect.Width, rect.Height),
				}
				componentsToGenerate = append(componentsToGenerate, config.DefaultValues...)

				// Generate the entity
				err := archetype.Generate(1, componentsToGenerate...)
				if err != nil {
					return err
				}
			}
		}
	}

	return nil
}

func (p *LDtkProject) LoadIntGrid(levelName string, sto warehouse.Storage, archetypes ...warehouse.Archetype) error {
	level, exists := p.parsedLevels[levelName]

	if !exists {
		log.Printf("Level '%s' not found", levelName)
		return nil

	}

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
			rectangles := mergeRectangles(grid, gridValue, cellSize, 0)

			// Create entities for the merged rectangles
			for _, rect := range rectangles {

				// Calculate center position
				centerX := rect.X + rect.Width/2
				centerY := rect.Y + rect.Height/2
				err := archetype.Generate(1,
					spatial.NewPosition(centerX, centerY),
					spatial.NewRectangle(rect.Width, rect.Height),
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
func mergeRectangles(grid [][]int, gridValue, cellSize int, padding float64) []Rectangle {
	var rectangles []Rectangle

	height := len(grid)
	if height == 0 {
		return rectangles
	}
	width := len(grid[0])

	visited := make([][]bool, height)
	for i := range visited {
		visited[i] = make([]bool, width)
	}

	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			if grid[y][x] != gridValue || visited[y][x] {
				continue
			}

			rectWidth, rectHeight := 1, 1

			for x+rectWidth < width && grid[y][x+rectWidth] == gridValue && !visited[y][x+rectWidth] {
				rectWidth++
			}

			canExpandVertically := true
			for canExpandVertically && y+rectHeight < height {
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

			for dy := 0; dy < rectHeight; dy++ {
				for dx := 0; dx < rectWidth; dx++ {
					visited[y+dy][x+dx] = true
				}
			}

			// Apply padding to the final rectangle dimensions
			paddedWidth := float64(rectWidth*cellSize) - (padding * 2)
			if paddedWidth < 0 {
				paddedWidth = 0
			}

			paddedHeight := float64(rectHeight*cellSize) - (padding * 2)
			if paddedHeight < 0 {
				paddedHeight = 0
			}

			rectangle := Rectangle{
				X:      float64(x*cellSize) + padding,
				Y:      float64(y*cellSize) + padding,
				Width:  paddedWidth,
				Height: paddedHeight,
			}

			rectangles = append(rectangles, rectangle)
		}
	}

	return rectangles
}
