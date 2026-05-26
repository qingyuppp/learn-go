// client.go — 对应 server.go 的协议客户端
//
// 运行方式（先启动 server.go）：go run client.go
//
// 演示：用代码发命令、读响应，理解"协议"是怎么工作的

//go:build ignore

package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"net"
	"os"
)

// 复用 server.go 里的类型定义（实际项目里放在独立 package）
type Request struct {
	Cmd    string            `json:"cmd"`
	Params map[string]string `json:"params,omitempty"`
}

type Response struct {
	Result interface{} `json:"result,omitempty"`
	Error  string      `json:"error,omitempty"`
}

const socketPath = "/tmp/learn-protocol.sock"

func main() {
	// 连接到 Unix socket
	conn, err := net.Dial("unix", socketPath)
	if err != nil {
		fmt.Printf("连接失败（先启动 server.go）: %v\n", err)
		os.Exit(1)
	}
	defer conn.Close()

	scanner := bufio.NewScanner(conn)

	// ============================================================
	// 演示 1：ping — 最简单的协议交互
	// ============================================================
	fmt.Println("=== 发送 ping ===")
	send(conn, Request{Cmd: "ping"})
	resp := receive(scanner)
	fmt.Printf("响应: %v\n\n", resp.Result)

	// ============================================================
	// 演示 2：info — 命令带结构化返回值
	// ============================================================
	fmt.Println("=== 发送 info ===")
	send(conn, Request{Cmd: "info"})
	resp = receive(scanner)
	// Result 是 map，打印每个字段
	if info, ok := resp.Result.(map[string]interface{}); ok {
		for k, v := range info {
			fmt.Printf("  %s: %v\n", k, v)
		}
	}
	fmt.Println()

	// ============================================================
	// 演示 3：echo — 命令带参数
	// ============================================================
	fmt.Println("=== 发送 echo ===")
	send(conn, Request{
		Cmd:    "echo",
		Params: map[string]string{"msg": "hello from client"},
	})
	resp = receive(scanner)
	fmt.Printf("响应: %v\n\n", resp.Result)

	// ============================================================
	// 演示 4：未知命令 — 验证错误处理
	// ============================================================
	fmt.Println("=== 发送未知命令 ===")
	send(conn, Request{Cmd: "fly-to-moon"})
	resp = receive(scanner)
	fmt.Printf("错误: %v\n\n", resp.Error)

	// ============================================================
	// 演示 5：quit
	// ============================================================
	fmt.Println("=== 发送 quit ===")
	send(conn, Request{Cmd: "quit"})
	resp = receive(scanner)
	fmt.Printf("响应: %v\n", resp.Result)
}

func send(conn net.Conn, req Request) {
	data, _ := json.Marshal(req)
	conn.Write(append(data, '\n')) // 加 \n 表示这条消息结束
}

func receive(scanner *bufio.Scanner) Response {
	scanner.Scan() // 阻塞，直到读到一个 \n
	var resp Response
	json.Unmarshal(scanner.Bytes(), &resp)
	return resp
}
