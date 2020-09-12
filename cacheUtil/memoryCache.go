package cacheUtil

import (
	"github.com/polariseye/goutil/cacheUtil/simplelru"
	"sync"
	"time"
)

// MemoryCache is a thread-safe fixed size LRU cache.
type MemoryCache struct {
	_lruPool       *simplelru.LruPool
	_lock          sync.RWMutex
	_expireSeconds int
}

// NewMemoryCache creates an LRU of the given size.
func NewMemoryCache(lruCapacity, size int, expireSeconds int) (*MemoryCache, error) {
	return NewMemoryCacheWithEvict(lruCapacity, size, expireSeconds, nil)
}

// NewMemoryCacheWithEvict constructs a fixed size cache with the given eviction
// callback.
func NewMemoryCacheWithEvict(lruCapacity, size int, _expireSeconds int, onEvicted func(mainKey, subKey string, value interface{})) (*MemoryCache, error) {
	lruPool, err := simplelru.NewLruPool(lruCapacity, size, simplelru.EvictCallback(onEvicted))
	if err != nil {
		return nil, err
	}
	c := &MemoryCache{
		_lruPool:       lruPool,
		_expireSeconds: _expireSeconds,
	}
	if _expireSeconds > 0 {
		go c.removeExpired()
	}

	return c, nil
}

// Purge is used to completely clear the cache.
func (c *MemoryCache) Purge() {
	c._lruPool.Purge()
}

// Add adds a value to the cache.  Returns true if an eviction occurred.
func (c *MemoryCache) Set(mainKey, subKey string, value interface{}) (evicted bool) {
	lru := c._lruPool.GetLruModel(mainKey)
	evicted = lru.Set(mainKey, subKey, value)
	return evicted
}

// Get looks up a key's value from the cache.
func (c *MemoryCache) Get(mainKey string) (value map[string]interface{}, ok bool) {
	lru := c._lruPool.GetLruModel(mainKey)
	value, ok = lru.Get(mainKey)
	return value, ok
}

// GetSub looks up a key's value from the cache.
func (c *MemoryCache) GetSub(mainKey string, subKey string) (value interface{}, ok bool) {
	lru := c._lruPool.GetLruModel(mainKey)
	value, ok = lru.GetSub(mainKey, subKey)
	return value, ok
}

// GetSubOrSet looks up a key's value from the cache.will add it if no exist
func (c *MemoryCache) GetSubOrSet(mainKey string, subKey string, newValFunc func() interface{}) (value interface{}) {
	ok := false
	lru := c._lruPool.GetLruModel(mainKey)
	value, ok = lru.GetSub(mainKey, subKey)
	if ok == false {
		value = newValFunc()
		lru.Set(mainKey, subKey, value)
	}

	return value
}

// Contains checks if a key is in the cache, without updating the
// recent-ness or deleting it for being stale.
func (c *MemoryCache) Contains(mainKey string) (ok bool) {
	lru := c._lruPool.GetLruModel(mainKey)
	containKey := lru.Contains(mainKey)
	return containKey
}

// Contains checks if a key is in the cache, without updating the
// recent-ness or deleting it for being stale.
func (c *MemoryCache) ContainsSub(mainKey, subKey string) (ok bool) {
	lru := c._lruPool.GetLruModel(mainKey)
	containKey := lru.Contains(mainKey)
	return containKey
}

// Peek returns the key value (or undefined if not found) without updating
// the "recently used"-ness of the key.
func (c *MemoryCache) Peek(mainKey, subKey string) (value interface{}, ok bool) {
	lru := c._lruPool.GetLruModel(mainKey)
	value, ok = lru.Peek(mainKey, subKey)
	return value, ok
}

// ContainsOrAdd checks if a key is in the cache  without updating the
// recent-ness or deleting it for being stale,  and if not, adds the value.
// Returns whether found and whether an eviction occurred.
func (c *MemoryCache) ContainsOrAdd(mainKey string, subKey string, value interface{}) (ok, evicted bool) {
	lru := c._lruPool.GetLruModel(mainKey)
	if lru.ContainsSub(mainKey, subKey) {
		return true, false
	}

	evicted = lru.Set(mainKey, subKey, value)
	return false, evicted
}

// Remove removes the provided key from the cache.
func (c *MemoryCache) Remove(mainKey string) (present bool) {
	lru := c._lruPool.GetLruModel(mainKey)
	present = lru.Remove(mainKey)
	return
}

// Remove removes the provided key from the cache.
func (c *MemoryCache) RemoveSub(mainKey, subKey string) (present bool) {
	lru := c._lruPool.GetLruModel(mainKey)
	present = lru.RemoveSub(mainKey, subKey)
	return
}

// Resize changes the cache size.
func (c *MemoryCache) Resize(size int) {
	c._lruPool.Resize(size)
}

// RemoveOldest removes the oldest item from the cache.
func (c *MemoryCache) RemoveOldest() (mainKey string, subKey string, value interface{}, ok bool) {
	lru := c._lruPool.GetLruModel(mainKey)
	mainKey, subKey, value, ok = lru.RemoveOldest()
	return
}

// GetOldest returns the oldest entry
func (c *MemoryCache) GetOldest() (mainKey string, subKey string, value interface{}, ok bool) {
	lru := c._lruPool.GetLruModel(mainKey)
	mainKey, subKey, value, ok = lru.GetOldest()
	return
}

// Keys returns a slice of the keys in the cache, from oldest to newest.
func (c *MemoryCache) Keys() []*simplelru.Key {
	keys := c._lruPool.Keys()
	return keys
}

// Len returns the number of items in the cache.
func (c *MemoryCache) Len() int {
	length := c._lruPool.Len()
	return length
}

func (c *MemoryCache) removeExpired() {
	for {
		time.Sleep(time.Duration(c._expireSeconds) * time.Second)
		c._lruPool.RemoveExpired(c._expireSeconds)
	}
}
