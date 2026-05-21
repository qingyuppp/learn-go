// Package todo 你的练习：实现 Todo API 的 Go SDK 客户端。
//
// 建议工作流：
//  1. 先看 reference/client.go，理解每一行
//  2. 关掉 reference，在这里凭记忆默写
//  3. 跑 go test -v 验证
package todo

// TODO: 你需要实现：
//
// 1. Todo 结构体
//    type Todo struct {
//        ID    int    `json:"id"`
//        Title string `json:"title"`
//        Done  bool   `json:"done"`
//    }
//
// 2. APIError 类型（实现 error 接口）
//    type APIError struct {
//        StatusCode int
//        Body       string
//    }
//    func (e *APIError) Error() string { ... }
//
// 3. Client 结构体
//    type Client struct {
//        baseURL    string
//        apiKey     string
//        httpClient *http.Client
//    }
//
// 4. NewClient(baseURL, apiKey string) *Client
//
// 5. doRequest 内部方法（核心，封装通用逻辑）
//    func (c *Client) doRequest(ctx context.Context, method, path string,
//        reqBody, respOut interface{}) error { ... }
//
// 6. 四个公开方法：
//    - ListTodos(ctx) ([]Todo, error)
//    - CreateTodo(ctx, title string) (*Todo, error)
//    - GetTodo(ctx, id int) (*Todo, error)
//    - DeleteTodo(ctx, id int) error
//
// 需要 import：bytes, context, encoding/json, fmt, io, net/http, time
