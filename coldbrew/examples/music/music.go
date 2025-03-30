package main

import (
	"embed"
	"log"

	"github.com/TheBitDrifter/blueprint"
	blueprintclient "github.com/TheBitDrifter/blueprint/client"

	"github.com/TheBitDrifter/blueprint/vector"
	"github.com/TheBitDrifter/coldbrew"
	coldbrew_rendersystems "github.com/TheBitDrifter/coldbrew/rendersystems"
	"github.com/TheBitDrifter/warehouse"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/hajimehoshi/ebiten/v2/text/v2"
	"golang.org/x/image/font/basicfont"
)

//go:embed assets/*
var assets embed.FS

func main() {
	client := coldbrew.NewClient(
		640,
		360,
		10,
		10,
		10,
		assets,
	)

	client.SetTitle("Playing Music")

	err := client.RegisterScene(
		"Example Scene",
		640,
		360,
		exampleScenePlan,
		[]coldbrew.RenderSystem{instructions{}},
		[]coldbrew.ClientSystem{
			&musicSystem{},
		},
		[]blueprint.CoreSystem{},
	)
	if err != nil {
		log.Fatal(err)
	}

	client.RegisterGlobalRenderSystem(coldbrew_rendersystems.GlobalRenderer{})
	client.ActivateCamera()

	if err := client.Start(); err != nil {
		log.Fatal(err)
	}
}

var musicSoundConfig = blueprintclient.SoundConfig{
	Path:             "music.wav",
	AudioPlayerCount: 1,
}

func exampleScenePlan(height, width int, sto warehouse.Storage) error {
	spriteArchetype, err := sto.NewOrExistingArchetype(
		blueprintclient.Components.SoundBundle,
	)
	if err != nil {
		return err
	}

	err = spriteArchetype.Generate(1,
		blueprintclient.NewSoundBundle().AddSoundFromConfig(musicSoundConfig),
	)
	if err != nil {
		return err
	}
	return nil
}

type musicSystem struct {
	volume float64
}

func (sys *musicSystem) Run(lc coldbrew.LocalClient, scene coldbrew.Scene) error {
	if inpututil.IsKeyJustPressed(ebiten.Key1) && sys.volume == 0 {
		sys.volume = 1
	} else if inpututil.IsKeyJustPressed(ebiten.Key1) && sys.volume == 1 {
		sys.volume = 0
	}

	musicQuery := warehouse.Factory.NewQuery().And(blueprintclient.Components.SoundBundle)
	cursor := scene.NewCursor(musicQuery)

	for range cursor.Next() {
		soundBundle := blueprintclient.Components.SoundBundle.GetFromCursor(cursor)

		sound, _ := coldbrew.MaterializeSound(soundBundle, musicSoundConfig)
		player := sound.GetAny()
		player.SetVolume(sys.volume)

		if !player.IsPlaying() {
			player.Rewind()
			player.Play()
		}
	}
	return nil
}

type instructions struct{}

func (instructions) Render(scene coldbrew.Scene, screen coldbrew.Screen, cu coldbrew.CameraUtility) {
	cam := cu.ActiveCamerasFor(scene)[0]
	instructionText := "Press 1 to toggle music!"
	textFace := text.NewGoXFace(basicfont.Face7x13)
	cam.DrawTextBasicStatic(instructionText, &text.DrawOptions{}, textFace, vector.Two{
		X: 230,
		Y: 160,
	})
	cam.PresentToScreen(screen, 0)
}
