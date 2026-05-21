package cache

import (
	"fmt"
	"sync"
	"testing"
	"time"
)

func TestSetAndGet(t *testing.T) {
	c := NewCache()
	c.Set("foo", "bar", time.Minute)

	v, ok := c.Get("foo")
	if !ok {
		t.Fatal("expected foo to exist")
	}
	if v != "bar" {
		t.Fatalf("expected bar, got %v", v)
	}
}

func TestGetMissingKey(t *testing.T) {
	c := NewCache()
	v, ok := c.Get("missing")
	if ok {
		t.Fatal("expected missing key to return ok=false")
	}
	if v != nil {
		t.Fatalf("expected nil value, got %v", v)
	}
}

func TestExpiration(t *testing.T) {
	c := NewCache()
	c.Set("foo", "bar", 50*time.Millisecond)

	if _, ok := c.Get("foo"); !ok {
		t.Fatal("expected foo to exist immediately after Set")
	}

	time.Sleep(100 * time.Millisecond)

	if _, ok := c.Get("foo"); ok {
		t.Fatal("expected foo to be expired")
	}
}

func TestDelete(t *testing.T) {
	c := NewCache()
	c.Set("foo", "bar", time.Minute)
	c.Delete("foo")

	if _, ok := c.Get("foo"); ok {
		t.Fatal("expected foo to be deleted")
	}
}

func TestLen(t *testing.T) {
	c := NewCache()
	c.Set("a", 1, time.Minute)
	c.Set("b", 2, time.Minute)
	c.Set("c", 3, 50*time.Millisecond)

	if got := c.Len(); got != 3 {
		t.Fatalf("expected Len=3, got %d", got)
	}

	time.Sleep(100 * time.Millisecond)

	if got := c.Len(); got != 2 {
		t.Fatalf("expected Len=2 after expiration, got %d", got)
	}
}

func TestConcurrentSet(t *testing.T) {
	c := NewCache()
	var wg sync.WaitGroup
	const n = 100

	for i := 0; i < n; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			c.Set(fmt.Sprintf("k%d", i), i, time.Minute)
		}(i)
	}
	wg.Wait()

	if got := c.Len(); got != n {
		t.Fatalf("expected Len=%d, got %d", n, got)
	}
}
