package combat

import (
	"encoding/json"
	"io/fs"
	"iter"
	"log"
	"os"
	"path/filepath"

	"github.com/TheBitDrifter/bappa/blueprint/vector"
	"github.com/TheBitDrifter/bappa/environment"
	"github.com/TheBitDrifter/bappa/tteokbokki/spatial"
)

type HurtBox struct {
	spatial.Shape
	RelativePos vector.Two
}

func NewHurtBox(width, height, relX, relY float64) HurtBox {
	rect := spatial.NewRectangle(width, height)
	return HurtBox{
		Shape:       rect,
		RelativePos: vector.Two{X: relX, Y: relY},
	}
}

type HurtBoxes [10]HurtBox

func (hbs HurtBoxes) Active() iter.Seq[HurtBox] {
	return func(yield func(HurtBox) bool) {
		for _, hb := range hbs {
			if hb.LocalAAB.Width != 0 && hb.LocalAAB.Height != 0 {
				if !yield(hb) {
					return
				}
			}
		}
	}
}

func NewHurtBoxesFromJSON(embeddedFS fs.FS, fileName string, path string) (HurtBoxes, error) {
	var fileData []byte
	var err error

	if environment.IsProd() || environment.IsWASM() {

		fileData, err = fs.ReadFile(embeddedFS, fileName)
		if err != nil {
			return HurtBoxes{}, err
		}

	} else {
		nativePath := filepath.Join(path, fileName)
		fileData, err = os.ReadFile(nativePath)
		if err != nil {
			return HurtBoxes{}, err
		}
	}

	type jsonHurtboxDef struct {
		X float64 `json:"x"`
		Y float64 `json:"y"`
		W float64 `json:"w"`
		H float64 `json:"h"`
	}

	jsonBoxes := []jsonHurtboxDef{}
	err = json.Unmarshal(fileData, &jsonBoxes)
	if err != nil {
		return HurtBoxes{}, err
	}

	var boxes HurtBoxes

	if len(jsonBoxes) > len(boxes) {
		log.Fatalf(
			"Error: JSON defines %d hurtboxes, but the maximum allowed is %d",
			len(jsonBoxes),
			len(boxes),
		)
	}

	for i, boxDef := range jsonBoxes {
		boxes[i] = NewHurtBox(boxDef.W, boxDef.H, boxDef.X, boxDef.Y)
	}
	return boxes, nil
}
