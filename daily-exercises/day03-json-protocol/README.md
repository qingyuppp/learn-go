# Day 3: JSON 协议服务端

## 题目

实现一个基于 **Unix socket + newline-delimited JSON** 的协议服务端，处理以下命令：

| 命令 | 参数 | 成功响应 | 失败响应 |
|------|------|---------|---------|
| `ping` | 无 | `{"result":"pong"}` | — |
| `info` | 无 | `{"result":{"go_version":"...","os":"...","time":"..."}}` | — |
| `echo` | `msg` | `{"result":"<msg 原文>"}` | `msg` 缺失时 error |
| 其他 | — | — | `{"error":"unknown command: <cmd>"}` |

每条消息是一个 JSON 对象，以 `\n` 结尾。

## 为什么有价值

这就是 **QEMU Monitor Protocol（QMP）** 的简化版：containerd 里的 kata-shim 用同样的方式控制 QEMU（JSON + `\n` + Unix socket）。实现这道题 = 亲手搭出了 QMP 的骨架。

理解这道题之后，去读 kata-containers 的 `src/runtime/pkg/qemu/qmp.go` 会发现结构几乎完全一样。

## 知识点

| 概念 | 这道题怎么用到 |
|------|--------------|
| `net.Listen("unix", ...)` | 监听 Unix socket |
| `bufio.Scanner` | 按 `\n` 切割消息边界 |
| `encoding/json` Marshal/Unmarshal | 消息序列化/反序列化 |
| `goroutine` | 每个连接独立 goroutine，并发处理 |
| `switch` | 命令分发（dispatch）|
| `map[string]interface{}` | 构造不定字段的响应体 |

## 文件结构

```
day03-json-protocol/
├── README.md
├── practice/          ← 你来实现
│   ├── server.go      ← 骨架：填 dispatch() 和 sendResponse()
│   └── server_test.go ← 5 个测试，全绿即通过
└── reference/         ← 参考实现（先尝试自己写，卡了再看）
    ├── server.go      ← 完整 server（go run server.go 独立运行）
    └── client.go      ← 配套 client（go run client.go 测试交互）
```

## 推荐工作流

```bash
# 1. 看 practice/server.go 里的 TODO，实现 dispatch 和 sendResponse
# 2. 验证
cd practice && go test -v

# 3. 全绿后，用 reference 的完整版体验端到端交互
cd ../reference
go run server.go &          # 后台起 server
go run client.go            # 看完整交互输出
kill %1                     # 关掉 server
```

## 提示（只在卡住时看）

<details>
<summary>dispatch 怎么写</summary>

```go
case "ping":
    return Response{Result: "pong"}
case "info":
    return Response{Result: map[string]interface{}{
        "go_version": runtime.Version(),
        "os":         runtime.GOOS,
        "time":       time.Now().Format(time.RFC3339),
    }}
```
</details>

<details>
<summary>sendResponse 怎么写</summary>

```go
data, _ := json.Marshal(resp)
conn.Write(append(data, '\n'))
```

关键：`'\n'` 是消息边界，客户端的 `bufio.Scanner` 靠它知道一条消息在哪里结束。
</details>

## 延伸阅读

- `05-protocol-and-net/containerd_example.go` — QMP / ttrpc 对照解读
- [kata-containers qmp.go](https://github.com/kata-containers/kata-containers/blob/main/src/runtime/pkg/qemu/qmp.go) — 真实 QMP 实现
- [containerd/ttrpc](https://github.com/containerd/ttrpc) — 二进制协议实现
