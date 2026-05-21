// Package cache 实现一个带 TTL 的线程安全键值缓存（参考实现）。
//
// 这是"参考实现"，每一行都有注释解释为什么这么写。
// 阅读顺序：从 entry 结构体开始，依次往下看 Cache → NewCache → Set → Get → Delete → Len。
package cache

import (
	"sync"
	"time"
)

// entry 是缓存中的一条记录。
// 为什么单独定义？因为缓存不只存"值"，还要存"何时过期"。把这两个信息绑定在一起最自然。
type entry struct {
	value     interface{} // 任意类型的值。interface{}（也叫 any）是 Go 的"万能类型"
	expiresAt time.Time   // 绝对过期时间点（不是 duration），用 time.Now().After(expiresAt) 判断是否过期
}

// Cache 是一个线程安全的 KV 缓存。
// 为什么字段都小写？小写字段不导出，外部不能直接访问，必须通过方法操作。这是封装。
type Cache struct {
	mu   sync.RWMutex     // 读写锁，读多写少场景比 Mutex 性能好。Get 用 RLock，Set/Delete 用 Lock。
	data map[string]entry // 真正存数据的 map。key 是 string，value 是 entry。
}

// NewCache 构造一个新的 Cache。
// 为什么不直接 var c Cache？因为 map 字段必须显式初始化（make），否则 nil map 写入会 panic。
// 工厂函数模式保证使用者不会拿到一个"半初始化"的对象。
func NewCache() *Cache {
	return &Cache{
		data: make(map[string]entry), // map 必须 make 才能用
		// mu 不用初始化，零值就是可用的 sync.RWMutex
	}
}

// Set 存入 key-value，ttl 后过期。
//
// 接收者为什么是 *Cache 而不是 Cache？
//   - Cache 内部有 mu（不可复制）和 data（要修改），必须用指针接收者
//   - 经验法则：只要方法会修改字段，或者 struct 里有锁/map/slice，都用指针接收者
func (c *Cache) Set(key string, value interface{}, ttl time.Duration) {
	c.mu.Lock()         // 写操作，加写锁
	defer c.mu.Unlock() // defer 保证函数退出时一定释放锁，即使中间 panic 也释放

	// 关键：过期时间 = 当前时间 + ttl
	c.data[key] = entry{
		value:     value,
		expiresAt: time.Now().Add(ttl),
	}
}

// Get 读取 key 对应的值。
// 如果 key 不存在或已过期，返回 (nil, false)。
// 如果发现过期，顺手清理掉。
//
// Go 惯用法：(value, ok) 返回模式。和 map 读取一样：v, ok := m["k"]
func (c *Cache) Get(key string) (interface{}, bool) {
	c.mu.Lock()         // 这里直接用写锁。原因：可能要删过期的，需要写权限。
	defer c.mu.Unlock() // 想优化的话可以先 RLock 读，发现过期再升级成 Lock 重新读+删，但代码会复杂。

	e, ok := c.data[key]
	if !ok {
		// key 不存在
		return nil, false
	}

	if time.Now().After(e.expiresAt) {
		// 过期了：删除并返回 false
		delete(c.data, key)
		return nil, false
	}

	return e.value, true
}

// Delete 删除一个 key（如果存在）。
// 即使 key 不存在，delete() 也是安全的，不会 panic。
func (c *Cache) Delete(key string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	delete(c.data, key)
}

// Len 返回当前未过期的 entry 数量。
// 注意：只统计，不修改 map（不顺手删过期的）。
// 用 RLock 而不是 Lock，因为不修改数据，可以让多个 goroutine 同时调 Len。
func (c *Cache) Len() int {
	c.mu.RLock()
	defer c.mu.RUnlock()

	now := time.Now() // 取一次 time.Now() 比循环里每次取效率高
	count := 0
	for _, e := range c.data {
		if !now.After(e.expiresAt) {
			count++
		}
	}
	return count
}
