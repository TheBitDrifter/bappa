package coldbrew

import (
	"fmt"
	"image"
	"sync"

	"github.com/hajimehoshi/ebiten/v2"
)

// Sprite manages image state and rendering properties
// It contains a reference to the drawable *ebiten.Image
type Sprite interface {
	// Name returns the sprite's identifier
	Name() string
	// Image returns the underlying ebiten.Image
	Image() *ebiten.Image
	// Draw renders the sprite to the target image with the provided options
	Draw(*ebiten.Image, *ebiten.DrawImageOptions)
	// GetFrame retrieves a specific frame from a sprite sheet
	GetFrame(rowIndex, frameIndex, frameWidth, frameHeight int) *ebiten.Image
	// GetTile retrieves a specific tile from a tileset
	GetTile(tileX, tileY, tileWidth, tileHeight int) *ebiten.Image
}

// sprite implements the Sprite interface
type sprite struct {
	name        string
	image       *ebiten.Image
	frames      map[string]*ebiten.Image // Cache for individual animation frames
	tilesMatrix map[string]*ebiten.Image // Cache for individual tileset tiles
	mutex       sync.RWMutex             // Protects the caches
}

func (s *sprite) Name() string {
	return s.name
}

func (s *sprite) Image() *ebiten.Image {
	return s.image
}

func (s *sprite) Draw(target *ebiten.Image, opts *ebiten.DrawImageOptions) {
	target.DrawImage(s.image, opts)
}

// generateFrameKey creates a unique key for an animation frame
func generateFrameKey(rowIndex, frameIndex, frameWidth, frameHeight int) string {
	return fmt.Sprintf("r%d_f%d_w%d_h%d", rowIndex, frameIndex, frameWidth, frameHeight)
}

// generateTileKey creates a unique key for a tileset tile
func generateTileKey(tileX, tileY, tileWidth, tileHeight int) string {
	return fmt.Sprintf("tx%d_ty%d_tw%d_th%d", tileX, tileY, tileWidth, tileHeight)
}

// GetFrame retrieves a frame from the sprite sheet, using the cache if available
func (s *sprite) GetFrame(rowIndex, frameIndex, frameWidth, frameHeight int) *ebiten.Image {
	frameKey := generateFrameKey(rowIndex, frameIndex, frameWidth, frameHeight)

	// Try to get from cache first (with read lock)
	s.mutex.RLock()
	if frame, ok := s.frames[frameKey]; ok {
		s.mutex.RUnlock()
		return frame
	}
	s.mutex.RUnlock()

	// Not in cache, extract the frame (with write lock)
	s.mutex.Lock()
	defer s.mutex.Unlock()

	// Check again after acquiring write lock
	if frame, ok := s.frames[frameKey]; ok {
		return frame
	}

	// Create the frame
	sx := frameIndex * frameWidth
	sy := rowIndex * frameHeight
	frame := s.image.SubImage(image.Rect(sx, sy, sx+frameWidth, sy+frameHeight)).(*ebiten.Image)

	// Initialize the frames map if needed
	if s.frames == nil {
		s.frames = make(map[string]*ebiten.Image)
	}

	// Cache the frame
	s.frames[frameKey] = frame
	return frame
}

// GetTile retrieves a tile from the tileset, using the cache if available
func (s *sprite) GetTile(tileX, tileY, tileWidth, tileHeight int) *ebiten.Image {
	tileKey := generateTileKey(tileX, tileY, tileWidth, tileHeight)

	// Try to get from cache first (with read lock)
	s.mutex.RLock()
	if tile, ok := s.tilesMatrix[tileKey]; ok {
		s.mutex.RUnlock()
		return tile
	}
	s.mutex.RUnlock()

	// Not in cache, extract the tile (with write lock)
	s.mutex.Lock()
	defer s.mutex.Unlock()

	// Check again after acquiring write lock
	if tile, ok := s.tilesMatrix[tileKey]; ok {
		return tile
	}

	// Create the tile
	sx := tileX * tileWidth
	sy := tileY * tileHeight
	tile := s.image.SubImage(image.Rect(sx, sy, sx+tileWidth, sy+tileHeight)).(*ebiten.Image)

	// Initialize the tilesMatrix map if needed
	if s.tilesMatrix == nil {
		s.tilesMatrix = make(map[string]*ebiten.Image)
	}

	// Cache the tile
	s.tilesMatrix[tileKey] = tile
	return tile
}

// NewSprite creates a new sprite instance with frame and tile caching capability
func NewSprite(name string, image *ebiten.Image) Sprite {
	return &sprite{
		name:        name,
		image:       image,
		frames:      make(map[string]*ebiten.Image),
		tilesMatrix: make(map[string]*ebiten.Image),
	}
}
