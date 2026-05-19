package main

import (
	"context"
	"fmt"
	"time"
)

// ============================================================
// 第一部分：Go by Example — context 基础
// 来源：https://gobyexample-cn.github.io/context
// ============================================================

// simulateWork 模拟一个耗时操作（比如查数据库、调外部 API）
// 它接收 context 作为第一个参数（Go 的惯例）
func simulateWork(ctx context.Context, name string, duration time.Duration) error {
	fmt.Printf("[%s] 开始工作，需要 %v\n", name, duration)

	// select 同时等待两个事件，谁先发生就执行谁
	select {
	case <-time.After(duration):
		// 工作正常完成
		fmt.Printf("[%s] 完成!\n", name)
		return nil
	case <-ctx.Done():
		// context 被取消了（超时或主动取消）
		// ctx.Err() 返回取消的原因
		fmt.Printf("[%s] 被取消: %v\n", name, ctx.Err())
		return ctx.Err()
	}
}

// ============================================================
// 第二部分：三种 context 用法
// ============================================================

func main() {
	// --- 用法 1：WithTimeout — 超时自动取消 ---
	fmt.Println("=== 用法 1：超时取消 ===")
	// 创建一个 2 秒后自动取消的 context
	ctx1, cancel1 := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel1() // 无论如何都要调 cancel，释放资源

	// 工作需要 1 秒，超时是 2 秒 → 能完成
	simulateWork(ctx1, "快任务", 1*time.Second)
	// 工作需要 3 秒，超时是 2 秒 → 被取消
	simulateWork(ctx1, "慢任务", 3*time.Second)

	fmt.Println()

	// --- 用法 2：WithCancel — 手动取消 ---
	fmt.Println("=== 用法 2：手动取消 ===")
	ctx2, cancel2 := context.WithCancel(context.Background())

	// 在另一个 goroutine 里 500ms 后取消
	go func() {
		time.Sleep(500 * time.Millisecond)
		fmt.Println("[主控] 不想等了，取消!")
		cancel2()
	}()

	// 工作需要 2 秒，但 500ms 后就被取消了
	simulateWork(ctx2, "被打断的任务", 2*time.Second)

	fmt.Println()

	// --- 用法 3：context.Background() — 根 context ---
	fmt.Println("=== 用法 3：无超时 ===")
	// Background() 返回一个永远不会取消的 context
	// 适合 main 函数或顶层调用
	simulateWork(context.Background(), "无限制任务", 500*time.Millisecond)
}
