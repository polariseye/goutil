package simplelru

import "errors"

// lru池
type LruPool struct {
	// 池容量
	_pollCapacity int

	// lru 对象
	_lruPool []LRUCache
}

/*
	func:创建lru池
	param:
		int:lru池大小
		int:单个lru大小
		EvictCallback:字典回调
	return:
		*LruPool:池对象
*/
func NewLruPool(poolCapacity int, lruCapacity int, callback EvictCallback) (*LruPool, error) {
	if poolCapacity <= 0 {
		return nil, errors.New("Must provide a positive size")
	}

	tempPools := make([]LRUCache, poolCapacity)
	for i := 0; i < poolCapacity; i++ {
		temp, _ := NewLRU(lruCapacity, callback)
		tempPools[i] = temp
	}

	return &LruPool{
		_pollCapacity: poolCapacity,
		_lruPool:      tempPools,
	}, nil
}

/*
	func:根据主模块获取lru对象
	param:
		string:主模块名称
	return:
		LRUCache:LRU缓存对象
*/
func (this *LruPool) GetLruModel(mainModuleName string) LRUCache {
	index := this.getLruIndex(mainModuleName)

	return this._lruPool[index]
}

/*
	func:根据主模块获取lru的key
	param:
		string:模块名称
	return:
		int:位置
*/
func (this *LruPool) getLruIndex(mainModuleName string) int {
	total := 0
	for i := 0; i < len(mainModuleName); i++ {
		total += int(mainModuleName[i])
	}

	return total % this._pollCapacity
}

/*
	func:清理所有键值对
*/
func (this *LruPool) Purge() {
	for k, _ := range this._lruPool {
		this._lruPool[k].Purge()
	}
}

/*
	func:重新设置lru容量
	param:
		int:lru容量
*/
func (this *LruPool) Resize(size int) {
	for k, _ := range this._lruPool {
		this._lruPool[k].Resize(size)
	}
}

/*
	func: Keys returns a slice of the keys in the cache, from oldest to newest.
*/
func (this *LruPool) Keys() []*Key {
	var result []*Key

	for k, _ := range this._lruPool {
		keys := this._lruPool[k].Keys()
		result = append(result, keys...)
	}

	return result
}

/*
	func:获取长度
	return:
		int:lru缓存的数量
*/
func (this *LruPool) Len() int {
	result := 0

	for k, _ := range this._lruPool {
		result += this._lruPool[k].Len()
	}

	return result
}

/*
	func:清理过期的数据
	param:
		int:过期时间
*/
func (this *LruPool) RemoveExpired(expiredTime int) {
	for k, _ := range this._lruPool {
		this._lruPool[k].RemoveExpired(expiredTime)
	}
}
