//go:build ignore

package cache

import (
	"sync"
	"time"
)

type entry struct {
	expiresAt time.Time
	value     interface{}
}

type Cache struct {
	mu   sync.RWMutex
	data map[string]entry
}

func NewCache() *Cache {
	return &Cache{
		// make(map[键类型]值类型)
		data: make(map[string]entry),
	}
}

func (c *Cache) Set(key string, value interface{}, ttl time.Duration) {
	c.mu.Lock()
	defer c.mu.Unlock()

	// entry{字段名: 值, ...}，过期时间 = time.Now().Add(ttl)
	c.data[key] = entry{
		value:     value,
		expiresAt: time.Now().Add(ttl),
	}
}

func (c *Cache) Get(key string) (interface{}, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()

	e, ok := c.data[key]
	if !ok {
		return nil, false
	}
	if time.Now().After(e.expiresAt) {
		return nil, false
	}
	return e.value, true
}

func (c *Cache) Delete(key string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	delete(c.data, key)
}

func (c *Cache) Len() int {
	c.mu.RLock()
	defer c.mu.RUnlock()

	now := time.Now()
	count := 0
	// for _, v := range map {}，用 now.After(e.expiresAt) 判断是否过期
	for _, e := range c.data {
		if !now.After(e.expiresAt) {
			count++
		}
	}
	return count
}
