package main

// ============================================================
// 第六部分：OpenSandbox 中的并发用法
// ============================================================

// ---- 用法 1：goroutine + channel + select 启动服务器 ----
// 来源：components/egress/policy_server.go 第94-112行
//
// errCh := make(chan error, 1)                      // 创建带缓冲的 channel
// safego.Go(func() {                                // 在 goroutine 里启动服务器
//     if err := srv.ListenAndServe(); err != nil {
//         errCh <- err                               // 出错就把 error 塞进 channel
//     }
// })
//
// select {
// case err := <-errCh:                 // 如果 200ms 内收到错误
//     return nil, err                  // → 启动失败
// case <-time.After(200 * time.Millisecond):  // 如果 200ms 内没出错
//     // → 认为启动成功                 // 继续后续初始化
// }
//
// 这和我们的 demoSelect 一模一样：同时等多个事件，谁先发生处理谁
// errCh 容量为 1（make(chan error, 1)），这样发送方不会阻塞

// ---- 用法 2：Mutex 保护并发写入 ----
// 来源：同文件第124、169、221行
//
// type policyServer struct {
//     mu sync.Mutex    // 互斥锁
//     ...
// }
//
// func (s *policyServer) handlePost(w, r) {
//     s.mu.Lock()         // 加锁
//     defer s.mu.Unlock() // 函数结束时自动解锁（无论正常返回还是 panic）
//     // ... 修改策略 ...
// }
//
// 为什么需要锁？
// 如果两个客户端同时 POST 更新策略，不加锁可能导致：
// 1. 请求 A 读到策略 v1
// 2. 请求 B 读到策略 v1
// 3. 请求 A 写入策略 v2
// 4. 请求 B 写入策略 v3（覆盖了 A 的修改）
// 加锁后，B 必须等 A 完成才能开始
//
// 和我们 demoMutex 的 counter++ 一样的道理
// learn-api 里也用了：todos map 的读写如果加上并发就需要锁

// ---- 用法 3：后台 goroutine 定时执行 ----
// 来源：同文件第296-299行
//
// func (s *policyServer) startAlwaysRuleReloadJob() {
//     safego.Go(func() {
//         wait.Until(s.reloadAlwaysRulesJob, time.Minute, s.stopAlwaysReload)
//     })
// }
//
// wait.Until 是 K8s 的工具函数：每隔 time.Minute 调一次 reloadAlwaysRulesJob
// 直到 s.stopAlwaysReload channel 被关闭
// 这就是"后台定时任务"模式

// ---- 用法 4：可观测性中的 goroutine ----
// 来源：components/execd/pkg/telemetry/init.go 第86-93行
//
// meter.Int64ObservableGauge(
//     "execd.system.process.count",
//     metric.WithInt64Callback(func(ctx context.Context, obs metric.Int64Observer) error {
//         obs.Observe(systemProcessCount())  // 每次被采集时调用 systemProcessCount()
//         return nil
//     }),
// )
//
// ObservableGauge 的 callback 由 OpenTelemetry 在后台 goroutine 里定时调用
// 每次调用时读取当前进程数、CPU、内存等系统指标
// 这些采集函数在 system.go 里：
//   - systemProcessCount()     → 进程数
//   - systemCPUUsagePercent()  → CPU 使用率
//   - systemMemoryUsageBytes() → 内存使用量
//
// 你后面做可观测性时会直接用到这个模式

// ---- 总结：OpenSandbox 的四种并发模式 ----
//
// 1. goroutine + channel + select → 异步启动 + 快速失败检测
// 2. sync.Mutex + defer Unlock   → 保护共享数据的并发安全
// 3. 后台定时 goroutine          → 周期性任务（规则重载）
// 4. callback 回调               → 可观测性指标的定时采集
