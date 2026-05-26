// server.go — 一个最小的 JSON 协议服务端
//
// 运行方式：go run server.go
// 然后另开一个终端运行：go run client.go
//
//go:build ignore
// 这个服务端监听一个 Unix socket，等待客户端发 JSON 命令，
// 处理后返回 JSON 响应——跟 QEMU Monitor Protocol (QMP) 思路完全一样。

package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"net"
	"os"
	"runtime"
	"time"
)

// ============================================================
// 第一步：定义"协议"
//
// 协议 = 双方事先约定好的"消息格式"。
// 如果没有约定，A 发了 100 字节，B 不知道这 100 字节从哪到哪是一条消息。
//
// 我们的协议规则（类 QMP）：
//   1. 每条消息是一个 JSON 对象
//   2. 消息以 '\n' 结尾（用换行符分隔消息边界）
//   3. 请求有 "cmd" 字段，响应有 "result" 或 "error" 字段
//
// 这种设计叫 "newline-delimited JSON"（NDJSON），QMP 就是这么做的。
// ============================================================

// Request 是客户端发给服务端的消息格式
type Request struct {
	Cmd    string            `json:"cmd"`              // 命令名，如 "ping" / "info"
	Params map[string]string `json:"params,omitempty"` // 可选参数
}

// Response 是服务端回给客户端的消息格式
type Response struct {
	Result interface{} `json:"result,omitempty"` // 成功时的返回值
	Error  string      `json:"error,omitempty"`  // 失败时的错误信息
}

const socketPath = "/tmp/learn-protocol.sock"

func main() {
	// 清理旧的 socket 文件（Unix socket 不会自动删除）
	os.Remove(socketPath)

	// 监听 Unix domain socket
	// Unix socket = 本机进程间通信，比 TCP 快，不需要网卡
	// containerd 和 containerd-shim 之间就用 Unix socket 通信
	listener, err := net.Listen("unix", socketPath)
	if err != nil {
		fmt.Printf("监听失败: %v\n", err)
		os.Exit(1)
	}
	defer listener.Close()
	defer os.Remove(socketPath)

	fmt.Printf("服务端启动，监听 %s\n", socketPath)
	fmt.Println("等待客户端连接...")

	for {
		// Accept 阻塞等待新连接
		conn, err := listener.Accept()
		if err != nil {
			fmt.Printf("Accept 失败: %v\n", err)
			continue
		}
		// 每个连接用一个 goroutine 处理，防止一个慢客户端阻塞所有人
		// 这是网络服务端的标准模式
		go handleConn(conn)
	}
}

// handleConn 处理单个连接的所有消息
func handleConn(conn net.Conn) {
	defer conn.Close()
	remoteAddr := conn.RemoteAddr().String()
	fmt.Printf("[%s] 新连接\n", remoteAddr)

	// bufio.Scanner 帮我们按 '\n' 分割消息
	// 这就是"消息边界"的解决方案：遇到 \n 就认为一条消息结束了
	scanner := bufio.NewScanner(conn)

	for scanner.Scan() {
		line := scanner.Bytes()
		if len(line) == 0 {
			continue
		}

		// 1. 反序列化：把 JSON 字节解析成 Go 结构体
		var req Request
		if err := json.Unmarshal(line, &req); err != nil {
			sendResponse(conn, Response{Error: fmt.Sprintf("JSON 格式错误: %v", err)})
			continue
		}

		fmt.Printf("[%s] 收到命令: %s\n", remoteAddr, req.Cmd)

		// 2. 根据命令分发处理
		resp := dispatch(req)

		// 3. 序列化：把 Go 结构体变成 JSON 字节，加 \n 发回去
		sendResponse(conn, resp)
	}

	fmt.Printf("[%s] 连接断开\n", remoteAddr)
}

// dispatch 根据命令名找对应的处理函数
// 这个模式在真实协议实现里叫"命令分发表"（command dispatch table）
func dispatch(req Request) Response {
	switch req.Cmd {

	case "ping":
		// 最简单的心跳命令，用来确认连接活着
		// QMP 也有 qmp_capabilities 这种握手命令
		return Response{Result: "pong"}

	case "info":
		// 返回一些运行时信息
		// 类似 QMP 的 "query-version" 命令
		return Response{Result: map[string]interface{}{
			"go_version": runtime.Version(),
			"os":         runtime.GOOS,
			"arch":       runtime.GOARCH,
			"goroutines": runtime.NumGoroutine(),
			"time":       time.Now().Format(time.RFC3339),
		}}

	case "echo":
		// 把 params["msg"] 原样返回
		msg, ok := req.Params["msg"]
		if !ok {
			return Response{Error: "缺少参数 msg"}
		}
		return Response{Result: msg}

	case "quit":
		// 客户端请求断开
		return Response{Result: "bye"}

	default:
		// 未知命令——返回错误而不是 panic
		// 这是协议设计的基本原则：对未知命令有明确的错误响应
		return Response{Error: fmt.Sprintf("未知命令: %q", req.Cmd)}
	}
}

// sendResponse 把响应序列化成 JSON 并加 \n 写入连接
func sendResponse(conn net.Conn, resp Response) {
	data, err := json.Marshal(resp)
	if err != nil {
		fmt.Printf("序列化响应失败: %v\n", err)
		return
	}
	// 关键：加 '\n' 作为消息结束标记
	// 客户端用 bufio.Scanner 按 '\n' 分割，才能正确读到完整消息
	conn.Write(append(data, '\n'))
}
