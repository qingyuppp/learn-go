package main

import (
	"bufio"
	"encoding/json"
	"net"
	"os"
	"testing"
	"time"
)

// TestServer 启动服务端，用客户端发命令，验证响应。
// go test -v 全绿即通过。
func TestServer(t *testing.T) {
	// 启动服务端（用独立 socket 路径避免和 main 冲突）
	sockPath := "/tmp/day03-test.sock"
	os.Remove(sockPath)
	ln, err := net.Listen("unix", sockPath)
	if err != nil {
		t.Fatalf("listen: %v", err)
	}
	t.Cleanup(func() { ln.Close(); os.Remove(sockPath) })

	go func() {
		for {
			conn, err := ln.Accept()
			if err != nil {
				return
			}
			go handleConn(conn)
		}
	}()
	time.Sleep(20 * time.Millisecond) // 等服务端就绪

	// 连接
	conn, err := net.Dial("unix", sockPath)
	if err != nil {
		t.Fatalf("dial: %v", err)
	}
	defer conn.Close()
	scanner := bufio.NewScanner(conn)

	send := func(req Request) Response {
		data, _ := json.Marshal(req)
		conn.Write(append(data, '\n'))
		scanner.Scan()
		var resp Response
		json.Unmarshal(scanner.Bytes(), &resp)
		return resp
	}

	t.Run("ping", func(t *testing.T) {
		resp := send(Request{Cmd: "ping"})
		if resp.Result != "pong" {
			t.Errorf("want pong, got %v", resp.Result)
		}
	})

	t.Run("info_has_fields", func(t *testing.T) {
		resp := send(Request{Cmd: "info"})
		if resp.Error != "" {
			t.Fatalf("unexpected error: %s", resp.Error)
		}
		m, ok := resp.Result.(map[string]interface{})
		if !ok {
			t.Fatalf("result should be a map, got %T", resp.Result)
		}
		for _, field := range []string{"go_version", "os", "time"} {
			if _, ok := m[field]; !ok {
				t.Errorf("info result missing field %q", field)
			}
		}
	})

	t.Run("echo", func(t *testing.T) {
		resp := send(Request{Cmd: "echo", Params: map[string]string{"msg": "hello"}})
		if resp.Result != "hello" {
			t.Errorf("want hello, got %v", resp.Result)
		}
	})

	t.Run("echo_missing_param", func(t *testing.T) {
		resp := send(Request{Cmd: "echo"})
		if resp.Error == "" {
			t.Error("expected error for missing msg param")
		}
	})

	t.Run("unknown_command", func(t *testing.T) {
		resp := send(Request{Cmd: "fly-to-moon"})
		if resp.Error == "" {
			t.Error("expected error for unknown command")
		}
	})
}
