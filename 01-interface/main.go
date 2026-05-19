package main

import "fmt"

// ============================================================
// 第一部分：Go by Example 经典例子 — 几何图形
// 来源：https://gobyexample-cn.github.io/interfaces
// ============================================================

// geometry 是一个接口，定义了"几何图形"应该有哪些能力
// 任何类型只要实现了 area() 和 perim() 两个方法，就自动满足这个接口
// 不需要显式声明 "implements"，这叫做"隐式实现"
type geometry interface {
	area() float64
	perim() float64
}

// rect 是一个具体类型：矩形
type rect struct {
	width, height float64
}

// circle 是另一个具体类型：圆形
type circle struct {
	radius float64
}

// rect 实现了 geometry 接口的两个方法
func (r rect) area() float64 {
	return r.width * r.height
}
func (r rect) perim() float64 {
	return 2*r.width + 2*r.height
}

// circle 也实现了 geometry 接口的两个方法
func (c circle) area() float64 {
	return 3.14159 * c.radius * c.radius
}
func (c circle) perim() float64 {
	return 2 * 3.14159 * c.radius
}

// measure 接收 geometry 接口类型
// 不关心传进来的是 rect 还是 circle，只要有 area() 和 perim() 就行
func measure(g geometry) {
	fmt.Printf("类型: %T\n", g)
	fmt.Printf("面积: %.2f\n", g.area())
	fmt.Printf("周长: %.2f\n\n", g.perim())
}

// ============================================================
// 第二部分：接口的核心价值 — 一个函数处理多种类型
// ============================================================

func main() {
	r := rect{width: 3, height: 4}
	c := circle{radius: 5}

	// 同一个函数 measure，既能处理矩形，也能处理圆形
	// 这就是接口的价值：定义"能力"而不是"类型"
	measure(r)
	measure(c)
}
