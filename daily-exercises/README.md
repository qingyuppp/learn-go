# Daily Exercises

每天一道实战题，20-30 分钟掌握一个后端/容器开发常用模式。

## 使用方法

```bash
# 1. 进入当天目录
cd dayNN-xxx/

# 2. 看 README.md 理解题目

# 3. 看骨架文件（一般是 main.go 或 xxx.go），实现 TODO 标注的函数

# 4. 跑测试验证
go test -v
```

测试全绿就过关。卡住可以看 README.md 的"提示"部分。

## 题目列表

| Day | 题目 | 知识点 | 状态 |
|-----|------|-------|------|
| 01 | [KV 缓存（带 TTL）](day01-kv-cache-with-ttl/) | struct、map、mutex、time、interface{} | ⏳ 待做 |
| 02 | [HTTP Client SDK](day02-http-client-sdk/) | net/http、json、context、APIError、httptest | ⏳ 待做 |
| 03 | [JSON 协议服务端](day03-json-protocol/) | Unix socket、bufio.Scanner、json、goroutine、dispatch | ⏳ 待做 |

## 难度梯度

- **Day 1-10**：单文件、单个数据结构、3-5 个方法（语法熟练）
- **Day 11-20**：多文件、组合多个组件（项目结构感）
- **Day 21-30**：mini 项目，包含 main + 测试 + README（独立写完整程序）

## 出题原则

1. **后端 + 容器场景为主**：KV 缓存、连接池、HTTP 中间件、worker pool、Pool 模式等
2. **能跑通测试 = 通过**：测试就是题目的精确定义
3. **题目自带知识点说明**：每天 README 告诉你这道题为什么有价值
