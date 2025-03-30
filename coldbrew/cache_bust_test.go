package coldbrew

import (
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/TheBitDrifter/blueprint"
	blueprintclient "github.com/TheBitDrifter/blueprint/client"
	"github.com/TheBitDrifter/warehouse"
)

// TestSuccessfulCacheBust tests a successful cache bust scenario
// where inactive scenes are released to make room for new assets
func TestSuccessfulCacheBust(t *testing.T) {
	// Create a channel to signal when the error handler is called
	errorHandlerCalled := make(chan bool, 1)

	// Use atomic flag to track if test is still running
	var testIsRunning atomic.Bool
	testIsRunning.Store(true)

	// Clear the flag when test completes
	defer testIsRunning.Store(false)

	// Set a custom error handler that signals through the channel
	SetCacheResolveErrorHandler(func(err error) {
		// Only log and signal if the test is still running
		if testIsRunning.Load() {
			t.Log("Cache resolution error detected (expected in this test)")
			select {
			case errorHandlerCalled <- true:
				// Successfully signaled
			default:
				// Channel already has a value, that's fine
			}
		}
	})

	testClient := NewTestClient(640, 480, 4, 2, 10)

	// Define assets for three scenes, each with unique assets
	sceneOneAssets := []string{"scene1_sprite1.png", "scene1_sprite2.png", "scene1_sound.wav"}
	sceneTwoAssets := []string{"scene2_sprite1.png", "scene2_sprite2.png", "scene2_sound.wav"}
	sceneThreeAssets := []string{"scene3_sprite1.png", "scene3_sprite2.png", "scene3_sound.wav"}
	sceneFourAssets := []string{"scene4_sprite1.png", "scene5_sprite2.png", "scene6_sound.wav"}

	// Get the internal client
	// Step 1: Register and activate scene one
	sceneName1 := "scene_one"
	err := testClient.RegisterScene(
		sceneName1,
		800, 600,
		createScenePlan(sceneOneAssets),
		[]RenderSystem{},
		[]ClientSystem{},
		[]blueprint.CoreSystem{},
	)
	if err != nil {
		t.Fatalf("Failed to register scene one: %v", err)
	}

	// Process the operation
	err = testClient.Update()
	if err != nil {
		t.Logf("Update note (can be ignored): %v", err)
	}
	t.Log("Scene one registered and updated")
	// Step 2: Register and activate scene two
	sceneName2 := "scene_two"
	err = testClient.RegisterScene(
		sceneName2,
		800, 600,
		createScenePlan(sceneTwoAssets),
		[]RenderSystem{},
		[]ClientSystem{},
		[]blueprint.CoreSystem{},
	)
	if err != nil {
		t.Fatalf("Failed to register scene two: %v", err)
	}

	// Activate scene two
	_, err = testClient.ActivateSceneByName(sceneName2)
	if err != nil {
		t.Fatalf("Failed to activate scene two: %v", err)
	}

	// Process the operation
	err = testClient.Update()
	if err != nil {
		t.Logf("Update note (can be ignored): %v", err)
	}
	t.Log("Scene two registered, activated and updated")

	// Get scene objects
	sceneOneIdx, _ := testClient.Cache().GetIndex(sceneName1)
	sceneOne := testClient.Cache().GetItem(sceneOneIdx)
	sceneTwoIdx, _ := testClient.Cache().GetIndex(sceneName2)
	sceneTwo := testClient.Cache().GetItem(sceneTwoIdx)

	// Check both scenes are active
	if !testClient.IsActive(sceneOne) || !testClient.IsActive(sceneTwo) {
		t.Error("Both scenes should be active at this point")
	}

	// Step 3: Deactivate scene one to free up cache space
	testClient.DeactivateScene(sceneOne)
	if testClient.IsActive(sceneOne) {
		t.Error("Scene one should be inactive after deactivation")
	}
	t.Log("Scene one deactivated")

	// Step 4: Register and activate scene three, which should trigger cache bust
	sceneName3 := "scene_three"
	err = testClient.RegisterScene(
		sceneName3,
		800, 600,
		createScenePlan(sceneThreeAssets),
		[]RenderSystem{},
		[]ClientSystem{},
		[]blueprint.CoreSystem{},
	)
	if err != nil {
		t.Fatalf("Failed to register scene three: %v", err)
	}

	// Activate scene three
	_, err = testClient.ActivateSceneByName(sceneName3)
	if err != nil {
		t.Fatalf("Failed to activate scene three: %v", err)
	}
	t.Log("Scene three registered and activated")

	// Process the operation - this should trigger a cache bust
	var updateErr error
	for i := 0; i < 5; i++ {
		updateErr = testClient.Update()
		if updateErr != nil {
			t.Logf("Update error: %v", updateErr)
		}
		time.Sleep(100 * time.Millisecond)
	}

	// Step 5: Verify both scene two and three are active and their assets are loaded
	sceneThreeIdx, _ := testClient.Cache().GetIndex(sceneName3)
	sceneThree := testClient.Cache().GetItem(sceneThreeIdx)

	if !testClient.IsActive(sceneTwo) || !testClient.IsActive(sceneThree) {
		t.Error("Scenes two and three should be active")
	}

	// Check scene one assets are not in cache (should be released during cache bust)
	for _, asset := range sceneOneAssets {
		if isAssetInCache(t, asset) {
			t.Errorf("Asset %s from inactive scene one should not be in cache", asset)
		}
	}

	// Check scene two and three assets are in cache
	for _, asset := range sceneTwoAssets {
		if !isAssetInCache(t, asset) {
			t.Errorf("Asset %s from active scene two should be in cache", asset)
		}
	}

	for _, asset := range sceneThreeAssets {
		if !isAssetInCache(t, asset) {
			t.Errorf("Asset %s from active scene three should be in cache", asset)
		}
	}

	t.Log("Successful cache bust test completed")
	t.Log("Now it's time for fail scenario")
	sceneName4 := "scene_four"
	err = testClient.RegisterScene(
		sceneName4,
		800, 600,
		createScenePlan(sceneFourAssets),
		[]RenderSystem{},
		[]ClientSystem{},
		[]blueprint.CoreSystem{},
	)
	if err != nil {
		t.Fatalf("Failed to register scene four: %v", err)
	}

	_, err = testClient.ActivateSceneByName(sceneName4)
	if err != nil {
		t.Fatalf("Failed to activate scene four: %v", err)
	}

	// Call Update several times to give the error a chance to occur
	for i := 0; i < 10; i++ {
		err = testClient.Update()
		if err != nil {
			t.Log("Update returned direct error:", err)
			break
		}

		// Check if our error handler was called
		select {
		case <-errorHandlerCalled:
			t.Log("Error handler was called successfully")
			return // Test passed
		default:
			// Not called yet, continue
		}

		time.Sleep(200 * time.Millisecond)
	}

	// Final check with timeout
	select {
	case <-errorHandlerCalled:
		t.Log("Error handler was called successfully")
	case <-time.After(3 * time.Second):
		t.Fatal("expected cache resolution error did not occur")
	}
}

// Helper function to create a scene plan
func createScenePlan(assets []string) func(width, height int, storage warehouse.Storage) error {
	return func(width, height int, storage warehouse.Storage) error {
		archetype, err := storage.NewOrExistingArchetype(
			blueprintclient.Components.SpriteBundle,
			blueprintclient.Components.SoundBundle,
		)
		if err != nil {
			return err
		}

		spriteBundle := blueprintclient.SpriteBundle{}
		for i, assetName := range assets[:2] {
			spriteBundle.Blueprints[i] = blueprintclient.SpriteBlueprint{
				Location: warehouse.CacheLocation{
					Key: assetName,
				},
			}
			spriteBundle.Blueprints[i].Config.Active = true
		}

		soundBundle := blueprintclient.SoundBundle{}
		soundBundle.Blueprints[0] = blueprintclient.SoundBlueprint{
			Location: warehouse.CacheLocation{
				Key: assets[2],
			},
			AudioPlayerCount: 1,
		}

		return archetype.Generate(1, spriteBundle, soundBundle)
	}
}

// Helper function to check if an asset is in cache
func isAssetInCache(_ *testing.T, assetName string) bool {
	cacheSwapMutex.RLock()
	defer cacheSwapMutex.RUnlock()

	if strings.HasSuffix(assetName, ".png") {
		_, exists := globalSpriteCache.GetIndex(assetName)
		return exists
	} else {
		_, exists := globalSoundCache.GetIndex(assetName)
		return exists
	}
}
