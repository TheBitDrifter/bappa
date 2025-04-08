package coldbrew

import (
	"errors"
	"image/color"
	"io/fs"
	"log"
	"sync"

	"github.com/TheBitDrifter/bappa/blueprint"
	client "github.com/TheBitDrifter/bappa/blueprint/client"
	"github.com/TheBitDrifter/bappa/blueprint/vector"
	"github.com/TheBitDrifter/bappa/environment"
	"github.com/TheBitDrifter/bappa/table"
	"github.com/TheBitDrifter/bappa/warehouse"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	text "github.com/hajimehoshi/ebiten/v2/text/v2"
	"golang.org/x/image/font/basicfont"
)

var (
	_    Client = &clientImpl{}
	tick        = 0
)

// Client manages game state, rendering, and input
type Client interface {
	LocalClient
	SceneManager
	CameraManager
}

type LocalClient interface {
	Start() error
	RenderUtility
	TickManager
	InputManager
	CameraManager
	SystemManager
	ConfigManager
	LocalClientSceneManager
	ebiten.Game
}

type clientImpl struct {
	*tickManager
	*inputManager
	*renderUtility
	*systemManager
	*sceneManager
	*configManager
	*assetManager
}

// NewClient creates a new client with specified resolution and cache settings
func NewClient(baseResX, baseResY, maxSpritesCached, maxSoundsCached, maxScenesCached int, embeddedFS fs.FS) Client {
	baseClient := newClientImplBase(baseResX, baseResY, maxSpritesCached, maxSoundsCached, maxScenesCached, embeddedFS)
	return baseClient
}

func newClientImplBase(baseResX, baseResY, maxSpritesCached, maxSoundsCached, maxScenesCached int, embeddedFS fs.FS) *clientImpl {
	cli := &clientImpl{
		tickManager:   newTickManager(),
		renderUtility: newRenderUtility(),
		systemManager: &systemManager{},
		configManager: newConfigManager(),
		sceneManager:  newSceneManager(maxScenesCached),
		assetManager:  newAssetManager(embeddedFS),
	}
	cli.inputManager = newInputManager(cli)

	ClientConfig.maxSoundsCached.Store(uint32(maxSoundsCached))
	ClientConfig.maxSpritesCached.Store(uint32(maxSpritesCached))

	// Store base resolution
	ClientConfig.baseResolution.x = baseResX
	ClientConfig.baseResolution.y = baseResY

	ClientConfig.resolution.x = baseResX
	ClientConfig.resolution.y = baseResY
	ClientConfig.windowSize.x = baseResX
	ClientConfig.windowSize.y = baseResY

	ebiten.SetWindowSize(baseResX, baseResY)

	return cli
}

// Start initializes and runs the game loop.
func (cli *clientImpl) Start() error {
	if len(cli.loadingScenes) == 0 {
		cli.loadingScenes = append(cli.loadingScenes, defaultLoadingScene)
	}

	err := ebiten.RunGame(cli)
	if err != nil {
		return err
	}
	return nil
}

func (cli *clientImpl) Update() error {
	return sharedClientUpdate(cli)
}

func (cli *clientImpl) run() error {
	for _, globalClientSystem := range cli.globalClientSystems {
		err := globalClientSystem.Run(cli)
		if err != nil {
			return err
		}
	}

	loadingScenes := cli.loadingScenes
	for activeScene := range cli.ActiveScenes() {
		cameraReady := true
		cameras := cli.ActiveCamerasFor(activeScene)
		for _, cam := range cameras {
			if !cam.Ready(cli) {
				cameraReady = false
			}
		}
		if !cameraReady || !activeScene.Ready() {
			if len(loadingScenes) > 0 {
				loadingScene := loadingScenes[0]
				for _, coreSys := range loadingScene.CoreSystems() {
					err := coreSys.Run(loadingScene, 1.0/float64(ClientConfig.tps))
					if err != nil {
						return err
					}
				}
				for _, clientSys := range loadingScene.ClientSystems() {
					err := clientSys.Run(cli, loadingScene)
					if err != nil {
						return err
					}
				}
			}
		}
		if activeScene.Ready() {
			for _, coreSys := range activeScene.CoreSystems() {
				err := coreSys.Run(activeScene, 1.0/float64(ClientConfig.tps))
				if err != nil {
					return err
				}
			}
			for _, clientSys := range activeScene.ClientSystems() {
				err := clientSys.Run(cli, activeScene)
				if err != nil {
					return err
				}
			}
		}
	}
	return nil
}

func (cli *clientImpl) toggleDebugView() {
	if inpututil.IsKeyJustReleased(ClientConfig.DebugKey()) && !environment.IsProd() {
		ClientConfig.DebugVisual = !ClientConfig.DebugVisual
	}
}

func (cli *clientImpl) processNonExecutedPlansForActiveScenes() error {
	for s := range cli.ActiveScenes() {
		_, err := s.ExecutePlan()
		if err != nil {
			return err
		}
	}
	return nil
}

func (cli *clientImpl) findAndLoadMissingAssetsForActiveScenesAsync() {
	for scene := range cli.ActiveScenes() {
		if !scene.IsLoaded() && !scene.IsLoading() {
			if scene.TryStartLoading() {
				go func(s Scene) {
					// Get read lock before accessing global caches
					cacheSwapMutex.RLock()
					defer cacheSwapMutex.RUnlock()

					err := cli.loadAssetsForScene(s, globalSpriteCache, globalSoundCache)
					if err != nil {
						isCacheFull.Store(true)
					}
				}(scene)
			}
		}
	}
}

func (cli *clientImpl) loadAssetsForScene(scene Scene, spriteCache warehouse.Cache[Sprite], soundCache warehouse.Cache[Sound]) error {
	sto := scene.Storage()
	cursor := warehouse.Factory.NewCursor(blueprint.Queries.SpriteBundle, sto)
	for range cursor.Next() {
		bundle := client.Components.SpriteBundle.GetFromCursor(cursor)
		err := cli.SpriteLoader.Load(bundle, spriteCache)
		if err != nil {
			return err
		}

	}

	cursor = warehouse.Factory.NewCursor(blueprint.Queries.SoundBundle, sto)
	for range cursor.Next() {
		bundle := client.Components.SoundBundle.GetFromCursor(cursor)
		err := cli.SoundLoader.Load(bundle, soundCache)
		if err != nil {
			return err
		}
	}

	err := cli.SpriteLoader.PreLoad(scene.PreloadAssetBundle(), spriteCache)
	if err != nil {
		return err
	}

	err = cli.SoundLoader.PreLoad(scene.PreloadAssetBundle(), soundCache)
	if err != nil {
		return err
	}

	scene.SetLoading(false)
	scene.SetLoaded(true)
	return nil
}

func (cli *clientImpl) resolveCacheForActiveScenes() {
	if isResolvingCache.CompareAndSwap(false, true) {
		swapCacheSpr := warehouse.FactoryNewCache[Sprite](int(ClientConfig.maxSpritesCached.Load()))
		swapCacheSnd := warehouse.FactoryNewCache[Sound](int(ClientConfig.maxSoundsCached.Load()))

		var wg sync.WaitGroup
		done := make(chan struct{})
		errChan := make(chan error, cli.SceneCount())

		// Process all active scenes in parallel
		for s := range cli.ActiveScenes() {
			// Let scenes continue operating normally
			wg.Add(1)
			go func(s Scene) {
				defer wg.Done()
				err := cli.loadAssetsForScene(s, swapCacheSpr, swapCacheSnd)
				if err != nil {
					errChan <- err
				}
			}(s)
		}

		// Start a goroutine to wait for all scene loading to complete
		go func() {
			wg.Wait()
			close(done)
		}()

		go func() {
			// Wait for all goroutines to finish
			<-done

			close(errChan)
			var lastErr error
			for err := range errChan {
				lastErr = err
			}

			if lastErr != nil {
				cannotResolveCache.Store(true)
			} else {
				// Reset the cache full flag
				isCacheFull.Store(false)
			}

			isResolvingCache.Store(false)

			// Callback
			cli.onCacheResolveComplete(swapCacheSpr, swapCacheSnd, lastErr)
		}()
	}
}

func (cli *clientImpl) onCacheResolveComplete(spriteCache warehouse.Cache[Sprite], soundCache warehouse.Cache[Sound], err error) {
	if err != nil {
		handler := GetCacheResolveErrorHandler()
		log.Println(err)
		handler(err)
		return
	}

	cacheSwapMutex.Lock()
	defer cacheSwapMutex.Unlock()

	globalSpriteCache = spriteCache
	globalSoundCache = soundCache
}

func (cli *clientImpl) captureInputs() {
	cli.capturers.keyboard.Capture()
	cli.capturers.mouse.Capture()
	cli.capturers.gamepad.Capture()
	cli.capturers.touch.Capture()
}

func (cli *clientImpl) Layout(int, int) (int, int) {
	return ClientConfig.resolution.x, ClientConfig.resolution.y
}

func (cli *clientImpl) Draw(image *ebiten.Image) {
	sharedDraw(cli, image)
}

func (cli clientImpl) CameraSceneTracker() CameraSceneTracker {
	return cli.cameraSceneTracker
}

func (cli clientImpl) Cameras() [MaxSplit]Camera {
	return cli.cameras
}

func (cli clientImpl) ActivateCamera() (Camera, error) {
	for _, cam := range cli.cameras {
		if !cam.Active() {
			cam.Activate()
			// Defaults:
			cam.SetDimensions(ClientConfig.resolution.x, ClientConfig.resolution.y)
			screenPos, _ := cam.Positions()
			screenPos.X = 0
			screenPos.Y = 0

			return cam, nil
		}
	}
	return nil, errors.New("all cameras occupied")
}

var defaultLoadingScene = func() *scene {
	ls := &scene{}
	ls.name = "default loading scene"
	schema := table.Factory.NewSchema()
	ls.storage = warehouse.Factory.NewStorage(schema)
	ls.systems.renderers = append(ls.systems.renderers, defaultLoaderTextSystem{"Loading!"})
	return ls
}()

type defaultLoaderTextSystem struct {
	LoadingText string
}

func (sys defaultLoaderTextSystem) Render(scene Scene, screen Screen, c LocalClient) {
	loadingText := sys.LoadingText
	if loadingText == "" {
		loadingText = "Loading!"
	}
	for _, cam := range c.ActiveCamerasFor(scene) {
		if c.Ready(cam) {
			continue
		}
		cam.Surface().Fill(color.RGBA{R: 20, G: 0, B: 10, A: 1})
		textFace := text.NewGoXFace(basicfont.Face7x13)
		textBoundsX, textBoundsY := text.Measure(loadingText, textFace, 0)
		width, height := cam.Dimensions()
		centerX := float64((width - int(textBoundsX)) / 2)
		centerY := float64((height - int(textBoundsY)) / 2)
		cam.DrawTextBasicStatic(loadingText, &text.DrawOptions{}, textFace, vector.Two{
			X: centerX,
			Y: centerY + textBoundsY,
		})
		cam.PresentToScreen(screen, ClientConfig.cameraBorderSize)
	}
}
