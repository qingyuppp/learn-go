package main

// ============================================================
// 第三部分：OpenSandbox 中的接口用法
// 来源：components/egress/policy_server.go 第38-49行
// ============================================================

// 在 OpenSandbox 的 egress 组件中，policyServer 需要做两件事：
// 1. 更新 DNS 代理的策略（policyUpdater）
// 2. 应用 nftables 防火墙规则（nftApplier）
//
// 但 policyServer 不关心这两件事的具体实现方式，
// 它只关心"你能不能做这件事"。所以用接口定义能力：

// ---- 接口定义（OpenSandbox 原始代码简化版）----

// policyUpdater 定义了"能更新策略"的能力
// 任何类型只要有这三个方法，就可以作为 policyUpdater 使用
//
// type policyUpdater interface {
//     CurrentPolicy() *policy.NetworkPolicy    // 获取当前策略
//     UpdatePolicy(*policy.NetworkPolicy)       // 更新策略
//     UpdateAlwaysRules(deny, allow []EgressRule) // 更新永久规则
// }

// nftApplier 定义了"能操作防火墙"的能力
//
// type nftApplier interface {
//     ApplyStatic(ctx, policy) error       // 应用静态规则
//     AddResolvedIPs(ctx, ips) error       // 添加 DNS 解析到的 IP
//     RemoveEnforcement(ctx) error         // 移除所有规则
// }

// ---- 为什么用接口？----
//
// policyServer 结构体持有的是接口，不是具体类型：
//
// type policyServer struct {
//     proxy policyUpdater    // ← 接口，不是具体的 Proxy 结构体
//     nft   nftApplier       // ← 接口，不是具体的 NftManager 结构体
// }
//
// 好处：
// 1. 解耦 — policyServer 不依赖 Proxy 和 NftManager 的具体实现
// 2. 可测试 — 测试时可以传一个 mock（假的实现），不需要真的操作防火墙
// 3. 可替换 — 换一种防火墙实现（比如 iptables），不需要改 policyServer 的代码

// ---- 对比第一部分 ----
//
// geometry 接口例子：            OpenSandbox 接口例子：
// ┌──────────────────┐          ┌─────────────────────┐
// │ geometry 接口     │          │ policyUpdater 接口    │
// │  area()          │          │  CurrentPolicy()     │
// │  perim()         │          │  UpdatePolicy()      │
// └──────┬───────────┘          └──────┬──────────────┘
//        │                             │
//   ┌────┴────┐                   ┌────┴────┐
//   │  rect   │  circle           │  Proxy  │  MockProxy（测试用）
//   └─────────┘                   └─────────┘
//
// 模式完全一样：接口定义能力，多个类型可以实现同一个接口
