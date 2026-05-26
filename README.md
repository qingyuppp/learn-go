# learn-go

Go 语言核心特性学习，每个概念一个可运行的例子。重点学习后端开发中高频使用的特性，而非语法大全。

## 学习路径

1. **error handling** — 错误处理惯用法、自定义错误、errors.Is/As、error wrapping
2. **interface** — 接口定义与实现、隐式实现、常用标准库接口（io.Reader、fmt.Stringer）
3. **goroutine & channel** — 并发基础、channel 通信、select 多路复用
4. **sync 包** — Mutex、WaitGroup、Once、Map，并发安全模式
5. **context** — 超时控制、取消传播、值传递，贯穿整个请求生命周期
6. **struct & method** — 值接收者 vs 指针接收者、组合（embedding）、构造函数模式
7. **slice & map 深入** — 底层原理、容量扩容、常见陷阱
8. **testing** — 表驱动测试、子测试、benchmark、testify
9. **io 与文件操作** — Reader/Writer 接口、文件读写、bufio
10. **标准库实战** — net/http、encoding/json、os/exec、flag

## 学习路径（目录）

| 目录 | 主题 | 状态 |
|---|---|---|
| [01-interface](./01-interface/) | 接口定义与实现、隐式实现 | ✅ |
| [02-struct](./02-struct/) | struct、值/指针接收者 | ✅ |
| [03-context](./03-context/) | 超时控制、取消传播 | ✅ |
| [04-goroutine](./04-goroutine/) | goroutine、channel、Mutex、select | ✅ |
| [05-protocol-and-net](./05-protocol-and-net/) | 协议是什么、Unix socket、类 QMP 实现 | ✅ |
| [daily-exercises](./daily-exercises/) | 每日实战小题 | 🔧 进行中 |

## 关联

本项目是 [后端开发技能路线图](https://github.com/qingyuppp/learn-api) 的第 2 项技能，是后续所有技能的基础。
