// Package todo 是 Todo API 的 Go SDK 客户端（参考实现）。
//
// 这是"参考实现"，详细注释每一行设计决策。
// 阅读顺序：Todo → APIError → Client → NewClient → doRequest → 4 个公开方法
package todo

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// Todo 是服务端返回的 todo 数据结构。
// json tag 告诉编解码器：Go 字段名 ID → JSON 字段名 id（驼峰转小写）。
type Todo struct {
	ID    int    `json:"id"`
	Title string `json:"title"`
	Done  bool   `json:"done"`
}

// APIError 是 HTTP 调用失败的自定义错误。
// 实现 error 接口（需要有 Error() string 方法）即可被 errors.Is / errors.As 识别。
type APIError struct {
	StatusCode int    // HTTP 状态码（4xx/5xx）
	Body       string // 服务端返回的错误体（便于排查）
}

func (e *APIError) Error() string {
	return fmt.Sprintf("API error: status=%d body=%s", e.StatusCode, e.Body)
}

// Client 是 SDK 的核心对象。
//
// 三个字段的作用：
//   - baseURL：服务地址（如 "https://api.example.com"）
//   - apiKey：认证凭证，自动加到每个请求的 Header
//   - httpClient：底层 HTTP 客户端，可以配超时、连接池等
//
// 为什么是 *http.Client 而不是新建？因为 http.Client 是线程安全的，应该复用。
type Client struct {
	baseURL    string
	apiKey     string
	httpClient *http.Client
}

// NewClient 构造一个 Client。
//
// 工厂函数好处：
//   - 强制用户提供必要参数（baseURL、apiKey）
//   - 给可选参数提供合理默认值（这里默认 10 秒超时）
//   - 隔离构造逻辑，未来加配置不破坏 API
func NewClient(baseURL, apiKey string) *Client {
	return &Client{
		baseURL: baseURL,
		apiKey:  apiKey,
		httpClient: &http.Client{
			Timeout: 10 * time.Second, // 防止请求挂死
		},
	}
}

// doRequest 是内部方法，封装 4 个公开方法共用的逻辑：
//   - 拼 URL
//   - 序列化请求体
//   - 注入 Auth Header
//   - 发送请求
//   - 处理状态码错误
//   - 反序列化响应体
//
// 把重复逻辑抽出来，公开方法只关心"业务"（路径、方法、参数）。
// 这是 SDK 设计的核心模式：thin public methods + thick internal helper。
func (c *Client) doRequest(ctx context.Context, method, path string, reqBody, respOut interface{}) error {
	// 1. 序列化请求体（如果有）
	var body io.Reader
	if reqBody != nil {
		buf := &bytes.Buffer{}
		if err := json.NewEncoder(buf).Encode(reqBody); err != nil {
			return fmt.Errorf("encode request: %w", err)
		}
		body = buf
	}

	// 2. 构造请求（带 context，支持上层取消/超时）
	url := c.baseURL + path
	req, err := http.NewRequestWithContext(ctx, method, url, body)
	if err != nil {
		return fmt.Errorf("new request: %w", err)
	}

	// 3. 注入通用 Header
	req.Header.Set("Authorization", "Bearer "+c.apiKey)
	if reqBody != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	// 4. 发送请求
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("do request: %w", err)
	}
	defer resp.Body.Close() // 必须关闭，否则连接泄漏

	// 5. 处理错误状态码
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		respBody, _ := io.ReadAll(resp.Body) // 读全错误体，便于排查
		return &APIError{
			StatusCode: resp.StatusCode,
			Body:       string(respBody),
		}
	}

	// 6. 反序列化响应体（如果调用方需要）
	if respOut != nil {
		if err := json.NewDecoder(resp.Body).Decode(respOut); err != nil {
			return fmt.Errorf("decode response: %w", err)
		}
	}

	return nil
}

// ListTodos 列出所有 todo。
func (c *Client) ListTodos(ctx context.Context) ([]Todo, error) {
	var todos []Todo
	if err := c.doRequest(ctx, "GET", "/todos", nil, &todos); err != nil {
		return nil, err
	}
	return todos, nil
}

// CreateTodo 创建一个新 todo。
// 请求体只包含 title，服务端会分配 ID 并返回完整 Todo。
func (c *Client) CreateTodo(ctx context.Context, title string) (*Todo, error) {
	reqBody := map[string]string{"title": title}
	var todo Todo
	if err := c.doRequest(ctx, "POST", "/todos", reqBody, &todo); err != nil {
		return nil, err
	}
	return &todo, nil
}

// GetTodo 获取指定 ID 的 todo。
func (c *Client) GetTodo(ctx context.Context, id int) (*Todo, error) {
	var todo Todo
	path := fmt.Sprintf("/todos/%d", id)
	if err := c.doRequest(ctx, "GET", path, nil, &todo); err != nil {
		return nil, err
	}
	return &todo, nil
}

// DeleteTodo 删除指定 ID 的 todo。
// 不需要返回值，所以 respOut 传 nil。
func (c *Client) DeleteTodo(ctx context.Context, id int) error {
	path := fmt.Sprintf("/todos/%d", id)
	return c.doRequest(ctx, "DELETE", path, nil, nil)
}
