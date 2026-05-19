package main

// ============================================================
// 第四部分：OpenSandbox 中的结构体用法
// ============================================================

// ---- 用法 1：结构体作为"服务对象" ----
// 来源：components/egress/policy_server.go 第115-128行
//
// policyServer 结构体把所有相关的数据和依赖放在一起：
//
// type policyServer struct {
//     proxy           policyUpdater   // 接口类型的字段（上一课学的）
//     nft             nftApplier      // 接口类型的字段
//     server          *http.Server    // 指针字段
//     token           string          // 简单字段
//     enforcementMode string
//     maxEgressRules  int
//     mu              sync.Mutex      // 互斥锁（后面会学）
// }
//
// 它的方法全部用指针接收者，因为：
// 1. policyServer 很大，拷贝浪费性能
// 2. handlePost 需要修改内部状态（加锁、更新策略）
//
// func (s *policyServer) handlePolicy(w, r)  { ... }
// func (s *policyServer) handleGet(w)         { ... }
// func (s *policyServer) handlePost(w, r)     { ... }
// func (s *policyServer) authorize(r) bool    { ... }
//       ↑
//    指针接收者，所有方法共享同一个 policyServer 实例

// ---- 用法 2：结构体作为"数据传输对象" ----
// 来源：同文件第130-136行
//
// policyStatusResponse 纯粹用来序列化 JSON，没有方法：
//
// type policyStatusResponse struct {
//     Status          string `json:"status,omitempty"`
//     Mode            string `json:"mode,omitempty"`
//     EnforcementMode string `json:"enforcementMode,omitempty"`
// }
//
// 它和 learn-api 里的 ErrorResponse 是同一个模式

// ---- 用法 3：结构体作为"配置对象" ----
// 来源：components/internal/telemetry/init.go 第35-39行
//
// Config 结构体传递初始化参数：
//
// type Config struct {
//     ServiceName        string
//     ResourceAttributes []attribute.KeyValue
//     RegisterMetrics    func() error           // 字段可以是函数！
// }
//
// 调用方式：
// telemetry.Init(ctx, Config{
//     ServiceName:     "opensandbox-execd",
//     RegisterMetrics: registerExecdMetrics,    // 传入一个函数
// })
//
// 这是 Go 的常见模式："配置结构体" 代替一长串函数参数

// ---- 用法 4：包级变量 + 注册函数 ----
// 来源：components/execd/pkg/telemetry/init.go 第36-40行
//
// var (
//     httpRequestDuration      metric.Float64Histogram
//     executionDuration        metric.Float64Histogram
//     filesystemOperationDurMs metric.Float64Histogram
// )
//
// 这三个变量是 metric.Float64Histogram 接口类型（又用到了接口！）
// 它们在 registerExecdMetrics() 中被创建：
//
// meter := otel.Meter("opensandbox/execd")
// httpRequestDuration, err = meter.Float64Histogram(
//     "execd.http.request.duration",
//     metric.WithDescription("HTTP request duration"),
//     metric.WithUnit("ms"),
// )
//
// 创建后在其他地方调用 .Record() 记录指标值
// 这就是你后面要学的可观测性的核心代码！

// ---- 总结：OpenSandbox 中结构体的三种角色 ----
//
// 1. 服务对象：持有依赖 + 方法，用指针接收者（policyServer）
// 2. 数据传输：纯数据 + json tag，没有方法（policyStatusResponse）
// 3. 配置对象：传递初始化参数，代替长参数列表（Config）
