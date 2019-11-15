package cacheUtil

import (
	"sync"
	"time"

	"github.com/polariseye/goutil/cacheUtil/simplelru"
)

// MemoryCache is a thread-safe fixed size LRU cache.
type MemoryCache struct {
	lru             simplelru.LRUCache
	lock            sync.RWMutex
	maxCacheSeconds int64
}

// NewMemoryCache creates an LRU of the given size.
func NewMemoryCache(size int, maxCacheSeconds int64) (*MemoryCache, error) {
	return NewMemoryCacheWithEvict(size, maxCacheSeconds, nil)
}

// NewMemoryCacheWithEvict constructs a fixed size cache with the given eviction
// callback.
func NewMemoryCacheWithEvict(size int, _maxCacheSeconds int64, onEvicted func(mainKey, subKey interface{}, value interface{})) (*MemoryCache, error) {
	lru, err := simplelru.NewLRU(size, simplelru.EvictCallback(onEvicted))
	if err != nil {
		return nil, err
	}
	c := &MemoryCache{
		lru:             lru,
		maxCacheSeconds: _maxCacheSeconds,
	}
	if _maxCacheSeconds > 0 {
		go c.removeTimeout()
	}

	return c, nil
}

// Purge is used to completely clear the cache.
func (c *MemoryCache) Purge() {
	c.lock.Lock()
	c.lru.Purge()
	c.lock.Unlock()
}

// Add adds a value to the cache.  Returns true if an eviction occurred.
func (c *MemoryCache) Set(mainKey, subKey string, value interface{}) (evicted bool) {
	c.lock.Lock()
	evicted = c.lru.Set(mainKey, subKey, value)
	c.lock.Unlock()
	return evicted
}

// Get looks up a key's value from the cache.
func (c *MemoryCache) Get(mainKey string) (value map[string]interface{}, ok bool) {
	c.lock.Lock()
	value, ok = c.lru.Get(mainKey)
	c.lock.Unlock()
	return value, ok
}

// GetSub looks up a key's value from the cache.
func (c *MemoryCache) GetSub(mainKey string, subKey string) (value interface{}, ok bool) {
	c.lock.Lock()
	value, ok = c.lru.GetSub(mainKey, subKey)
	c.lock.Unlock()
	return value, ok
}

// Contains checks if a key is in the cache, without updating the
// recent-ness or deleting it for being stale.
func (c *MemoryCache) Contains(mainKey string) (ok bool) {
	c.lock.RLock()
	containKey := c.lru.Contains(mainKey)
	c.lock.RUnlock()
	return containKey
}

// Contains checks if a key is in the cache, without updating the
// recent-ness or deleting it for being stale.
func (c *MemoryCache) ContainsSub(mainKey, subKey string) (ok bool) {
	c.lock.RLock()
	containKey := c.lru.Contains(mainKey)
	c.lock.RUnlock()
	return containKey
}

// Peek returns the key value (or undefined if not found) without updating
// the "recently used"-ness of the key.
func (c *MemoryCache) Peek(mainKey, subKey string) (value interface{}, ok bool) {
	c.lock.RLock()
	value, ok = c.lru.Peek(mainKey, subKey)
	c.lock.RUnlock()
	return value, ok
}

// ContainsOrAdd checks if a key is in the cache  without updating the
// recent-ness or deleting it for being stale,  and if not, adds the value.
// Returns whether found and whether an eviction occurred.
func (c *MemoryCache) ContainsOrAdd(mainKey string, subKey string, value interface{}) (ok, evicted bool) {
	c.lock.Lock()
	defer c.lock.Unlock()

	if c.lru.ContainsSub(mainKey, subKey) {
		return true, false
	}

	evicted = c.lru.Set(mainKey, subKey, value)
	return false, evicted
}

// Remove removes the provided key from the cache.
func (c *MemoryCache) Remove(mainKey string) (present bool) {
	c.lock.Lock()
	present = c.lru.Remove(mainKey)
	c.lock.Unlock()
	return
}

// Remove removes the provided key from the cache.
func (c *MemoryCache) RemoveSub(mainKey, subKey string) (present bool) {
	c.lock.Lock()
	present = c.lru.RemoveSub(mainKey, subKey)
	c.lock.Unlock()
	return
}

// Resize changes the cache size.
func (c *MemoryCache) Resize(size int) (evicted int) {
	c.lock.Lock()
	evicted = c.lru.Resize(size)
	c.lock.Unlock()
	return evicted
}

// RemoveOldest removes the oldest item from the cache.
func (c *MemoryCache) RemoveOldest() (mainKey string, subKey string, value interface{}, ok bool) {
	c.lock.Lock()
	mainKey, subKey, value, ok = c.lru.RemoveOldest()
	c.lock.Unlock()
	return
}

// GetOldest returns the oldest entry
func (c *MemoryCache) GetOldest() (mainKey string, subKey string, value interface{}, ok bool) {
	c.lock.Lock()
	mainKey, subKey, value, ok = c.lru.GetOldest()
	c.lock.Unlock()
	return
}

// Keys returns a slice of the keys in the cache, from oldest to newest.
func (c *MemoryCache) Keys() []*simplelru.Key {
	c.lock.RLock()
	keys := c.lru.Keys()
	c.lock.RUnlock()
	return keys
}

// Len returns the number of items in the cache.
func (c *MemoryCache) Len() int {
	c.lock.RLock()
	length := c.lru.Len()
	c.lock.RUnlock()
	return length
}

func (c *MemoryCache) removeTimeout() {
	for {
		<-time.After(time.Duration(c.maxCacheSeconds) * time.Second)

		c.lock.Lock()
		c.lru.RemoveTimeoutCache(c.maxCacheSeconds)
		c.lock.Unlock()
	}
}
