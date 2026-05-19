package main

import (
	"fmt"
	"sync"
	"time"
)

// ============================================================
// 第一部分：Go by Example — goroutine 基础
// 来源：https://gobyexample-cn.github.io/goroutines
// ============================================================

func worker(id int) {
	fmt.Printf("[worker %d] 开始\n", id)
	time.Sleep(500 * time.Millisecond) // 模拟耗时工作
	fmt.Printf("[worker %d] 完成\n", id)
}

// ============================================================
// 第二部分：Go by Example — WaitGroup 等待多个 goroutine 完成
// 来源：https://gobyexample-cn.github.io/waitgroups
// ============================================================

func demoWaitGroup() {
	fmt.Println("=== WaitGroup：等待所有 goroutine 完成 ===")
	var wg sync.WaitGroup

	for i := 1; i <= 3; i++ {
		wg.Add(1) // 计数器 +1，表示多了一个任务

		// 注意：把 i 作为参数传入，不要直接在闭包里用 i
		// 否则所有 goroutine 可能看到同一个 i 值（闭包陷阱）
		go func(id int) {
			defer wg.Done() // 任务完成时计数器 -1
			worker(id)
		}(i)
	}

	wg.Wait() // 阻塞，直到计数器变成 0（所有任务都完成）
	fmt.Println("所有 worker 完成!\n")
}

// ============================================================
// 第三部分：Go by Example — Channel 在 goroutine 之间通信
// 来源：https://gobyexample-cn.github.io/channels
// ============================================================

func demoChannel() {
	fmt.Println("=== Channel：goroutine 之间传数据 ===")

	// make(chan string) 创建一个字符串类型的 channel
	// channel 就像一个管道：一端塞数据，另一端取数据
	messages := make(chan string)

	// 在 goroutine 里发送消息
	go func() {
		time.Sleep(300 * time.Millisecond)
		messages <- "来自 goroutine 的消息" // 发送（塞进管道）
	}()

	msg := <-messages // 接收（从管道取出），会阻塞直到有数据
	fmt.Println("收到:", msg)
	fmt.Println()
}

// ============================================================
// 第四部分：Go by Example — Mutex 互斥锁
// 来源：https://gobyexample-cn.github.io/mutexes
// ============================================================

func demoMutex() {
	fmt.Println("=== Mutex：多个 goroutine 安全地修改同一个变量 ===")

	var (
		counter int
		mu      sync.Mutex
	)

	var wg sync.WaitGroup
	// 启动 100 个 goroutine，每个都给 counter +1
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			mu.Lock()   // 加锁：同一时刻只有一个 goroutine 能执行下面的代码
			counter++
			mu.Unlock() // 解锁：让其他 goroutine 进来
		}()
	}

	wg.Wait()
	// 如果不加锁，counter 的结果可能小于 100（数据竞争）
	// 加了锁，结果一定是 100
	fmt.Printf("counter = %d（应该是 100）\n\n", counter)
}

// ============================================================
// 第五部分：select — 同时等待多个 channel
// ============================================================

func demoSelect() {
	fmt.Println("=== Select：同时等多个事件，谁先来处理谁 ===")

	ch1 := make(chan string)
	ch2 := make(chan string)

	go func() {
		time.Sleep(100 * time.Millisecond)
		ch1 <- "结果A"
	}()

	go func() {
		time.Sleep(200 * time.Millisecond)
		ch2 <- "结果B"
	}()

	// select 同时等 ch1 和 ch2，谁先有数据就处理谁
	// 上一课 context 里的 select 也是这个模式
	select {
	case msg := <-ch1:
		fmt.Println("先收到 ch1:", msg)
	case msg := <-ch2:
		fmt.Println("先收到 ch2:", msg)
	}
	fmt.Println()
}

func main() {
	demoWaitGroup()
	demoChannel()
	demoMutex()
	demoSelect()
}
