// 本文件把上面实现的"玩具协议"和 containerd/kata 里的真实协议做对照。
// 不可编译运行，全部是注释形式的代码解读。

//go:build ignore

package main

// ============================================================
// 对照点 1：QMP（QEMU Monitor Protocol）
//
// kata-shim 用 QMP 控制 QEMU：热插内存、热插设备、查询 VM 状态。
// 来源：kata-containers/src/runtime/pkg/qemu/qmp.go
// ============================================================
//
// QMP 协议规则（和我们实现的几乎完全一样）：
//   - 传输层：Unix socket（kata-shim 启动 QEMU 时指定 -qmp unix:/path/qmp.sock）
//   - 消息格式：JSON 对象
//   - 消息边界：每条消息以 '\n' 结尾（newline-delimited）
//   - 握手：连接后服务端先发一条 capabilities 消息
//
// 真实的 QMP 交互（你可以用 nc 或 socat 直接测试）：
//
// 1. 连接后 QEMU 主动发：
//    {"QMP": {"version": {...}, "capabilities": [...]}}
//
// 2. 客户端发握手命令（告诉 QEMU"我准备好了"）：
//    {"execute": "qmp_capabilities"}
//
// 3. QEMU 回：
//    {"return": {}}
//
// 4. 之后就可以发任意命令，比如查内存：
//    {"execute": "query-memory-size-summary"}
//    QEMU 回：
//    {"return": {"base-memory": 536870912, "plugged-memory": 0}}
//
// 5. 热插内存（kata 的内存热插拔就是这么做的，你之前撞过这个 bug）：
//    {"execute": "object-add", "arguments": {"qom-type": "memory-backend-ram", "id": "mem1", "size": 134217728}}
//    {"execute": "device_add", "arguments": {"driver": "virtio-mem-pci", "memdev": "mem1", "id": "vm1"}}
//
// 对比我们的实现：
//   我们：{"cmd": "info"}         → {"result": {...}}
//   QMP:  {"execute": "query-..."} → {"return": {...}}
//   字段名不同，但结构完全一样。

// ============================================================
// 对照点 2：ttrpc（containerd ↔ kata-shim 的通信协议）
//
// containerd 用 ttrpc 调用 kata-shim，让它起容器、停容器、查状态。
// 来源：https://github.com/containerd/ttrpc
// ============================================================
//
// ttrpc 是一个精简版 gRPC，专为低内存环境设计（kata-agent 跑在 VM 里，内存很紧）：
//
// 和我们"玩具协议"的对比：
//
//   玩具协议           ttrpc
//   ──────────         ──────
//   JSON 序列化        protobuf 序列化（体积小 3-10 倍，解析更快）
//   文本（可读）       二进制（不可读，需要 proto 定义才能解码）
//   \n 分隔消息        固定 header（4 字节长度）+ payload
//   无连接复用         一个连接上并发多个请求（用 stream-id 区分）
//   Unix socket        vsock（VM ↔ 宿主机的专用 socket）
//
// 真实的 ttrpc 消息格式（简化）：
//
//   ┌──────────────────────────────────────────┐
//   │ Header（10 字节固定）                     │
//   │   length  (4 bytes) = payload 长度        │
//   │   stream-id (4 bytes) = 请求 ID           │
//   │   type    (1 byte)  = Request/Response     │
//   │   flags   (1 byte)                        │
//   ├──────────────────────────────────────────┤
//   │ Payload（protobuf 序列化的 Request/Response）│
//   └──────────────────────────────────────────┘
//
// 为什么用固定 4 字节 length header，而不用 \n 分隔？
//   - protobuf 是二进制，payload 里可能包含 \n 字节（会被误认为消息结束）
//   - 先读 4 字节知道后面还有多少字节，再读那么多字节，100% 准确
//   - 这种设计叫 "length-prefixed framing"，是二进制协议的标准做法

// ============================================================
// 对照点 3：消息边界问题（framing problem）
// ============================================================
//
// "协议"最核心的设计问题：接收方怎么知道"一条消息"从哪里到哪里？
//
// 三种常见解决方案：
//
// 方案 A：分隔符（我们的玩具协议 + QMP）
//   发：  {"cmd":"ping"}\n{"cmd":"info"}\n
//   收：  按 \n 分割，每段是一条消息
//   限制：消息内容里不能含有分隔符（或者需要转义）
//   适合：文本协议（JSON、HTTP headers、Redis RESP）
//
// 方案 B：长度前缀（ttrpc、protobuf over TCP）
//   发：  [0,0,0,4]["ping"][0,0,0,4]["info"]   (前4字节是后面内容的长度)
//   收：  先读4字节得到length，再读length字节，得到完整消息
//   限制：客户端必须先知道整条消息的长度（发送前要 buffer 整条消息）
//   适合：二进制协议
//
// 方案 C：固定长度
//   每条消息永远是 N 字节
//   限制：消息大小不灵活，浪费空间
//   适合：硬件协议、游戏同步包
//
// HTTP 是混合：headers 用 \r\n\r\n 分隔，body 用 Content-Length 指定长度。

// ============================================================
// 对照点 4：vsock — Kata 专用的 socket 类型
// ============================================================
//
// 我们的玩具协议和 QMP 用的是 Unix socket（仅限本机进程通信）。
// kata-shim 和 kata-agent 之间跨了一个 VM，不能用 Unix socket，
// 用的是 vsock（Virtual Socket）：
//
//   宿主机（kata-shim）
//       ↕ vsock（穿越 VM 边界，不需要网络配置）
//   VM 内部（kata-agent）
//
// vsock 在 Go 里的用法（和 Unix socket 几乎一样，只是 network 类型不同）：
//
//   // 普通 Unix socket（我们的玩具协议）
//   conn, err := net.Dial("unix", "/tmp/learn-protocol.sock")
//
//   // vsock（需要第三方库 github.com/mdlayher/vsock）
//   conn, err := vsock.Dial(vmCID, port, nil)
//   // vmCID = VM 的 Context ID（类似 IP 地址）
//   // port  = 端口（类似 TCP 端口）
//
// 协议本身（ttrpc/JSON）的代码完全不需要改，只是换了一层传输。
// 这就是"协议分层"：传输层和应用层解耦。

// ============================================================
// 总结：你现在能理解 containerd 调 kata 的完整链路了
// ============================================================
//
//   containerd
//       │ 调用 ttrpc（protobuf + 长度前缀 + vsock）
//       ↓
//   kata-shim（宿主机上的进程）
//       │ 调用 QMP（JSON + \n 分隔 + Unix socket）
//       ↓
//   QEMU（hypervisor）
//       │ 通过 vsock 转发命令
//       ↓
//   kata-agent（VM 内部）
//       │ 调用 OCI runtime（runc）
//       ↓
//   容器进程
//
// 每一层都是一个"协议"：
//   - 消息格式（JSON vs protobuf）不同
//   - 分隔方式（\n vs 长度前缀）不同
//   - 传输层（Unix socket vs vsock vs TCP）不同
//   但设计思想完全一样：约定好发什么、收什么、怎么分隔。
