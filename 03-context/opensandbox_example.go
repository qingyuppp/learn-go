package main

// ============================================================
// 第四部分：OpenSandbox 中的 context 用法
// ============================================================

// ---- 用法 1：HTTP 请求的 context ----
// 来源：components/egress/policy_server.go 第182行
//
// func (s *policyServer) handlePost(w http.ResponseWriter, r *http.Request) {
//     ...
//     s.commitPolicy(r.Context(), w, pol, "post")
//                    ^^^^^^^^^^
//     r.Context() 返回这个 HTTP 请求的 context
//     如果客户端断开连接，这个 context 会自动取消
//     commitPolicy 内部的所有操作都会收到取消信号
// }
//
// 这是 context 最常见的来源：从 HTTP 请求中获取

// ---- 用法 2：给危险操作加超时 ----
// 来源：同文件第283-284行
//
// func (s *policyServer) commitPolicy(ctx context.Context, ...) bool {
//     ...
//     nftCtx, nftCancel := context.WithTimeout(context.Background(), 30*time.Second)
//     defer nftCancel()
//     s.nft.ApplyStatic(nftCtx, ...)
// }
//
// 操作 nftables 防火墙可能卡住，所以加了 30 秒超时
// 注意用的是 context.Background() 而不是传入的 ctx
// 为什么？因为即使客户端断开，防火墙规则也必须应用完成
// 这是一个重要的设计决策：有些操作不应该跟随请求取消

// ---- 用法 3：可观测性中的 context ----
// 来源：components/execd/pkg/telemetry/record.go
//
// func RecordHTTPRequest(ctx context.Context, method, route string, ...) {
//     httpRequestDuration.Record(ctx, durationMillis, opt)
//                                ^^^
// }
//
// 每个 metric 记录都要传 ctx，这是 OpenTelemetry 的要求
// ctx 里携带了 trace ID、span 信息等元数据
// 这样 metric 数据可以和链路追踪关联起来
//
// 你后面做可观测性时，写的代码就是这个模式：
// telemetry.RecordHTTPRequest(r.Context(), r.Method, routeName, statusCode, duration)

// ---- context 传播链路 ----
//
// 一个完整的请求中，context 是这样流动的：
//
// 客户端发起请求
//   │
//   ▼
// http.Server 创建 context（绑定了连接生命周期）
//   │
//   ▼
// handler 通过 r.Context() 获取
//   │
//   ├──→ commitPolicy(ctx, ...)         传给业务逻辑
//   │       │
//   │       ├──→ nft.ApplyStatic(nftCtx, ...)  给危险操作单独加超时
//   │       └──→ proxy.UpdatePolicy(pol)
//   │
//   └──→ telemetry.RecordHTTPRequest(ctx, ...)  传给可观测性
//
// 如果客户端断开 → r.Context() 取消 → 所有下游操作收到取消信号
// 但 nftCtx 是独立的，不受影响（因为它基于 Background()）

// ---- 小结：context 的三条规则 ----
//
// 1. context 永远是函数的第一个参数，变量名叫 ctx
// 2. 不要把 context 存到结构体里，每次调用时传入
// 3. 需要超时时用 WithTimeout，需要手动取消时用 WithCancel
