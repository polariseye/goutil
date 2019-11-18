package cacheUtil

import (
	"testing"
	"time"

	"github.com/polariseye/goutil/redisUtil"
	"go.mongodb.org/mongo-driver/bson"
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
	cacheObj, err := NewRedisCache(2, 10, redisPoolObj, 100, bson.Marshal, bson.Unmarshal)
	if err != nil {
		t.Fatal(err.Error())
		return
	}

	err = cacheObj.Set("1", "", &TVal{1})
	if err != nil {
		t.Fatal(err.Error())
		return
	}

	tmpVal, ok, err := cacheObj.Get("1", "", func() interface{} { return &TVal{} })
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
