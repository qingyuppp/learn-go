package main

import "fmt"

// ============================================================
// 第一部分：Go by Example — 结构体基础
// 来源：https://gobyexample-cn.github.io/structs
// ============================================================

// person 定义了一个结构体，包含两个字段
type person struct {
	name string
	age  int
}

// newPerson 是构造函数（Go 没有 constructor 关键字，用普通函数代替）
// 返回指针是惯例，避免拷贝大结构体
func newPerson(name string, age int) *person {
	return &person{name: name, age: age}
}

// ============================================================
// 第二部分：Go by Example — 方法（Methods）
// 来源：https://gobyexample-cn.github.io/methods
// ============================================================

// 方法就是绑定在结构体上的函数
// (p person) 叫"接收者"，表示这个方法属于 person 类型

// greet 使用值接收者 — 拿到的是 person 的副本
// 在方法内修改 p 不会影响原始数据
func (p person) greet() string {
	return fmt.Sprintf("你好，我是%s，今年%d岁", p.name, p.age)
}

// birthday 使用指针接收者 — 拿到的是 person 的指针
// 在方法内修改 p 会直接修改原始数据
func (p *person) birthday() {
	p.age++ // 直接修改原始的 age
}

// ============================================================
// 第三部分：值接收者 vs 指针接收者
// ============================================================

func main() {
	// 创建结构体的几种方式
	p1 := person{name: "张三", age: 25}       // 值
	p2 := &person{name: "李四", age: 30}       // 指针（&取地址）
	p3 := newPerson("王五", 28)                // 用构造函数

	fmt.Println(p1.greet())
	fmt.Println(p2.greet())
	fmt.Println(p3.greet())

	// 值接收者 vs 指针接收者的区别
	fmt.Printf("\n--- 值接收者 vs 指针接收者 ---\n")
	fmt.Printf("生日前: %s, age=%d\n", p1.name, p1.age)
	p1.birthday() // birthday 是指针接收者，Go 自动取地址
	fmt.Printf("生日后: %s, age=%d\n", p1.name, p1.age)
	// age 变成了 26，说明指针接收者修改了原始数据

	fmt.Printf("\n--- 何时用值，何时用指针？---\n")
	fmt.Println("值接收者：只读取数据，不修改（如 greet）")
	fmt.Println("指针接收者：需要修改数据，或结构体很大避免拷贝（如 birthday）")
	fmt.Println("经验法则：如果拿不准，用指针接收者")
}
