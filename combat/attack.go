package combat

import (
	"encoding/json"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/TheBitDrifter/bappa/blueprint/vector"
	"github.com/TheBitDrifter/bappa/environment"
	"github.com/TheBitDrifter/bappa/tteokbokki/spatial"
)

const MaxHitBox = 15

var (
	ATTACK_REPOSITORY = map[string]int{}
	prevAtkID         = 0
)

type AttackCollection struct {
	Attacks []Attack `json:"attacks"`
}

type Attack struct {
	Name                 string
	ID                   int
	Boxes                [MaxHitBox]HitBoxes `json:"boxes"`
	BoxesPositionOffsets [MaxHitBox][MaxHitBox]vector.Two
	Length               int
	StartTick            int
	Speed                int
	Damage               int
	LRDirection          spatial.Direction
	LastHitTick          int
}

type (
	HitBoxes [MaxHitBox]HitBox
	HitBox   spatial.Shape
)

type HitBoxPrimitive struct {
	X int `json:"x"`
	Y int `json:"y"`
	W int `json:"w"`
	H int `json:"h"`
}

func NewAttack(name string) Attack {
	id, ok := ATTACK_REPOSITORY[name]
	if !ok {
		prevAtkID++
		ATTACK_REPOSITORY[name] = prevAtkID
		return Attack{ID: prevAtkID, Name: name}
	} else {
		return Attack{ID: id, Name: name}
	}
}

func (a *Attack) FirstActiveBoxIndex() int {
	for i, b := range a.Boxes {
		for _, bb := range b {
			if bb.LocalAAB.Height != 0 && bb.LocalAAB.Width != 0 {
				return i
			}
		}
	}
	return 0
}

func (a *Attack) UnmarshalJSON(data []byte) error {
	type AttackAlias Attack

	aux := &struct {
		Name      string              `json:"name"`
		ID        int                 `json:"id"`
		Boxes     [][]HitBoxPrimitive `json:"boxes"`
		Length    int                 `json:"length"`
		StartTick int                 `json:"startTick"`
		Speed     int                 `json:"speed"`
		Damage    int                 `json:"damage"`
	}{}

	if err := json.Unmarshal(data, aux); err != nil {
		return fmt.Errorf("error unmarshaling attack data: %w", err)
	}

	a.Name = aux.Name
	a.ID = aux.ID
	a.Length = aux.Length
	a.StartTick = aux.StartTick
	a.Speed = aux.Speed
	a.Damage = aux.Damage

	if len(aux.Boxes) > MaxHitBox {
		return fmt.Errorf("too many outer box arrays (frames) for attack '%s': got %d, max %d", a.Name, len(aux.Boxes), MaxHitBox)
	}

	for i, primHitBoxesInFrame := range aux.Boxes {
		if len(primHitBoxesInFrame) > MaxHitBox {
			return fmt.Errorf("too many inner hitboxes in frame %d for attack '%s': got %d, max %d", i, a.Name, len(primHitBoxesInFrame), MaxHitBox)
		}
		for j, prim := range primHitBoxesInFrame {
			a.Boxes[i][j] = HitBox(spatial.NewRectangle(float64(prim.W), float64(prim.H)))
			a.BoxesPositionOffsets[i][j].X = float64(prim.X)
			a.BoxesPositionOffsets[i][j].Y = float64(prim.Y)
		}
	}

	return nil
}

func NewAttacksFromJSON(embeddedFS fs.FS, fileName string, path string) ([]Attack, error) {
	var fileData []byte
	var err error

	if environment.IsProd() || environment.IsWASM() {
		// In Wasm, we read from the embedded filesystem.
		// The path must be relative to the embed root.
		fileData, err = fs.ReadFile(embeddedFS, fileName)
		if err != nil {
			return nil, err
		}

	} else {

		// In a native build, we read from the OS filesystem.
		// The path is relative to the executable's working directory.
		nativePath := filepath.Join(path, fileName)
		fileData, err = os.ReadFile(nativePath)
		if err != nil {
			return nil, err
		}
	}

	collection := &AttackCollection{}
	err = json.Unmarshal(fileData, collection)
	if err != nil {
		return nil, err
	}

	res := []Attack{}
	for _, atk := range collection.Attacks {
		if atk.Name == "" {
			continue
		}
		id, ok := ATTACK_REPOSITORY[atk.Name]
		if !ok {
			prevAtkID++
			ATTACK_REPOSITORY[atk.Name] = prevAtkID
			atk.ID = prevAtkID
			res = append(res, atk)
		} else {
			atk.ID = id
			res = append(res, atk)
		}

	}

	return res, nil
}
