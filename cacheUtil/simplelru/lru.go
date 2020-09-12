package simplelru

import (
	"container/list"
	"errors"
	"fmt"
	"sync"
	"sync/atomic"
	"time"
)

// 最大被引用次数(当到达该次数以后该元素会被放置到列表头部)
const maxReferenceNum = 5

// 新插入的元素放置在列表的位置(百分比),认为后面的数据为冷数据
const newElemInsertIndex = 10

// EvictCallback is used to get a callback when a cache entry is evicted
type EvictCallback func(mainKey, subKey string, value interface{})

// LRU implements a non-thread safe fixed size LRU cache
type LRU struct {
	size      int
	evictList *list.List
	items     map[string]map[string]*list.Element
	onEvict   EvictCallback
	sync.RWMutex
}

// entry is used to hold a value in the evictList
type entry struct {
	mainKey      string
	subKey       string
	value        interface{}
	referenceNum int32
	lastGetTime  int64
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
	c.Lock()
	defer c.Unlock()

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
	c.Lock()

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
		c.Unlock()

		// 尝试将数据放到列表头部
		c.tryPushElemToHead(ent, subItem)
		return false
	}

	// Add new item
	ent := &entry{
		mainKey:     mainKey,
		subKey:      subKey,
		value:       value,
		lastGetTime: time.Now().Unix(),
	}

	insertIndex := c.evictList.Len() * newElemInsertIndex / 100
	var entry *list.Element
	if insertIndex <= 0 {
		entry = c.evictList.PushBack(ent)
	} else {
		temp := c.evictList.Back()
		for i := 0; i < insertIndex; i++ {
			temp = temp.Prev()
		}

		entry = c.evictList.InsertAfter(ent, temp)
	}

	mainEntry[subKey] = entry
	evict := c.evictList.Len() > c.size
	c.Unlock()

	// Verify size not exceeded
	if evict {
		c.removeOldest()
	}

	return evict
}

// Get looks up a key's value from the cache.
func (c *LRU) GetSub(mainKey string, subKey string) (value interface{}, ok bool) {
	var ent *entry
	var elem *list.Element

	c.RLock()
	var mainEntry map[string]*list.Element
	mainEntry, ok = c.items[mainKey]
	if ok == true {
		elem, ok = mainEntry[subKey]
		if ok == true {
			//c.evictList.MoveToFront(ent)
			ent, _ = elem.Value.(*entry)
			if ent != nil {
				ent.lastGetTime = time.Now().Unix()
				value = ent.value
				ok = true
			}
		}
	}

	c.RUnlock()

	// 尝试将元素防止列表头部
	c.tryPushElemToHead(ent, elem)

	return
}

// GetSub looks up a mainkey's values from cache
func (c *LRU) Get(mainKey string) (value map[string]interface{}, ok bool) {
	var subKeys []string
	c.RLock()
	mainEntry, ifExist := c.items[mainKey]
	if ifExist == true {
		subKeys = make([]string, 0)
		for subKey, _ := range mainEntry {
			subKeys = append(subKeys, subKey)
		}
	}
	c.RUnlock()

	if subKeys != nil && len(subKeys) > 0 {
		value = make(map[string]interface{})
		for _, v := range subKeys {
			subValue, exist := c.GetSub(mainKey, v)
			if exist {
				ok = true
				value[v] = subValue
			}
		}
	}

	return
}

// ContainsSub checks if a key is in the cache, without updating the recent-ness
// or deleting it for being stale.
func (c *LRU) ContainsSub(mainKey, subKey string) (ok bool) {
	c.RLock()

	var mainEntry map[string]*list.Element
	mainEntry, ok = c.items[mainKey]
	if ok == true {
		_, ok = mainEntry[subKey]
	}

	c.RUnlock()
	return ok
}

// Contains checks if a key is in the cache, without updating the recent-ness
// or deleting it for being stale.
func (c *LRU) Contains(mainKey string) (ok bool) {
	c.RLock()
	_, ok = c.items[mainKey]
	c.RUnlock()
	return ok
}

// Peek returns the key value (or undefined if not found) without updating
// the "recently used"-ness of the key.
func (c *LRU) Peek(mainKey, subKey string) (value interface{}, ok bool) {
	c.RLock()
	var mainEntry map[string]*list.Element
	if mainEntry, ok = c.items[mainKey]; ok == true {
		var ent *list.Element
		if ent, ok = mainEntry[subKey]; ok == true {
			value = ent.Value.(*entry).value
			ok = true
		}
	}

	c.RUnlock()

	return
}

// Remove removes the provided key from the cache, returning if the
// key was contained.
func (c *LRU) RemoveSub(mainKey, subKey string) (present bool) {
	var subEntry *list.Element

	c.RLock()
	mainEntry, ok := c.items[mainKey]
	if ok == true {
		subEntry, ok = mainEntry[subKey]
	}
	c.RUnlock()

	if ok == true {
		c.removeElement(subEntry)
		present = true
	}

	return
}

// Remove removes the provided key from the cache, returning if the
// key was contained.
func (c *LRU) Remove(mainKey string) (present bool) {
	c.RLock()
	mainEntry, ok := c.items[mainKey]
	var node []*list.Element
	if ok {
		node = make([]*list.Element, len(mainEntry))
		count := 0
		for k, _ := range mainEntry {
			node[count] = mainEntry[k]
			count += 1
		}
	}
	c.RUnlock()

	if ok == true {
		for _, entry := range node {
			c.removeElement(entry)
		}
	}

	return true
}

// RemoveOldest removes the oldest item from the cache.
func (c *LRU) RemoveOldest() (mainKey string, subKey string, value interface{}, ok bool) {
	c.RLock()
	ent := c.evictList.Back()
	c.RUnlock()

	if ent != nil {
		entryValue := ent.Value.(*entry)

		// remove from list
		c.removeElement(ent)

		mainKey = entryValue.mainKey
		subKey = entryValue.subKey
		value = entryValue.value
		ok = true
	}

	return
}

// GetOldest returns the oldest entry
func (c *LRU) GetOldest() (mainKey string, subKey string, value interface{}, ok bool) {
	c.RLock()
	ent := c.evictList.Back()
	if ent != nil {
		kv := ent.Value.(*entry)
		mainKey = kv.mainKey
		subKey = kv.subKey
		value = kv.value
		ok = true
	}

	c.RUnlock()
	return
}

// Keys returns a slice of the keys in the cache, from oldest to newest.
func (c *LRU) Keys() []*Key {
	c.RLock()

	keys := make([]*Key, c.evictList.Len())
	i := 0
	for ent := c.evictList.Back(); ent != nil; ent = ent.Prev() {
		kv := ent.Value.(*entry)
		keys[i] = &Key{kv.mainKey, kv.subKey}
		i++
	}

	c.RUnlock()
	return keys
}

// Len returns the number of items in the cache.
func (c *LRU) Len() int {
	c.RLock()
	result := c.evictList.Len()
	c.RUnlock()
	return result
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
	c.Lock()
	defer c.Unlock()
	for {
		backItem := c.evictList.Back()
		if backItem == nil {
			break
		}

		entry := backItem.Value.(*entry)
		if entry.lastGetTime > minSaveTime {
			break
		}

		// remove expire item
		c.removeElementNoLock(backItem)
	}
}

func (c *LRU) Print() {
	c.Lock()
	defer c.Unlock()
	element := c.evictList.Back()
	count := 0
	for {
		if element == nil {
			break
		}
		entry := element.Value.(*entry)
		fmt.Println("得到的值为", entry.value.(string))
		element = element.Prev()
		count += 1
	}

	fmt.Println("计数器为:", count)
}

// 尝试将元素放置列表头部
func (c *LRU) tryPushElemToHead(ent *entry, elem *list.Element) {
	// 验证是否需要放到队列头部
	if ent != nil && elem != nil {
		// 引用计数加1
		atomic.AddInt32(&ent.referenceNum, 1)

		// 超过一定的引用次数才将数据放到列表头部
		if atomic.LoadInt32(&ent.referenceNum) >= maxReferenceNum {
			c.Lock()
			if atomic.LoadInt32(&ent.referenceNum) >= maxReferenceNum {
				ifExist := false
				main, mainExist := c.items[ent.mainKey]
				if mainExist == true {
					if _, subExist := main[ent.subKey]; subExist == true {
						c.evictList.MoveToFront(elem)
						ifExist = true
					} else {
						main = make(map[string]*list.Element)
						c.items[ent.mainKey] = main
					}
				}
				if ifExist == false {
					c.evictList.PushFront(elem)
					main[ent.subKey] = elem
				}

				atomic.StoreInt32(&ent.referenceNum, 0)
			}

			c.Unlock()
		}
	}

}

// removeOldest removes the oldest item from the cache.
func (c *LRU) removeOldest() {
	c.RLock()
	ent := c.evictList.Back()
	c.RUnlock()

	if ent != nil {
		c.removeElement(ent)
	}
}

// removeElement is used to remove a given list element from the cache
func (c *LRU) removeElement(e *list.Element) {
	c.Lock()
	defer c.Unlock()
	c.removeElementNoLock(e)
}
func (c *LRU) removeElementNoLock(e *list.Element) {
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
