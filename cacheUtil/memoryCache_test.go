package cacheUtil

import (
	"math/rand"
	"testing"

	"fmt"
	"time"
)

func BenchmarkLRU_Rand(b *testing.B) {
	l, err := NewMemoryCache(1, 8192, 0)
	if err != nil {
		b.Fatalf("err: %v", err)
	}

	trace := make([]int64, b.N*2)
	for i := 0; i < b.N*2; i++ {
		trace[i] = rand.Int63() % 32768
	}

	b.ResetTimer()

	var hit, miss int
	for i := 0; i < 2*b.N; i++ {
		if i%2 == 0 {
			l.Set(toString(trace[i]), "", trace[i])
		} else {
			_, ok := l.GetSub(toString(trace[i]), "")
			if ok {
				hit++
			} else {
				miss++
			}
		}
	}
	b.Logf("hit: %d miss: %d ratio: %f", hit, miss, float64(hit)/float64(miss))
}

func toString(val interface{}) string {
	return fmt.Sprintf("%v", val)
}

func BenchmarkLRU_Freq(b *testing.B) {
	l, err := NewMemoryCache(1, 8192, 0)
	if err != nil {
		b.Fatalf("err: %v", err)
	}

	trace := make([]int64, b.N*2)
	for i := 0; i < b.N*2; i++ {
		if i%2 == 0 {
			trace[i] = rand.Int63() % 16384
		} else {
			trace[i] = rand.Int63() % 32768
		}
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		l.Set(toString(trace[i]), "", trace[i])
	}
	var hit, miss int
	for i := 0; i < b.N; i++ {
		_, ok := l.GetSub(toString(trace[i]), "")
		if ok {
			hit++
		} else {
			miss++
		}
	}
	b.Logf("hit: %d miss: %d ratio: %f", hit, miss, float64(hit)/float64(miss))
}

func TestLRU(t *testing.T) {
	evictCounter := 0
	onEvicted := func(mainKey, subKey string, v interface{}) {
		if mainKey != toString(v) {
			t.Fatalf("Evict values not equal (%v!=%v)", mainKey, v)
		}
		evictCounter++
	}
	l, err := NewMemoryCacheWithEvict(1, 128, 0, onEvicted)
	if err != nil {
		t.Fatalf("err: %v", err)
	}

	for i := 0; i < 256; i++ {
		l.Set(toString(i), "", i)
	}
	if l.Len() != 128 {
		t.Fatalf("bad len: %v", l.Len())
	}

	if evictCounter != 128 {
		t.Fatalf("bad evict count: %v", evictCounter)
	}

	for i, k := range l.Keys() {
		if v, ok := l.GetSub(k.MainKey, k.SubKey); !ok || toString(v) != k.MainKey || v != i+128 {
			t.Fatalf("bad key: %v v:%v i:%v", k, v, i)
		}
	}
	for i := 0; i < 128; i++ {
		_, ok := l.GetSub(toString(i), "")
		if ok {
			t.Fatalf("should be evicted")
		}
	}
	for i := 128; i < 256; i++ {
		_, ok := l.GetSub(toString(i), "")
		if !ok {
			t.Fatalf("should not be evicted")
		}
	}
	for i := 128; i < 192; i++ {
		l.RemoveSub(toString(i), "")
		_, ok := l.GetSub(toString(i), "")
		if ok {
			t.Fatalf("should be deleted")
		}
	}

	l.GetSub(toString(192), "") // expect 192 to be last key in l.Keys()

	for i, k := range l.Keys() {
		if (i < 63 && k.MainKey != toString(i+193)) || (i == 63 && k.MainKey != toString(192)) {
			t.Fatalf("out of order key: %v", k)
		}
	}

	l.Purge()
	if l.Len() != 0 {
		t.Fatalf("bad len: %v", l.Len())
	}
	if _, ok := l.GetSub(toString(200), ""); ok {
		t.Fatalf("should contain nothing")
	}
}

// test that Add returns true/false if an eviction occurred
func TestLRUAdd(t *testing.T) {
	evictCounter := 0
	onEvicted := func(mainKey, subKey string, v interface{}) {
		evictCounter++
	}

	l, err := NewMemoryCacheWithEvict(1, 1, 0, onEvicted)
	if err != nil {
		t.Fatalf("err: %v", err)
	}

	if l.Set(toString(1), "", 1) == true || evictCounter != 0 {
		t.Errorf("should not have an eviction")
	}
	if l.Set(toString(2), "", 2) == false || evictCounter != 1 {
		t.Errorf("should have an eviction")
	}
}

// test that Contains doesn't update recent-ness
func TestLRUContains(t *testing.T) {
	l, err := NewMemoryCache(1, 2, 0)
	if err != nil {
		t.Fatalf("err: %v", err)
	}

	l.Set(toString(1), "", 1)
	l.Set(toString(2), "", 2)
	if !l.ContainsSub(toString(1), "") {
		t.Errorf("1 should be contained")
	}

	l.Set(toString(3), "", 3)
	if l.ContainsSub(toString(1), "") {
		t.Errorf("Contains should not have updated recent-ness of 1")
	}
}

// test that Contains doesn't update recent-ness
func TestLRUContainsOrAdd(t *testing.T) {
	l, err := NewMemoryCache(1, 2, 0)
	if err != nil {
		t.Fatalf("err: %v", err)
	}

	l.Set(toString(1), "", 1)
	l.Set(toString(2), "", 2)
	contains, evict := l.ContainsOrAdd(toString(1), "", 1)
	if !contains {
		t.Errorf("1 should be contained")
	}
	if evict {
		t.Errorf("nothing should be evicted here")
	}

	l.Set(toString(3), "", 3)
	contains, evict = l.ContainsOrAdd(toString(1), "", 1)
	if contains {
		t.Errorf("1 should not have been contained")
	}
	if !evict {
		t.Errorf("an eviction should have occurred")
	}
	if !l.ContainsSub(toString(1), "") {
		t.Errorf("now 1 should be contained")
	}
}

// test that Peek doesn't update recent-ness
func TestLRUPeek(t *testing.T) {
	l, err := NewMemoryCache(1, 2, 0)
	if err != nil {
		t.Fatalf("err: %v", err)
	}

	l.Set(toString(1), "", 1)
	l.Set(toString(2), "", 2)
	if v, ok := l.Peek(toString(1), ""); !ok || v != 1 {
		t.Errorf("1 should be set to 1: %v, %v", v, ok)
	}

	l.Set(toString(3), "", 3)
	if l.ContainsSub(toString(1), "") {
		t.Errorf("should not have updated recent-ness of 1")
	}
}

// test that Resize can upsize and downsize
func TestLRUResize(t *testing.T) {
	onEvictCounter := 0
	onEvicted := func(mainKey, subKey string, v interface{}) {
		onEvictCounter++
	}
	l, err := NewMemoryCacheWithEvict(1, 2, 0, onEvicted)
	if err != nil {
		t.Fatalf("err: %v", err)
	}

	// Downsize
	l.Set(toString(1), "", 1)
	l.Set(toString(2), "", 2)
	l.Resize(1)
	if onEvictCounter != 1 {
		t.Errorf("onEvicted should have been called 1 time: %v", onEvictCounter)
	}

	l.Set(toString(3), "", 3)
	if l.ContainsSub(toString(1), "") {
		t.Errorf("Element 1 should have been evicted")
	}

	// Upsize
	l.Resize(2)

	l.Set(toString(4), "", 4)
	if !l.ContainsSub(toString(3), "") || !l.ContainsSub(toString(4), "") {
		t.Errorf("Cache should have contained 2 elements")
	}
}

// test that expire is valid
func TestExpire(t *testing.T) {
	l, err := NewMemoryCache(1, 2, 10)
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	l.Set(toString(1), "1", 1)
	l.Set(toString(2), "2", 2)
	l.Get("1")
	l.RemoveSub(toString(1), "1")
	l.GetSub("1", "1")
	l.GetSub("1", "2")

	waitSecond := 0
	for {
		val, ok := l.GetSub(toString(2), "2")
		if !ok {
			break
		}
		if waitSecond > 20 {
			t.Fatalf("not expired waitTime:%v val:%v", waitSecond, val)
			break
		}

		time.Sleep(time.Second)
		waitSecond++
	}
	println(waitSecond)
}
