// Package cache 你的练习实现：带 TTL 的线程安全键值缓存。
//
// 建议工作流：
//  1. 先看 reference/cache.go，理解每一行为什么这么写
//  2. 关闭 reference/，在这里凭记忆默写
//  3. 跑 go test -v 验证
//  4. 卡住可以回看 reference，但理解后再写
package cache

import (
	"sync"
	"time"
)

// TODO: 你需要在这里定义：
//  1. entry 结构体（存值 + 过期时间）
type entry struct {
	expiresAt time.Time
	value interface{}
}
//  2. Cache 结构体（map + 锁）
type Cache struct {
	mu sync.RWMutex
	data map[string]entry
}
//  3. NewCache() *Cache
func NewCache() *Cache {
	return &Cache{
		data: make(map[string]entry),
	}
}
//  4. (c *Cache) Set(key string, value interface{}, ttl time.Duration)
func (c *Cache) Set(key string, value interface{}, ttl time.Duration) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.data[key] = entry{
		value: value,
		expiresAt: time.Now().Add(ttl),
	}
}
//  5. (c *Cache) Get(key string) (interface{}, bool)
func (c *Cache) Get(key string) (interface{}, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()

	e, ok := c.data[key]
	if !ok {
		return	nil,false
	}
	if time.Now().After(e.expiresAt) {
		delete(c.data, key)
		return	nil,false
	}
	return e.value, true
}
//  6. (c *Cache) Delete(key string)
func (c *Cache) Delete(key string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	delete(c.data, key)
}
//  7. (c *Cache) Len() int
func (c *Cache) Len() int {
	c.mu.RLock()
	defer c.mu.RUnlock()

	now := time.Now()
	count := 0

	for _, e := range c.data {
		if !now.After(e.expiresAt){
			count++
		}
	}
	return count
}

// 提示：需要 import "sync" 和 "time"