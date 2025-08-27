package client

import (
	"embed"
	"encoding/json"
	"log"
	"os"

	"github.com/TheBitDrifter/bappa/blueprint/vector"
	"github.com/TheBitDrifter/bappa/environment"
)

type AnimationCollection struct {
	Animations []AnimationData `json:"animations"`
}

// AnimationData contains configuration for sprite-based animations
type AnimationData struct {
	// Name is the animation name
	Name string `json:"name"`

	// PositionOffset represents the sprite's displacement from the entity's position
	PositionOffset vector.Two `json:"offset"`

	// RowIndex indicates which row in the sprite sheet to use
	RowIndex int `json:"rowIndex"`

	// FrameWidth specifies the width of each animation frame in pixels
	FrameWidth int `json:"frameWidth"`

	// FrameHeight specifies the height of each animation frame in pixels
	FrameHeight int `json:"frameHeight"`

	// FrameCount defines how many frames are in this animation
	FrameCount int `json:"frameCount"`

	// Speed controls how quickly the animation plays (how many ticks per frame)
	Speed int `json:"speed"`

	// StartTick defines when the animation begins
	StartTick int `json:"-"`

	// Freeze indicates whether the animation should stay on the last frame once finished
	Freeze bool `json:"freeze"`
}

// Only works with freeze true
func (a AnimationData) IsFinished(currentTick int) bool {
	duration := a.FrameCount * a.Speed

	if duration <= 0 {
		return true
	}

	return currentTick >= a.StartTick+duration
}

func LoadAnimationsFromJSON(filename, path string, fs embed.FS) (*AnimationCollection, error) {
	collection := &AnimationCollection{}
	if environment.IsProd() || environment.IsWASM() {
		jsonDataBytes, err := fs.ReadFile(filename)
		if err != nil {
			return nil, err
		}

		err = json.Unmarshal(jsonDataBytes, collection)
		if err != nil {
			return nil, err
		}
		return collection, nil
	}

	jsonDataBytes, err := os.ReadFile(path + filename)
	if err != nil {
		log.Fatalf("Error reading JSON file '%s': %v", path, err)
	}

	err = json.Unmarshal(jsonDataBytes, collection)
	if err != nil {
		log.Fatal(err)
	}

	return collection, nil
}
