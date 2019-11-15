package simplelru

import (
	"container/list"
	"errors"
	"time"
)

// EvictCallback is used to get a callback when a cache entry is evicted
type EvictCallback func(mainKey, subKey string, value interface{})

// LRU implements a non-thread safe fixed size LRU cache
type LRU struct {
	size      int
	evictList *list.List
	items     map[string]map[string]*list.Element
	onEvict   EvictCallback
}

// entry is used to hold a value in the evictList
type entry struct {
	mainKey     string
	subKey      string
	value       interface{}
	lastGetTime int64
}

// Key is used to return entry's key
type Key struct {
	MainKey string
	SubKey  string
}

// NewLRU constructs an LRU of the given size
func NewLRU(size int, onEvict EvictCallback) (*LRU, error) {
	if size <= 0 {
		return nil, errors.New("Must provide a positive size")
	}
	c := &LRU{
		size:      size,
		evictList: list.New(),
		items:     make(map[string]map[string]*list.Element),
		onEvict:   onEvict,
	}
	return c, nil
}

// Purge is used to completely clear the cache.
func (c *LRU) Purge() {
	for mainKey, valueByMainKey := range c.items {
		for subKey, value := range valueByMainKey {
			if c.onEvict != nil {
				c.onEvict(mainKey, subKey, value.Value.(*entry).value)
			}
		}

		delete(c.items, mainKey)
	}
	c.evictList.Init()
}

// Set adds a value to the cache.  Returns true if an eviction occurred.
func (c *LRU) Set(mainKey, subKey string, value interface{}) (evicted bool) {
	// Check for existing item
	mainEntry, exist := c.items[mainKey]
	if exist == false {
		mainEntry = make(map[string]*list.Element)
		c.items[mainKey] = mainEntry
	}
	subItem, exist := mainEntry[subKey]
	if exist {
		ent := subItem.Value.(*entry)
		ent.value = value
		ent.lastGetTime = time.Now().Unix()

		c.evictList.MoveToFront(subItem)

		return false
	}

	// Add new item
	ent := &entry{mainKey, subKey, value, time.Now().Unix()}
	entry := c.evictList.PushFront(ent)
	mainEntry[subKey] = entry

	evict := c.evictList.Len() > c.size
	// Verify size not exceeded
	if evict {
		c.removeOldest()
	}
	return evict
}

// Get looks up a key's value from the cache.
func (c *LRU) GetSub(mainKey string, subKey string) (value interface{}, ok bool) {
	mainEntry, ok := c.items[mainKey]
	if !ok {
		return
	}
	ent, ok := mainEntry[subKey]
	if !ok {
		return
	}

	c.evictList.MoveToFront(ent)
	if ent.Value.(*entry) == nil {
		return nil, false
	}
	return ent.Value.(*entry).value, true
}

// GetSub looks up a mainkey's values from cache
func (c *LRU) Get(mainKey string) (value map[string]interface{}, ok bool) {
	mainEntry, ok := c.items[mainKey]
	if !ok {
		return
	}

	value = make(map[string]interface{}, len(mainEntry))
	for subKey, ent := range mainEntry {
		c.evictList.MoveToFront(ent)
		kv := ent.Value.(*entry)
		value[subKey] = kv.value
	}

	return value, true
}

// ContainsSub checks if a key is in the cache, without updating the recent-ness
// or deleting it for being stale.
func (c *LRU) ContainsSub(mainKey, subKey string) (ok bool) {
	var mainEntry map[string]*list.Element
	mainEntry, ok = c.items[mainKey]
	if !ok {
		return
	}

	_, ok = mainEntry[subKey]

	return ok
}

// Contains checks if a key is in the cache, without updating the recent-ness
// or deleting it for being stale.
func (c *LRU) Contains(mainKey string) (ok bool) {
	_, ok = c.items[mainKey]

	return ok
}

// Peek returns the key value (or undefined if not found) without updating
// the "recently used"-ness of the key.
func (c *LRU) Peek(mainKey, subKey string) (value interface{}, ok bool) {
	var mainEntry map[string]*list.Element
	if mainEntry, ok = c.items[mainKey]; !ok {
		return nil, ok
	}

	var ent *list.Element
	if ent, ok = mainEntry[subKey]; !ok {
		return nil, ok
	}

	return ent.Value.(*entry).value, true
}

// Remove removes the provided key from the cache, returning if the
// key was contained.
func (c *LRU) RemoveSub(mainKey, subKey string) (present bool) {
	mainEntry, ok := c.items[mainKey]
	if !ok {
		return false
	}

	subEntry, ok := mainEntry[subKey]
	if !ok {
		return false
	}

	c.removeElement(subEntry)

	return true
}

// Remove removes the provided key from the cache, returning if the
// key was contained.
func (c *LRU) Remove(mainKey string) (present bool) {
	mainEntry, ok := c.items[mainKey]
	if !ok {
		return false
	}

	for _, entry := range mainEntry {
		c.removeElement(entry)
	}

	return true
}

// RemoveOldest removes the oldest item from the cache.
func (c *LRU) RemoveOldest() (mainKey string, subKey string, value interface{}, ok bool) {
	ent := c.evictList.Back()
	if ent == nil {
		return "", "", nil, false
	}

	entryValue := ent.Value.(*entry)

	// remove from list
	c.removeElement(ent)

	return entryValue.mainKey, entryValue.subKey, entryValue.value, true
}

// GetOldest returns the oldest entry
func (c *LRU) GetOldest() (mainKey string, subKey string, value interface{}, ok bool) {
	ent := c.evictList.Back()
	if ent != nil {
		kv := ent.Value.(*entry)
		return kv.mainKey, kv.subKey, kv.value, true
	}

	return "", "", nil, false
}

// Keys returns a slice of the keys in the cache, from oldest to newest.
func (c *LRU) Keys() []*Key {
	keys := make([]*Key, c.evictList.Len())
	i := 0
	for ent := c.evictList.Back(); ent != nil; ent = ent.Prev() {
		kv := ent.Value.(*entry)
		keys[i] = &Key{kv.mainKey, kv.subKey}
		i++
	}
	return keys
}

// Len returns the number of items in the cache.
func (c *LRU) Len() int {
	return c.evictList.Len()
}

// Resize changes the cache size.
func (c *LRU) Resize(size int) (evicted int) {
	diff := c.Len() - size
	if diff < 0 {
		diff = 0
	}
	for i := 0; i < diff; i++ {
		c.removeOldest()
	}
	c.size = size
	return diff
}

// RemoveExpired remove timeout cache
func (c *LRU) RemoveExpired(expireSeconds int) {
	minSaveTime := time.Now().Unix() - int64(expireSeconds)
	for {
		backItem := c.evictList.Back()
		if backItem == nil {
			continue
		}

		entry := backItem.Value.(*entry)
		if entry.lastGetTime >= minSaveTime {
			break
		}

		// remove expire item
		c.removeElement(backItem)
	}
}

// removeOldest removes the oldest item from the cache.
func (c *LRU) removeOldest() {
	ent := c.evictList.Back()
	if ent != nil {
		c.removeElement(ent)
	}
}

// removeElement is used to remove a given list element from the cache
func (c *LRU) removeElement(e *list.Element) {
	c.evictList.Remove(e)
	kv := e.Value.(*entry)

	mainEntry := c.items[kv.mainKey]
	delete(mainEntry, kv.subKey)
	if len(mainEntry) <= 0 {
		delete(c.items, kv.mainKey)
	}

	if c.onEvict != nil {
		c.onEvict(kv.mainKey, kv.subKey, kv.value)
	}
}
