// 本文件把 main.go 里实现的"玩具协议"和 containerd / kata-containers 里的真实协议做对照。
// 不可编译运行，全部是注释形式的代码解读。
//
// 深入阅读：learn-containerd（独立仓库，待建）

//go:build ignore

package main

// ============================================================
// 对照点 1：QMP（QEMU Monitor Protocol）
//
// kata-shim 用 QMP 控制 QEMU：热插内存、热插设备、查询 VM 状态。
// 来源：https://github.com/kata-containers/kata-containers/blob/main/src/runtime/pkg/qemu/qmp.go
// ============================================================
//
// QMP 协议规则（和 main.go 第三部分实现的几乎完全一样）：
//   - 传输层：Unix socket（kata-shim 启动 QEMU 时指定 -qmp unix:/path/qmp.sock）
//   - 消息格式：JSON 对象
//   - 消息边界：每条消息以 '\n' 结尾（newline-delimited）
//   - 握手：连接后服务端先发一条 capabilities 消息
//
// 真实 QMP 交互（可用 nc 或 socat 直接测试）：
//
//   连接后 QEMU 主动发：
//     {"QMP": {"version": {...}, "capabilities": [...]}}
//
//   客户端发握手：
//     {"execute": "qmp_capabilities"}
//   QEMU 回：
//     {"return": {}}
//
//   查内存：
//     {"execute": "query-memory-size-summary"}
//   QEMU 回：
//     {"return": {"base-memory": 536870912, "plugged-memory": 0}}
//
//   热插内存（virtio-mem-pci 设备）：
//     {"execute": "object-add", "arguments": {"qom-type": "memory-backend-ram", "id": "mem1", "size": 134217728}}
//     {"execute": "device_add",  "arguments": {"driver": "virtio-mem-pci", "memdev": "mem1", "id": "vm1"}}
//
// 对比 main.go 第三部分：
//   我们：{"cmd": "info"}         → {"result": {...}}
//   QMP:  {"execute": "query-..."} → {"return": {...}}
//   字段名不同，但协议三要素（传输/格式/边界）完全一样。

// ============================================================
// 对照点 2：ttrpc（containerd ↔ kata-shim 的通信协议）
//
// containerd 用 ttrpc 调用 kata-shim，让它起容器、停容器、查状态。
// 来源：https://github.com/containerd/ttrpc
// ============================================================
//
// ttrpc 是精简版 gRPC，专为低内存环境设计（kata-agent 跑在 VM 里，内存紧张）：
//
//   玩具协议（main.go）    ttrpc
//   ──────────────         ──────────────────────
//   JSON 序列化            protobuf 序列化（体积小 3-10 倍，解析更快）
//   文本（可读）           二进制（需要 proto 定义才能解码）
//   \n 分隔消息            固定 10 字节 header（含 4 字节长度）+ payload
//   单路（一问一答）        一个连接上并发多个请求（stream-id 区分）
//   Unix socket            vsock（VM ↔ 宿主机的专用 socket）
//
// 真实 ttrpc 消息格式（简化）：
//
//   ┌──────────────────────────────────────────────┐
//   │ Header（10 字节）                             │
//   │   length    (4 bytes) = payload 字节数        │
//   │   stream-id (4 bytes) = 请求 ID（用于并发）   │
//   │   type      (1 byte)  = Request / Response    │
//   │   flags     (1 byte)                          │
//   ├──────────────────────────────────────────────┤
//   │ Payload（protobuf 序列化）                    │
//   └──────────────────────────────────────────────┘
//
// 为什么用长度前缀而不是 \n 分隔？
//   → protobuf 是二进制，payload 里可能包含 0x0a（= '\n'）字节
//   → 长度前缀 100% 准确：先读 4 字节得到长度，再读对应字节

// ============================================================
// 对照点 3：消息边界（framing）三种方案
// ============================================================
//
//   方案 A：分隔符         适合文本协议（QMP / HTTP headers / Redis RESP）
//     发：{"cmd":"ping"}\n{"cmd":"info"}\n
//     收：bufio.Scanner 按 \n 切割
//     限制：消息内容不能含分隔符（或需转义）
//
//   方案 B：长度前缀       适合二进制协议（ttrpc / gRPC / protobuf over TCP）
//     发：[0,0,0,4]ping[0,0,0,4]info
//     收：先读 4 字节得长度，再读对应字节
//     限制：发送前必须知道完整长度（需先 buffer 整条消息）
//
//   方案 C：固定长度       适合硬件协议、游戏同步包
//     每条消息永远 N 字节，超出截断、不足补零
//
//   HTTP 是混合：header 用 \r\n\r\n 分隔，body 用 Content-Length 指定长度。

// ============================================================
// 对照点 4：vsock — 跨 VM 的 socket
// ============================================================
//
//   Unix socket（本文件玩具协议、QMP）：仅限同一宿主机上的进程通信
//   vsock：跨 VM 边界通信，kata-shim（宿主机） ↔ kata-agent（VM 内部）
//
//   Go 中用法几乎一样，只换 network 类型：
//
//     // Unix socket
//     conn, _ := net.Dial("unix", "/tmp/learn-protocol.sock")
//
//     // vsock（第三方库 github.com/mdlayher/vsock）
//     conn, _ := vsock.Dial(vmCID, port, nil)  // vmCID = VM 的 Context ID
//
//   协议层代码（JSON 解析、dispatch、sendResponse）完全不需要改。
//   这就是"协议分层"：传输层和应用层解耦。

// ============================================================
// 总结：containerd 调 kata 的完整调用链
// ============================================================
//
//   containerd
//       │  ttrpc（protobuf + 长度前缀 + Unix socket）
//       ↓
//   kata-shim（宿主机进程）
//       │  QMP（JSON + \n 分隔 + Unix socket）
//       ↓
//   QEMU（hypervisor）
//       │  vsock
//       ↓
//   kata-agent（VM 内部）
//       │  OCI runtime（runc）
//       ↓
//   容器进程
//
// 每一层都是一个协议，三要素各不相同，但设计思想完全一样。
// 进一步学习：see learn-containerd（待建）
