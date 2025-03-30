package coldbrew

import (
	"fmt"

	"github.com/TheBitDrifter/bappa/warehouse"
)

type LocalClientSceneManager interface {
	ActivateSceneByName(string, ...warehouse.Entity) (uint32, error)
	ActivateSceneByIndex(int, ...warehouse.Entity) error

	ChangeSceneByName(string, ...warehouse.Entity) (uint32, error)
	ChangeSceneByIndex(int, ...warehouse.Entity) error
}

func (c *clientImpl) ActivateSceneByName(sceneName string, entities ...warehouse.Entity) (uint32, error) {
	cache := c.sceneManager.cache

	idx, ok := cache.GetIndex(sceneName)
	if !ok {
		return 0, fmt.Errorf("scene %s not found", sceneName)
	}

	scene := cache.GetItem(idx)

	return uint32(idx), c.sceneManager.ActivateScene(scene, entities...)
}

func (c *clientImpl) ActivateSceneByIndex(idx int, entities ...warehouse.Entity) error {
	cache := c.sceneManager.cache
	scene := cache.GetItem(idx)
	return c.sceneManager.ActivateScene(scene, entities...)
}

func (c *clientImpl) ChangeSceneByName(sceneName string, entities ...warehouse.Entity) (uint32, error) {
	cache := c.sceneManager.cache

	idx, ok := cache.GetIndex(sceneName)
	if !ok {
		return 0, fmt.Errorf("scene %s not found", sceneName)
	}

	scene := cache.GetItem(idx)

	return uint32(idx), c.sceneManager.ChangeScene(scene, entities...)
}

func (c *clientImpl) ChangeSceneByIndex(idx int, entities ...warehouse.Entity) error {
	cache := c.sceneManager.cache
	scene := cache.GetItem(idx)
	return c.sceneManager.ChangeScene(scene, entities...)
}
