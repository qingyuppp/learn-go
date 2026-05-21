# Day 2: HTTP Client SDK — 给 Todo API 写客户端

## 题目

为一个虚构的 Todo HTTP API 实现 Go SDK 客户端。

服务端假设已存在（你不需要写），提供 4 个接口：

| 方法 | 路径 | 作用 | 鉴权 |
|------|------|------|------|
| GET | `/todos` | 列出所有 todo | 需要 |
| POST | `/todos` | 创建一个 todo | 需要 |
| GET | `/todos/{id}` | 获取单个 todo | 需要 |
| DELETE | `/todos/{id}` | 删除一个 todo | 需要 |

所有请求需要带 `Authorization: Bearer <api_key>` header。

你要实现的 SDK 包要让使用者这样调：

```go
client := todo.NewClient("https://api.example.com", "my-api-key")

// 列出
todos, err := client.ListTodos(ctx)

// 创建
todo, err := client.CreateTodo(ctx, "buy milk")

// 获取
todo, err := client.GetTodo(ctx, 42)

// 删除
err := client.DeleteTodo(ctx, 42)
```

## 为什么有价值

**SDK 的本质就是"HTTP 客户端 + 类型封装"**。理解了自己写的 SDK，你就理解了 OpenSandbox/Stripe/AWS 这些大厂 SDK 的设计。

实习项目对应：
- OpenSandbox 的 Python SDK（`sandbox.get_metrics()` 底层就是 `requests.get(/metrics)`）
- JoyCode 后端调 OpenSandbox 时用的 SDK

## 知识点

| 概念 | 这道题怎么用到 |
|------|--------------|
| `net/http.Client` | 发送 HTTP 请求的核心对象 |
| `http.NewRequestWithContext` | 构造请求（支持 context 取消） |
| `json.Marshal` / `json.NewDecoder` | 请求体/响应体的 JSON 序列化 |
| 自定义 error 类型 | 把 HTTP 状态码包装成语义化错误 |
| `context.Context` | 控制超时和取消 |
| `httptest.NewServer` | 测试 HTTP 客户端不需要真服务器 |
| 工厂函数 `NewClient` | 配置 baseURL、apiKey、超时等 |
| Header 自动注入 | 每次请求自动加 Authorization |

## 文件结构

```
day02-http-client-sdk/
├── README.md          ← 你正在看
├── reference/         ← 参考实现
│   ├── client.go      ← Client 实现（Todo SDK）
│   └── client_test.go ← 测试（用 httptest 模拟服务器）
└── practice/          ← 你的练习场
    ├── client.go      ← 空骨架
    └── client_test.go ← 同样测试
```

## 推荐工作流

### 阶段 1：读（10 min）

```bash
cd reference
go test -v   # 确认参考实现 OK
cat client.go
```

重点理解：
1. `Client` 结构体的字段（为什么需要 baseURL、apiKey、httpClient）
2. `doRequest` 内部方法的作用（封装重复逻辑）
3. 怎么把 API Key 加到每个请求的 Header
4. JSON 编码：用 `bytes.Buffer` + `json.NewEncoder`
5. JSON 解码：直接 `json.NewDecoder(resp.Body).Decode(&out)`
6. 状态码不是 2xx 时怎么转 `APIError`
7. `httptest.NewServer` 怎么模拟一个真实 HTTP 服务

### 阶段 2：抄（15 min）

```bash
cd ../practice
go test -v  # 红 → 绿
```

### 阶段 3：改（5 min）

挑一个变体改：
- 在 `Client` 里加一个 `timeout time.Duration`，让 NewClient 接受这个参数
- 增加 `UpdateTodo(ctx, id, title)` 方法（PATCH 请求）
- 自动重试：5xx 错误自动重试 3 次

## 验证通过

`cd practice && go test -v` 6 个测试 PASS：

```
TestListTodos
TestCreateTodo
TestGetTodo
TestDeleteTodo
TestAuthHeader
TestErrorHandling
```

## 进阶（有时间再做）

- 加请求/响应日志（debug 模式打印）
- 实现"分页"：ListTodos 支持 limit/offset 参数
- 把 baseURL 解析成 `url.URL`，避免拼接 bug
- 仿照 OpenSandbox SDK，加一个 SSE 流式接口（参考 `/metrics/watch`）
