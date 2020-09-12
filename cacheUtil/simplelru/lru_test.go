package simplelru

import (
	"fmt"
	"strconv"
	"sync"
	"testing"
	"time"
)

var (
	wg sync.WaitGroup
)

func TestLRU(t *testing.T) {
	wg.Add(1)
	evictCounter := 0
	onEvicted := func(mainKey, subKey string, value interface{}) {
		if mainKey != fmt.Sprintf("%v", value) {
			//t.Fatalf("Evict values not equal (%v!=%v)", mainKey, value)
		}
		evictCounter++
	}
	l, err := NewLRU(500, onEvicted)
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	go func() {
		for {
			for i := 0; i < 300; i++ {
				l.Set("main", fmt.Sprintf("%d", i), "sss")
			}
			time.Sleep(time.Second * 2)
		}
	}()

	go func() {
		for {
			for i := 0; i < 300; i++ {
				l.GetSub("main", fmt.Sprintf("%d", i))
			}
			time.Sleep(time.Second * 2)
		}
	}()

	go func() {
		for {
			l.RemoveExpired(1)
			time.Sleep(time.Second * 1)
		}
	}()

	wg.Wait()
	//
	//for i := 0; i < 256; i++ {
	//	l.Set("main", fmt.Sprintf("%v", i), i)
	//}
	//if l.Len() != 128 {
	//	t.Fatalf("bad len: %v", l.Len())
	//}
	//
	//if evictCounter != 128 {
	//	t.Fatalf("bad evict count: %v", evictCounter)
	//}
	//
	//for i, k := range l.Keys() {
	//	if v, ok := l.GetSub(k.MainKey, k.SubKey); !ok || fmt.Sprintf("%v", v) != k.MainKey || v != i+128 {
	//		t.Fatalf("bad key: %v", k)
	//	}
	//}
	//go func() {
	//	for {
	//		for i := 0; i < 128; i++ {
	//			_, ok := l.GetSub("main", strconv.Itoa(i))
	//			if ok {
	//				t.Fatalf("should be evicted")
	//			}
	//		}
	//	}
	//}()
	//for i := 128; i < 256; i++ {
	//	_, ok := l.GetSub("main", strconv.Itoa(i))
	//	if !ok {
	//		//t.Fatalf("should not be evicted")
	//	}
	//}
	//for i := 128; i < 192; i++ {
	//	ok := l.RemoveSub("main", strconv.Itoa(i))
	//	if !ok {
	//		t.Fatalf("should be contained")
	//	}
	//	ok = l.RemoveSub("main", strconv.Itoa(i))
	//	if ok {
	//		t.Fatalf("should not be contained")
	//	}
	//	_, ok = l.GetSub("main", strconv.Itoa(i))
	//	if ok {
	//		t.Fatalf("should be deleted")
	//	}
	//}
	//
	//l.GetSub("main", fmt.Sprintf("%v", 192)) // expect 192 to be last key in l.Keys()
	//
	//for i, k := range l.Keys() {
	//	if (i < 63 && k.MainKey != strconv.Itoa(i+193)) || (i == 63 && k.MainKey != strconv.Itoa(192)) {
	//		t.Fatalf("out of order key: %v", k)
	//	}
	//}
	//
	//l.Purge()
	//if l.Len() != 0 {
	//	t.Fatalf("bad len: %v", l.Len())
	//}
	//if _, ok := l.GetSub("main", "200"); ok {
	//	t.Fatalf("should contain nothing")
	//}
}

func TestLRU_GetOldest_RemoveOldest(t *testing.T) {
	l, err := NewLRU(128, nil)
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	for i := 0; i < 256; i++ {
		l.Set(strconv.Itoa(i), "", i)
	}
	k, _, _, ok := l.GetOldest()
	if !ok {
		t.Fatalf("missing")
	}
	if k != "128" {
		t.Fatalf("bad: %v", k)
	}

	k, _, _, ok = l.RemoveOldest()
	if !ok {
		t.Fatalf("missing")
	}
	if k != "128" {
		t.Fatalf("bad: %v", k)
	}

	k, _, _, ok = l.RemoveOldest()
	if !ok {
		t.Fatalf("missing")
	}
	if k != "129" {
		t.Fatalf("bad: %v", k)
	}
}

// Test that Add returns true/false if an eviction occurred
func TestLRU_Add(t *testing.T) {
	evictCounter := 0
	onEvicted := func(mainKey, subKey string, v interface{}) {
		evictCounter++
	}

	l, err := NewLRU(1, onEvicted)
	if err != nil {
		t.Fatalf("err: %v", err)
	}

	if l.Set("1", "", 1) == true || evictCounter != 0 {
		t.Errorf("should not have an eviction")
	}
	if l.Set("2", "", 2) == false || evictCounter != 1 {
		t.Errorf("should have an eviction")
	}
}

// Test that Contains doesn't update recent-ness
func TestLRU_Contains(t *testing.T) {
	l, err := NewLRU(2, nil)
	if err != nil {
		t.Fatalf("err: %v", err)
	}

	l.Set("1", "", 1)
	l.Set("2", "", 2)
	if !l.ContainsSub("1", "") {
		t.Errorf("1 should be contained")
	}

	l.Set("3", "", 3)
	if l.ContainsSub("1", "") {
		t.Errorf("Contains should not have updated recent-ness of 1")
	}
}

// Test that Peek doesn't update recent-ness
func TestLRU_Peek(t *testing.T) {
	l, err := NewLRU(2, nil)
	if err != nil {
		t.Fatalf("err: %v", err)
	}

	l.Set("1", "", 1)
	l.Set("2", "", 2)
	if v, ok := l.Peek("1", ""); !ok || v != 1 {
		t.Errorf("1 should be set to 1: %v, %v", v, ok)
	}

	l.Set("3", "", 3)
	if l.ContainsSub("1", "") {
		t.Errorf("should not have updated recent-ness of 1")
	}
}

// Test that Resize can upsize and downsize
func TestLRU_Resize(t *testing.T) {
	onEvictCounter := 0
	onEvicted := func(mainKey, subKey string, v interface{}) {
		onEvictCounter++
	}
	l, err := NewLRU(2, onEvicted)
	if err != nil {
		t.Fatalf("err: %v", err)
	}

	// Downsize
	l.Set("1", "", 1)
	l.Set("2", "", 2)
	evicted := l.Resize(1)
	if evicted != 1 {
		t.Errorf("1 element should have been evicted: %v", evicted)
	}
	if onEvictCounter != 1 {
		t.Errorf("onEvicted should have been called 1 time: %v", onEvictCounter)
	}

	l.Set("3", "", 3)
	if l.ContainsSub("1", "") {
		t.Errorf("Element 1 should have been evicted")
	}

	// Upsize
	evicted = l.Resize(2)
	if evicted != 0 {
		t.Errorf("0 elements should have been evicted: %v", evicted)
	}

	l.Set("4", "", 4)
	if !l.ContainsSub("3", "") || !l.ContainsSub("4", "") {
		t.Errorf("Cache should have contained 2 elements")
	}
}
