// Package cache 你的练习实现：带 TTL 的线程安全键值缓存。
//
// 建议工作流：
//  1. 先看 reference/cache.go，理解每一行为什么这么写
//  2. 关闭 reference/，在这里凭记忆默写
//  3. 跑 go test -v 验证
//  4. 卡住可以回看 reference，但理解后再写
package cache

// TODO: 你需要在这里定义：
//  1. entry 结构体（存值 + 过期时间）
//  2. Cache 结构体（map + 锁）
//  3. NewCache() *Cache
//  4. (c *Cache) Set(key string, value interface{}, ttl time.Duration)
//  5. (c *Cache) Get(key string) (interface{}, bool)
//  6. (c *Cache) Delete(key string)
//  7. (c *Cache) Len() int
//
// 提示：需要 import "sync" 和 "time"
