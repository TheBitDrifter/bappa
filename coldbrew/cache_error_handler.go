package coldbrew

import (
	"log"
	"sync"
)

// CacheResolveErrorHandler is a function type for handling cache resolution errors
type CacheResolveErrorHandler func(error)

// Default handler that uses log.Fatal
func defaultCacheResolveErrorHandler(err error) {
	log.Fatal("Cannot resolve cache")
}

// Mutex to protect access to the error handler
var errorHandlerMutex sync.RWMutex

// The current error handler function - can be swapped for testing
var cacheResolveErrorHandler CacheResolveErrorHandler = defaultCacheResolveErrorHandler

// SetCacheResolveErrorHandler allows changing the error handler, returns the previous handler
func SetCacheResolveErrorHandler(handler CacheResolveErrorHandler) CacheResolveErrorHandler {
	errorHandlerMutex.Lock()
	defer errorHandlerMutex.Unlock()

	previous := cacheResolveErrorHandler
	cacheResolveErrorHandler = handler
	return previous
}

// GetCacheResolveErrorHandler safely retrieves the current error handler
func GetCacheResolveErrorHandler() CacheResolveErrorHandler {
	errorHandlerMutex.RLock()
	defer errorHandlerMutex.RUnlock()

	return cacheResolveErrorHandler
}
