package ldtk

import (
	"encoding/json"
	"fmt"
	"io/fs"
	"log"
	"os"
	"runtime"
)

// LDtkProject represents the main structure for an LDtk project
type LDtkProject struct {
	Levels []json.RawMessage `json:"levels"`
	Defs   struct {
		Tilesets []struct {
			Identifier   string `json:"identifier"`
			RelPath      string `json:"relPath"`
			PxWid        int    `json:"pxWid"`
			PxHei        int    `json:"pxHei"`
			TileGridSize int    `json:"tileGridSize"`
			UID          int    `json:"uid"`
		} `json:"tilesets"`
	} `json:"defs"`
	ldtkFS fs.FS

	// Cache for parsed levels
	parsedLevels map[string]parsedLevel
}

// parsedLevel holds the parsed data for a level
type parsedLevel struct {
	Identifier     string                `json:"identifier"`
	PxWid          int                   `json:"pxWid"`
	PxHei          int                   `json:"pxHei"`
	LayerInstances []parsedLayerInstance `json:"layerInstances"`
	EntityRawData  map[string][]LDtkEntityInstance
	IntGridRawData map[string][][]int
	TileRawData    map[string]tileLayerData
	Raw            []byte
}

type parsedLayerInstance struct {
	Identifier      string               `json:"__identifier"`
	Type            string               `json:"__type"`
	CWid            int                  `json:"__cWid"`     // Width in cells
	CHei            int                  `json:"__cHei"`     // Height in cells
	GridSize        int                  `json:"__gridSize"` // Cell size in pixels
	IntGridCSV      []int                `json:"intGridCsv"` // Grid values
	TilesetDefUID   *int                 `json:"__tilesetDefUid"`
	TilesetRelPath  string               `json:"__tilesetRelPath"`
	EntityInstances []LDtkEntityInstance `json:"entityInstances"`
	GridTiles       []struct {
		Src [2]int `json:"src"`
		Px  [2]int `json:"px"`
		T   int    `json:"t"`
		F   int    `json:"f"`
	} `json:"gridTiles"`
}

type tileLayerData struct {
	TilesetDefUID  int
	TilesetRelPath string
	GridTiles      []parsedGridTile
}

type parsedGridTile struct {
	Src [2]int
	Px  [2]int
	T   int
	F   int
}

// Parse loads an LDtk file from the filesystem or embedded assets
func Parse(ldtkFS fs.FS, relPath string) (*LDtkProject, error) {
	var data []byte
	var err error
	isProd := os.Getenv("BAPPA_ENV") == "production"
	if isProd || isWASM() {
		data, err = fs.ReadFile(ldtkFS, "data.ldtk")
		if err != nil {
			log.Printf("Error reading LDtk file from embedded assets: %v", err)
			return nil, err
		}
	} else {
		// Development: load from filesystem
		data, err = os.ReadFile(relPath)
		if err != nil {
			return nil, err
		}
	}

	var project LDtkProject
	err = json.Unmarshal(data, &project)
	if err != nil {
		log.Printf("Error parsing LDtk file: %v", err)
		return nil, err
	}

	project.ldtkFS = ldtkFS
	project.parsedLevels = make(map[string]parsedLevel)

	// Pre-parse all levels
	for _, rawLevel := range project.Levels {
		var levelInfo struct {
			Identifier string `json:"identifier"`
		}
		if err := json.Unmarshal(rawLevel, &levelInfo); err != nil {
			continue
		}

		var parsedLevel parsedLevel
		if err := json.Unmarshal(rawLevel, &parsedLevel); err != nil {
			log.Printf("Error parsing level '%s': %v", levelInfo.Identifier, err)
			continue
		}

		// Store raw data for future use
		parsedLevel.Raw = []byte(rawLevel)

		// Initialize maps for different layer types
		parsedLevel.EntityRawData = make(map[string][]LDtkEntityInstance)
		parsedLevel.IntGridRawData = make(map[string][][]int)
		parsedLevel.TileRawData = make(map[string]tileLayerData)

		// Process each layer by type
		for _, layer := range parsedLevel.LayerInstances {
			switch layer.Type {
			case "Entities":
				parsedLevel.EntityRawData[layer.Identifier] = layer.EntityInstances

			case "IntGrid":
				// Convert 1D IntGridCSV to 2D grid for easier processing
				grid := make([][]int, layer.CHei)
				for y := 0; y < layer.CHei; y++ {
					grid[y] = make([]int, layer.CWid)
					for x := 0; x < layer.CWid; x++ {
						index := y*layer.CWid + x
						if index < len(layer.IntGridCSV) {
							grid[y][x] = layer.IntGridCSV[index]
						}
					}
				}
				parsedLevel.IntGridRawData[layer.Identifier] = grid

			case "Tiles":
				if layer.TilesetDefUID != nil {
					// Convert GridTiles to our internal type
					parsedGridTiles := make([]parsedGridTile, len(layer.GridTiles))
					for i, tile := range layer.GridTiles {
						parsedGridTiles[i] = parsedGridTile{
							Src: tile.Src,
							Px:  tile.Px,
							T:   tile.T,
							F:   tile.F,
						}
					}

					parsedLevel.TileRawData[layer.Identifier] = tileLayerData{
						TilesetDefUID:  *layer.TilesetDefUID,
						TilesetRelPath: layer.TilesetRelPath,
						GridTiles:      parsedGridTiles,
					}
				}
			}
		}

		project.parsedLevels[levelInfo.Identifier] = parsedLevel
	}

	return &project, nil
}

// GetLevelParsed gets the parsed level data by name
func (p *LDtkProject) GetLevelParsed(levelName string) (*parsedLevel, error) {
	level, exists := p.parsedLevels[levelName]
	if !exists {
		return nil, fmt.Errorf("level '%s' not found", levelName)
	}
	return &level, nil
}

// WidthFor returns the width of a specific level in pixels
func (p *LDtkProject) WidthFor(levelName string) int {
	level, exists := p.parsedLevels[levelName]
	if !exists {
		log.Printf("Level '%s' not found", levelName)
		return 0
	}
	return level.PxWid
}

// HeightFor returns the height of a specific level in pixels
func (p *LDtkProject) HeightFor(levelName string) int {
	level, exists := p.parsedLevels[levelName]
	if !exists {
		log.Printf("Level '%s' not found", levelName)
		return 0
	}
	return level.PxHei
}

func isWASM() bool {
	return runtime.GOOS == "js" && runtime.GOARCH == "wasm"
}
