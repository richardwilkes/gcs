package gurps

import "sync"

// ScriptResolveCache is a thread-safe cache for resolved script values.
type ScriptResolveCache struct {
	lock  sync.RWMutex
	cache map[string]string
}

// NewScriptResolveCache creates a new instance.
func NewScriptResolveCache() *ScriptResolveCache {
	return &ScriptResolveCache{
		cache: make(map[string]string),
	}
}

// Get retrieves a value from the cache.
func (rc *ScriptResolveCache) Get(key string) (string, bool) {
	rc.lock.RLock()
	defer rc.lock.RUnlock()
	value, exists := rc.cache[key]
	return value, exists
}

// Set adds or updates a value in the cache.
func (rc *ScriptResolveCache) Set(key, value string) {
	rc.lock.Lock()
	defer rc.lock.Unlock()
	rc.cache[key] = value
}
