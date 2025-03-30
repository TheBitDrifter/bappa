package coldbrew

import (
	"sync"
	"sync/atomic"

	"github.com/TheBitDrifter/warehouse"
)

var (
	globalSoundCache   = warehouse.FactoryNewCache[Sound](int(ClientConfig.maxSoundsCached.Load()))
	globalSpriteCache  = warehouse.FactoryNewCache[Sprite](int(ClientConfig.maxSpritesCached.Load()))
	cacheSwapMutex     sync.RWMutex
	isCacheFull        atomic.Bool
	cannotResolveCache atomic.Bool
	isResolvingCache   atomic.Bool
)

type CacheBustError struct{}

func (*CacheBustError) Error() string {
	return "cache bust failed active scenes require more assets than available for settings"
}
