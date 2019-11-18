package cacheUtil

import (
	"fmt"
	"strings"

	"github.com/polariseye/goutil/syncUtil"
	"github.com/polariseye/goutil/redisUtil"
)

// redisCache is the cache that contains memory cache and redis cache
// RedisCache will set memory multiply when multiple goroutine get
type RedisCache struct {
	redisPool   *redisUtil.RedisPool
	memoryCache *MemoryCache

	marshalFunc   func(interface{}) ([]byte, error)
	unmarshalFunc func([]byte, interface{}) error

	doGroup *syncUtil.Group

	defaultRedisExpireSeconds int
}

// Set add value to memory and redis
func (r *RedisCache) Set(mainKey, subKey string, value interface{}) (err error) {
	err = r.setToRedis(mainKey, subKey, value, r.defaultRedisExpireSeconds)
	if err != nil {
		return
	}

	r.memoryCache.Set(mainKey, subKey, value)

	return
}

// set add data to memory and redis.but no expire in redis
func (r *RedisCache) SetNoExpire(mainKey, subKey string, value interface{}) (err error) {
	err = r.setToRedis(mainKey, subKey, value, 0)
	if err != nil {
		return
	}

	r.memoryCache.Set(mainKey, subKey, value)

	return
}

func (r *RedisCache) setToRedis(mainKey, subKey string, value interface{}, expireSeconds int) (err error) {
	var bytesData []byte
	bytesData, err = r.marshalFunc(value)
	if err != nil {
		return
	}

	key := r.ConvertToRedisKey(mainKey, subKey)
	if expireSeconds > 0 {
		err = r.redisPool.Set2(key, bytesData, redisUtil.Expire_Seond, expireSeconds)
	} else {
		err = r.redisPool.Set(key, bytesData)
	}
	if err != nil {
		return
	}

	return
}

// Get get from memory and redis.
// it will save to memory cache when get from redis
// it will set a nil to memory when not exist in redis
func (r *RedisCache) Get(mainKey string, subKey string, newValueFunc func() interface{}) (actualValue interface{}, ok bool, err error) {
	if actualValue, ok = r.memoryCache.GetSub(mainKey, subKey); ok {
		if actualValue == nil {
			ok = false
		}

		return
	}

	actualValue, ok, err = r.GetFromRedis(mainKey, subKey, newValueFunc)
	if err != nil {
		return
	} else if ok == false {
		// add nil to cache to avoid too many when get
		r.memoryCache.Set(mainKey, subKey, nil)
		return
	}

	// add to memory cache
	r.memoryCache.Set(mainKey, subKey, actualValue)
	return
}

// getFromRedis get value from redis directly
// it will combine get request by redis's key
func (r *RedisCache) GetFromRedis(mainKey string, subKey string, newValueFunc func() interface{}) (actualValue interface{}, ok bool, err error) {
	key := r.ConvertToRedisKey(mainKey, subKey)

	// combine redis request
	var doValue interface{}
	doValue, err = r.doGroup.Do(key, func() (interface{}, error) {
		tmpVal, tmpOk, tmpErr := r.getFromRedis(mainKey, subKey, newValueFunc)
		return []interface{}{tmpVal, tmpOk}, tmpErr
	})
	if err != nil {
		return
	}

	// return result
	var doValueList = doValue.([]interface{})
	ok = doValueList[1].(bool)
	actualValue = doValueList[0]
	return
}

// getFromRedis get value from redis directly
func (r *RedisCache) getFromRedis(mainKey string, subKey string, newValueFunc func() interface{}) (actualValue interface{}, ok bool, err error) {
	key := r.ConvertToRedisKey(mainKey, subKey)
	var bytesData []byte
	bytesData, ok, err = r.redisPool.GetBytes(key)
	if err != nil {
		return
	} else if ok == false {
		r.memoryCache.Set(mainKey, subKey, nil)
		return
	}

	actualValue=newValueFunc()
	err = r.unmarshalFunc(bytesData, actualValue)
	if err!=nil{
		actualValue=nil
	}
	return
}

// ContainsSubInMemory check if contains in memory
func (r *RedisCache) ContainsSubInMemory(mainKey, subKey string) (ok bool) {
	var val interface{}
	val, ok = r.memoryCache.GetSub(mainKey, subKey)
	if ok == false {
		return ok
	}
	if val == nil {
		return false
	}

	return true
}

// RemoveFromMemory remove from memory
func (r *RedisCache) RemoveFromMemory(mainKey string) {
	r.memoryCache.Remove(mainKey)
}

// RemoveFromMemory remove from memory
func (r *RedisCache) RemoveSubFromMemory(mainKey string, subKey string) {
	r.memoryCache.RemoveSub(mainKey, subKey)
}

// Remove remove from redis and memory
func (r *RedisCache) Remove(mainKey string, subKey string) (err error) {
	key := r.ConvertToRedisKey(mainKey, subKey)
	_, err = r.redisPool.Del(key)
	if err != nil {
		return
	}

	r.memoryCache.RemoveSub(mainKey, subKey)
	return
}

// ConvertToRedisKey convert to redis key
func (r *RedisCache) ConvertToRedisKey(mainKey string, subKey string) string {
	return fmt.Sprintf("%s.%s", mainKey, subKey)
}

// ConvertFromRedisKey convert from redis key
func (r *RedisCache) ConvertFromRedisKey(key string) (mainKey string, subKey string, err error) {
	valList := strings.Split(key, ".")
	if len(valList) < 2 {
		err = fmt.Errorf("error rediskey format")
		return
	}

	mainKey = valList[0]
	subKey = key[len(mainKey):]
	return
}

// NewRedisCache create and initialize RedisCache
func NewRedisCache(memoryCacheElementCount int, memoryExpireSeconds int,
	redisPool *redisUtil.RedisPool, redisExpireSeconds int,
	marshalFunc func(interface{}) ([]byte, error),
	unmarshalFunc func([]byte, interface{}) error) (cacheContainer *RedisCache, err error) {
	var memoryCache *MemoryCache
	memoryCache, err = NewMemoryCache(memoryCacheElementCount, memoryExpireSeconds)
	return &RedisCache{
		redisPool:                 redisPool,
		memoryCache:               memoryCache,
		defaultRedisExpireSeconds: redisExpireSeconds,
		marshalFunc:               marshalFunc,
		unmarshalFunc:             unmarshalFunc,
		doGroup:                   &syncUtil.Group{},
	}, nil
}
