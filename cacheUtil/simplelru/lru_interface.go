package simplelru

// LRUCache is the interface for simple LRU cache.
type LRUCache interface {
	// Adds a value to the cache, returns true if an eviction occurred and
	// updates the "recently used"-ness of the key.
	Set(mainKey, subKey string, value interface{}) (evicted bool)

	// GetSub looks up a mainkey's values from cache
	// updates the "recently used"-ness of the key. #value, isFound
	Get(mainKey string) (value map[string]interface{}, ok bool)

	// Returns key's value from the cache and
	// updates the "recently used"-ness of the key. #value, isFound
	GetSub(mainKey string, subKey string) (value interface{}, ok bool)

	// Checks if a key exists in cache without updating the recent-ness.
	Contains(mainKey string) (ok bool)

	// Checks if a key exists in cache without updating the recent-ness.
	ContainsSub(mainKey, subKey string) (ok bool)

	// Returns key's value without updating the "recently used"-ness of the key.
	Peek(mainKey, subKey string) (value interface{}, ok bool)

	// Removes a key from the cache.
	Remove(mainKey string) (present bool)

	// Removes a key from the cache.
	RemoveSub(mainKey, subKey string) (present bool)

	// Removes the oldest entry from cache.
	RemoveOldest() (mainKey string, subKey string, value interface{}, ok bool)

	// Returns the oldest entry from the cache. #key, value, isFound
	GetOldest() (mainKey string, subKey string, value interface{}, ok bool)

	// Returns a slice of the keys in the cache, from oldest to newest.
	Keys() []*Key

	// Returns the number of items in the cache.
	Len() int

	// Clears all cache entries.
	Purge()

	// Resizes cache, returning number evicted
	Resize(int) int

	// remove timeout cache
	RemoveTimeoutCache(maxCacheSeconds int64)
}
