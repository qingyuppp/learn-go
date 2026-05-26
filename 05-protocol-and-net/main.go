package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"os"
	"strings"
)

func main() {
	fmt.Println("=== 第一部分：JSON 序列化 ===")
	part1JSON()

	fmt.Println("\n=== 第二部分：HTTP Server 是一种协议 ===")
	part2HTTP()

	fmt.Println("\n=== 第三部分：自己实现一个协议 ===")
	part3CustomProtocol()
}

// ============================================================
// 第一部分：Go by Example — JSON
// 来源：https://gobyexample-cn.github.io/json
//
// JSON 是"消息格式"：约定了数据怎么表达成字节。
// 协议 = 消息格式 + 传输规则。这里先只学消息格式部分。
// ============================================================

func part1JSON() {
	// 1a. 编码：Go 数据结构 → JSON 字节
	// json tag 控制字段名（和 K8s types.go 里的 tag 一样）
	type Command struct {
		Cmd    string            `json:"cmd"`
		Params map[string]string `json:"params,omitempty"` // omitempty：空时不输出
	}

	cmd := Command{
		Cmd:    "ping",
		Params: map[string]string{"from": "client"},
	}
	data, _ := json.Marshal(cmd)
	fmt.Println("编码结果:", string(data))
	// {"cmd":"ping","params":{"from":"client"}}

	// 1b. 解码：JSON 字节 → Go 数据结构
	raw := `{"cmd":"info","params":{"target":"vm1"}}`
	var decoded Command
	json.Unmarshal([]byte(raw), &decoded)
	fmt.Printf("解码结果: cmd=%s, params=%v\n", decoded.Cmd, decoded.Params)

	// 1c. 解码成 map（不知道结构时用）
	// QMP 收到 QEMU 的响应时就是这样——结构不固定，先解成 map 再取字段
	payload := `{"result":{"memory_total":536870912,"memory_used":26529792}}`
	var generic map[string]interface{}
	json.Unmarshal([]byte(payload), &generic)
	if result, ok := generic["result"].(map[string]interface{}); ok {
		fmt.Printf("内存总量: %.0f bytes\n", result["memory_total"])
	}

	// 1d. 流式编码（写到 io.Writer）
	// json.NewEncoder 直接写到连接/文件，不需要先生成字节再写
	// 后面的协议实现会用到这个
	enc := json.NewEncoder(os.Stdout)
	enc.Encode(map[string]string{"status": "ok"}) // 自动加 \n
}

// ============================================================
// 第二部分：Go by Example — HTTP Server
// 来源：https://gobyexample-cn.github.io/http-servers
//
// HTTP 本身就是一种协议：
//   - 传输层：TCP socket
//   - 消息格式：文本（请求行 + headers + body）
//   - 消息边界：\r\n\r\n 分隔 header 和 body，Content-Length 指定 body 长度
//
// Go 的 net/http 把这些细节全包了，你只用写处理函数。
// 理解 HTTP 是"协议的一种"，之后自己实现协议就是把这件事手工做一遍。
// ============================================================

func part2HTTP() {
	// 注册处理函数（Go by Example 原版写法）
	mux := http.NewServeMux()

	// handler 函数 = 协议的"命令处理器"
	// 对应我们自定义协议里的 dispatch() 函数
	mux.HandleFunc("/ping", func(w http.ResponseWriter, r *http.Request) {
		// w 是"往客户端写响应"的接口，底层是 TCP 连接
		// r 是"客户端发来的请求"，已经解析好了
		fmt.Fprintf(w, `{"result":"pong"}`)
	})

	mux.HandleFunc("/info", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{
			"server":  "learn-protocol",
			"version": "v1",
		})
	})

	// 在后台起服务（这里用 goroutine 避免阻塞主流程）
	go func() {
		http.ListenAndServe(":18090", mux)
	}()

	// 用 HTTP 客户端请求自己（模拟客户端 ↔ 服务端的完整交互）
	fmt.Println("HTTP /ping →", httpGet("http://localhost:18090/ping"))
	fmt.Println("HTTP /info →", httpGet("http://localhost:18090/info"))

	// ──────────────────────────────────────────────────────
	// HTTP 的"协议三要素"对照：
	//
	//   传输层：TCP（net.Listen("tcp", ":8090")）
	//   消息格式：文本，"GET /ping HTTP/1.1\r\nHost: ...\r\n\r\n"
	//   消息边界：\r\n\r\n 分隔请求头和 body；Content-Length 指定 body 长度
	//
	// Go 的 net/http 把这三件事都帮你做了。
	// 如果要实现自己的协议，就要把这三件事自己写出来。
	// ──────────────────────────────────────────────────────
}

func httpGet(url string) string {
	resp, err := http.Get(url)
	if err != nil {
		return fmt.Sprintf("(服务还没起来: %v)", err)
	}
	defer resp.Body.Close()
	var sb strings.Builder
	bufio.NewReader(resp.Body).WriteTo(&sb)
	return strings.TrimSpace(sb.String())
}

// ============================================================
// 第三部分：自己实现一个最小协议
//
// 把第一部分的 JSON 和第二部分的"协议三要素"结合起来，
// 不用 net/http，直接操作 net.Conn，实现一个类 QMP 的协议。
//
// 协议三要素：
//   1. 传输层：Unix socket（进程间通信，比 TCP 快，不需要网卡）
//   2. 消息格式：JSON（用第一部分学的 encoding/json）
//   3. 消息边界：'\n'（每条消息末尾加换行符，接收方按行读）
// ============================================================

const socketPath = "/tmp/learn-protocol-main.sock"

type Request struct {
	Cmd    string            `json:"cmd"`
	Params map[string]string `json:"params,omitempty"`
}

type Response struct {
	Result interface{} `json:"result,omitempty"`
	Error  string      `json:"error,omitempty"`
}

func part3CustomProtocol() {
	// 清理旧 socket
	os.Remove(socketPath)

	// ── 服务端：一个 goroutine ──
	ready := make(chan struct{})
	go runServer(ready)
	<-ready // 等服务端准备好再让客户端连

	// ── 客户端：和服务端交互 ──
	conn, err := net.Dial("unix", socketPath)
	if err != nil {
		fmt.Println("连接失败:", err)
		return
	}
	defer conn.Close()

	scanner := bufio.NewScanner(conn)

	// 发 ping
	sendMsg(conn, Request{Cmd: "ping"})
	resp := recvMsg(scanner)
	fmt.Printf("ping → %v\n", resp.Result)

	// 发 echo（带参数）
	sendMsg(conn, Request{Cmd: "echo", Params: map[string]string{"msg": "hello"}})
	resp = recvMsg(scanner)
	fmt.Printf("echo → %v\n", resp.Result)

	// 发未知命令（验证错误处理）
	sendMsg(conn, Request{Cmd: "unknown"})
	resp = recvMsg(scanner)
	fmt.Printf("unknown → error: %v\n", resp.Error)

	os.Remove(socketPath)

	// ──────────────────────────────────────────────────────
	// 和 HTTP / QMP / ttrpc 的对比：
	//
	//   这里            HTTP（net/http）     QMP              ttrpc
	//   ─────────       ───────────────     ─────────        ──────────
	//   Unix socket     TCP                 Unix socket      vsock
	//   JSON + \n       文本 + \r\n\r\n     JSON + \n        protobuf + 4字节长度
	//   手写 dispatch   HandleFunc          {"execute":"..."}proto 定义 RPC
	//
	// 本质完全一样，只是"三要素"的具体选择不同。
	// 理解了这里，再看 opensandbox_example.go 里 QMP/ttrpc 的代码就直接能懂了。
	// ──────────────────────────────────────────────────────
}

func runServer(ready chan struct{}) {
	ln, _ := net.Listen("unix", socketPath)
	close(ready) // 通知主 goroutine 服务端已就绪
	conn, _ := ln.Accept()
	defer conn.Close()
	defer ln.Close()

	scanner := bufio.NewScanner(conn)
	for scanner.Scan() {
		var req Request
		if err := json.Unmarshal(scanner.Bytes(), &req); err != nil {
			sendMsg(conn, Response{Error: "bad json"})
			continue
		}
		switch req.Cmd {
		case "ping":
			sendMsg(conn, Response{Result: "pong"})
		case "echo":
			sendMsg(conn, Response{Result: req.Params["msg"]})
		default:
			sendMsg(conn, Response{Error: fmt.Sprintf("unknown cmd: %s", req.Cmd)})
		}
	}
}

func sendMsg(conn net.Conn, v interface{}) {
	data, _ := json.Marshal(v)
	conn.Write(append(data, '\n')) // \n = 消息边界
}

func recvMsg(scanner *bufio.Scanner) Response {
	scanner.Scan()
	var resp Response
	json.Unmarshal(scanner.Bytes(), &resp)
	return resp
}
