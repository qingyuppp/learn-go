# Day 1: KV 缓存（带 TTL）

## 题目

实现一个**线程安全的内存键值缓存**，支持：
- 存入键值对，并指定过期时间（TTL）
- 读取键值，过期或不存在则返回 `(nil, false)`
- 删除键
- 查询当前未过期的 entry 数量

## 为什么有价值

后端最常用的数据结构之一。OpenSandbox 的 Pool 控制器里就有类似的"已分配沙箱"缓存。Redis、本地缓存（go-cache）、SDK 客户端的请求缓存，本质都是这个。

## 知识点

| 概念 | 这道题怎么用到 |
|------|--------------|
| `struct` | 定义 Cache 和 entry 结构体 |
| `map` | 内部存储 key → entry 的映射 |
| `sync.RWMutex` | 多 goroutine 并发读写时保证安全 |
| `time.Time` 与 `time.Duration` | 计算过期时间 |
| `interface{}` (any) | value 可以是任意类型 |
| 指针接收者 | 修改 Cache 内部状态必须用 `(c *Cache)` |
| `defer` | 保证锁一定被释放 |

## 文件结构

```
day01-kv-cache-with-ttl/
├── README.md          ← 你正在看
├── reference/         ← 参考实现（先读这里）
│   ├── cache.go       ← 完整代码，每一行有注释
│   └── cache_test.go  ← 跑通验证 reference 是对的
└── practice/          ← 你的练习场
    ├── cache.go       ← 空骨架，你来写
    └── cache_test.go  ← 同样的测试，验证你写的代码
```

## 推荐工作流

### 阶段 1：读（10 min）

```bash
cd reference
go test -v   # 先跑通，确认环境 OK
cat cache.go # 一行一行读，理解每个设计决策为什么这么做
```

重点理解：
1. 为什么要单独定义 `entry` 结构体？
2. 为什么 Cache 字段都小写（不导出）？
3. 为什么 `NewCache()` 返回指针？
4. 为什么 Set/Delete 用 `Lock()`，Len 用 `RLock()`？
5. 为什么所有方法的接收者都是 `*Cache`？
6. `defer c.mu.Unlock()` 的作用？

### 阶段 2：抄（15 min）

```bash
cd ../practice
# 关掉 reference 文件（或新开终端在 reference 看，practice 这边只看本文件）
# 凭记忆把 cache.go 写出来
go test -v  # 反复跑，红灯改到绿灯
```

如果中间忘了细节，回去看 reference 一行，理解后再回来写。**不要直接复制粘贴**——手敲一遍肌肉记忆更牢。

### 阶段 3：改（5 min）

挑一个变体改：
- 把 `RWMutex` 改成 `Mutex`，看测试还能不能过
- 在 Set 时如果 key 已存在，返回旧值
- 加一个 `Keys()` 方法返回所有未过期的 key
- 加一个 `Flush()` 方法清空所有数据

## 验证通过

`cd practice && go test -v` 显示 6 个 `PASS` 就过关：

```
=== RUN   TestSetAndGet
--- PASS: TestSetAndGet
=== RUN   TestGetMissingKey
--- PASS: TestGetMissingKey
=== RUN   TestExpiration
--- PASS: TestExpiration
=== RUN   TestDelete
--- PASS: TestDelete
=== RUN   TestLen
--- PASS: TestLen
=== RUN   TestConcurrentSet
--- PASS: TestConcurrentSet
PASS
```

## 进阶（有时间再做）

- 加一个 `cleanup()` 后台 goroutine 定时清理过期 entry
- 用 `context.Context` 控制 cleanup 退出
- 加一个 `OnEviction` 回调，每次过期触发
- 用 LRU 算法在 Len 超过上限时驱逐最老的
