package cacheUtil

import (
	"testing"
	"time"

	"github.com/polariseye/goutil/redisUtil"
)

type TVal struct {
	Val int
}

func TestCache(t *testing.T) {
	redisPoolObj := redisUtil.NewRedisPool2("tst", &redisUtil.RedisConfig{
		ConnectionString:   "127.0.0.1:6379",
		Database:           0,
		MaxActive:          10,
		MaxIdle:            1,
		IdleTimeout:        60 * time.Second,
		DialConnectTimeout: 2 * time.Second,
	})
	defer redisPoolObj.Close()
	if err := redisPoolObj.TestConnection(); err != nil {
		t.Fatal(err.Error())
		return
	}
	cacheObj, err := NewRedisCache(2, 10, 100, redisPoolObj, 100, nil, nil)
	if err != nil {
		t.Fatal(err.Error())
		return
	}

	err = cacheObj.Set("1", "", &TVal{1})
	if err != nil {
		t.Fatal(err.Error())
		return
	}
	cacheObj.RemoveFromMemory("1")

	// get from redis test
	tmpVal, ok, err := cacheObj.GetFromRedis("1", "", func() interface{} { return &TVal{} })
	if err != nil {
		t.Fatal(err.Error())
	} else if ok == false {
		t.Fatal("data no exist")
	}
	if tmpVal.(*TVal).Val != 1 {
		t.Fatal("data no correct")
	}
	tmpVal, ok = cacheObj.memoryCache.GetSub("1", "")
	if ok {
		t.Fatal("should not exist")
	}

	// normal get set
	tmpVal, ok, err = cacheObj.Get("1", "", func() interface{} { return &TVal{} })
	if err != nil {
		t.Fatal(err.Error())
	} else if ok == false {
		t.Fatal("data no exist")
	}
	if tmpVal.(*TVal).Val != 1 {
		t.Fatal("data no correct")
	}

	cacheObj.RemoveFromMemory("1")
	tmpVal, ok, err = cacheObj.Get("1", "", func() interface{} { return &TVal{} })
	if err != nil {
		t.Fatal(err.Error())
	} else if ok == false {
		t.Fatal("data no exist")
	}
	if tmpVal.(*TVal).Val != 1 {
		t.Fatal("data no correct")
	}

	tmpVal, ok = cacheObj.memoryCache.GetSub("1", "")
	if ok == false {
		t.Fatal("data no exist")
	}
	if tmpVal.(*TVal).Val != 1 {
		t.Fatal("data no correct")
	}
}

func TestRemove(t *testing.T) {
	redisPoolObj := redisUtil.NewRedisPool2("tst", &redisUtil.RedisConfig{
		ConnectionString:   "127.0.0.1:6379",
		Database:           0,
		MaxActive:          10,
		MaxIdle:            1,
		IdleTimeout:        60 * time.Second,
		DialConnectTimeout: 2 * time.Second,
	})
	defer redisPoolObj.Close()
	if err := redisPoolObj.TestConnection(); err != nil {
		t.Fatal(err.Error())
		return
	}
	cacheObj, err := NewRedisCache(2, 10, 10, redisPoolObj, 100, nil, nil)
	if err != nil {
		t.Fatal(err.Error())
		return
	}

	err = cacheObj.Set("1", "", &TVal{1})
	if err != nil {
		t.Fatal(err.Error())
		return
	}
	cacheObj.Remove("1", "")
	if cacheObj.ContainsSubInMemory("1", "") {
		t.Fatal("data should not exist")
	}
	if exist, err := cacheObj.ContainsInRedis("1", ""); err != nil {
		t.Fatal("error", err.Error())
	} else if exist {
		t.Fatal("data should not exist")
	}
}
