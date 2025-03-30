package client

type Tile struct {
	SourceX  int     // X coordinate in the tileset (in tiles, not pixels)
	SourceY  int     // Y coordinate in the tileset (in tiles, not pixels)
	TileID   int     // Tile ID in the tileset
	FlippedX bool    // Whether the tile is horizontally flipped
	FlippedY bool    // Whether the tile is vertically flipped
	X        float64 // X position of the tile in the world
	Y        float64 // Y position of the tile in the world
}
