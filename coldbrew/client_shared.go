package coldbrew

import (
	"fmt"

	"github.com/TheBitDrifter/bappa/environment"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
)

func sharedClientUpdate(cli Client, interpolateCoreSystemsForNetworked bool) error {
	clientAsStandard, isStandard := cli.(*clientImpl)
	if isStandard {
		clientAsStandard.toggleDebugView()

		err := clientAsStandard.processNonExecutedPlansForActiveScenes()
		if err != nil {
			return err
		}

		clientAsStandard.findAndLoadMissingAssetsForActiveScenesAsync()
		if isCacheFull.Load() {
			clientAsStandard.resolveCacheForActiveScenes()
		}

		clientAsStandard.captureInputs()
		err = clientAsStandard.run()
		if err != nil {
			return err
		}

	}

	clientAsNetworked, isNetworked := cli.(*networkClientImpl)
	if isNetworked {
		clientAsNetworked.toggleDebugView()

		err := clientAsNetworked.processNonExecutedPlansForActiveScenes()
		if err != nil {
			return err
		}

		clientAsNetworked.findAndLoadMissingAssetsForActiveScenesAsync()
		if isCacheFull.Load() {
			clientAsNetworked.resolveCacheForActiveScenes()
		}

		clientAsNetworked.captureInputs()
		err = clientAsNetworked.run(interpolateCoreSystemsForNetworked)
		if err != nil {
			return err
		}
	}

	tick++
	return nil
}

func sharedDraw(cliFace Client, image *ebiten.Image) {
	cli, _ := cliFace.(*clientImpl)
	cliAsNet, isNet := cliFace.(*networkClientImpl)

	if isNet {
		cli = cliAsNet.clientImpl
	}

	for i := range cli.cameras {
		c := cli.cameras[i]
		c.Surface().Clear()
	}
	screen := Screen{
		sprite{name: "screen", image: image},
	}
	for _, renderSys := range cli.globalRenderers {
		renderSys.Render(cli, screen)
	}

	// Take a snapshot of active scenes for rendering
	for activeScene := range cli.ActiveScenes() {
		renderers := activeScene.Renderers()
		cameraReady := true
		cameras := cli.ActiveCamerasFor(activeScene)
		for _, cam := range cameras {
			if !cam.Ready(cli) {
				cameraReady = false
			}
		}

		if !activeScene.Ready() || !cameraReady {
			if len(cli.loadingScenes) > 0 {
				loadingScene := cli.loadingScenes[0]
				for _, renderSys := range loadingScene.Renderers() {
					if isNet {
						renderSys.Render(activeScene, screen, cliAsNet)
					} else {
						renderSys.Render(activeScene, screen, cli)
					}
				}
			}
		}

		for _, renderSys := range renderers {
			if !activeScene.Ready() {
				continue
			}

			if isNet {
				renderSys.Render(activeScene, screen, cliAsNet)
			} else {
				renderSys.Render(activeScene, screen, cli)
			}
		}
	}

	if ClientConfig.DebugVisual && !environment.IsProd() {
		stats := fmt.Sprintf("FRAMES: %v\nTICKS: %v", ebiten.ActualFPS(), ebiten.ActualTPS())
		ebitenutil.DebugPrint(screen.Image(), stats)
	}
}
