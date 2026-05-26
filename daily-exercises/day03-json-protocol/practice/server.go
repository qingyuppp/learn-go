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

// Request 是客户端发来的消息。
// 协议约定：JSON 对象 + '\n' 结尾。
type Request struct {
	Cmd    string            `json:"cmd"`
	Params map[string]string `json:"params,omitempty"`
}

// Response 是服务端回给客户端的消息。
type Response struct {
	Result interface{} `json:"result,omitempty"`
	Error  string      `json:"error,omitempty"`
}

const SocketPath = "/tmp/day03-protocol.sock"

func main() {
	os.Remove(SocketPath)
	ln, err := net.Listen("unix", SocketPath)
	if err != nil {
		fmt.Println("listen error:", err)
		os.Exit(1)
	}
	defer ln.Close()
	defer os.Remove(SocketPath)

	fmt.Println("server listening on", SocketPath)
	for {
		conn, err := ln.Accept()
		if err != nil {
			continue
		}
		go handleConn(conn)
	}
}

// handleConn 处理一个连接的全部请求。
// 按行读取（bufio.Scanner），每行是一条 JSON 消息。
func handleConn(conn net.Conn) {
	defer conn.Close()
	scanner := bufio.NewScanner(conn)
	for scanner.Scan() {
		var req Request
		if err := json.Unmarshal(scanner.Bytes(), &req); err != nil {
			sendResponse(conn, Response{Error: "invalid json"})
			continue
		}
		sendResponse(conn, dispatch(req))
	}
}

// dispatch 根据 req.Cmd 返回对应的 Response。
//
// 需要实现以下命令：
//   - "ping"  → Result: "pong"
//   - "info"  → Result: map，包含 "go_version"、"os"、"time" 三个字段
//   - "echo"  → Result: req.Params["msg"]；若缺少 msg 则 Error: "missing param: msg"
//   - 其他    → Error: "unknown command: <cmd>"
//
// TODO: 实现 switch 分支
func dispatch(req Request) Response {
	switch req.Cmd {
	case "ping":
		// TODO
	case "info":
		// TODO：返回包含 go_version / os / time 三个字段的 map
		_ = runtime.Version() // 提示：用这个拿 Go 版本
		_ = time.Now()        // 提示：用这个拿当前时间
	case "echo":
		// TODO：从 req.Params["msg"] 取消息原路返回；msg 不存在时返回 error
	default:
		// TODO：返回 "unknown command: <cmd>" 错误
	}
	return Response{} // 删掉这行，替换成各分支的返回值
}

// sendResponse 把 resp 序列化成 JSON，加 '\n' 写入 conn。
//
// TODO: 实现这个函数
// 提示：json.Marshal → append data '\n' → conn.Write
func sendResponse(conn net.Conn, resp Response) {
}
