package todo

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// 用 httptest.NewServer 启动一个真实的本地 HTTP 服务器（端口随机），
// 比 mock 库更接近真实场景。

func TestListTodos(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/todos" || r.Method != "GET" {
			t.Errorf("unexpected request: %s %s", r.Method, r.URL.Path)
		}
		json.NewEncoder(w).Encode([]Todo{
			{ID: 1, Title: "a", Done: false},
			{ID: 2, Title: "b", Done: true},
		})
	}))
	defer server.Close()

	client := NewClient(server.URL, "key")
	todos, err := client.ListTodos(context.Background())
	if err != nil {
		t.Fatalf("ListTodos: %v", err)
	}
	if len(todos) != 2 {
		t.Fatalf("expected 2 todos, got %d", len(todos))
	}
	if todos[0].Title != "a" {
		t.Fatalf("expected first todo title=a, got %s", todos[0].Title)
	}
}

func TestCreateTodo(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/todos" || r.Method != "POST" {
			t.Errorf("unexpected request: %s %s", r.Method, r.URL.Path)
		}
		var body map[string]string
		json.NewDecoder(r.Body).Decode(&body)
		if body["title"] != "buy milk" {
			t.Errorf("expected title=buy milk, got %s", body["title"])
		}
		json.NewEncoder(w).Encode(Todo{ID: 42, Title: body["title"], Done: false})
	}))
	defer server.Close()

	client := NewClient(server.URL, "key")
	todo, err := client.CreateTodo(context.Background(), "buy milk")
	if err != nil {
		t.Fatalf("CreateTodo: %v", err)
	}
	if todo.ID != 42 || todo.Title != "buy milk" {
		t.Fatalf("unexpected todo: %+v", todo)
	}
}

func TestGetTodo(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/todos/42" || r.Method != "GET" {
			t.Errorf("unexpected request: %s %s", r.Method, r.URL.Path)
		}
		json.NewEncoder(w).Encode(Todo{ID: 42, Title: "hello", Done: false})
	}))
	defer server.Close()

	client := NewClient(server.URL, "key")
	todo, err := client.GetTodo(context.Background(), 42)
	if err != nil {
		t.Fatalf("GetTodo: %v", err)
	}
	if todo.Title != "hello" {
		t.Fatalf("expected title=hello, got %s", todo.Title)
	}
}

func TestDeleteTodo(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/todos/42" || r.Method != "DELETE" {
			t.Errorf("unexpected request: %s %s", r.Method, r.URL.Path)
		}
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	client := NewClient(server.URL, "key")
	if err := client.DeleteTodo(context.Background(), 42); err != nil {
		t.Fatalf("DeleteTodo: %v", err)
	}
}

func TestAuthHeader(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		auth := r.Header.Get("Authorization")
		if auth != "Bearer secret-key" {
			t.Errorf("expected Authorization=Bearer secret-key, got %q", auth)
		}
		json.NewEncoder(w).Encode([]Todo{})
	}))
	defer server.Close()

	client := NewClient(server.URL, "secret-key")
	if _, err := client.ListTodos(context.Background()); err != nil {
		t.Fatalf("ListTodos: %v", err)
	}
}

func TestErrorHandling(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte("todo not found"))
	}))
	defer server.Close()

	client := NewClient(server.URL, "key")
	_, err := client.GetTodo(context.Background(), 999)
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	var apiErr *APIError
	if !errors.As(err, &apiErr) {
		t.Fatalf("expected *APIError, got %T: %v", err, err)
	}
	if apiErr.StatusCode != 404 {
		t.Fatalf("expected status 404, got %d", apiErr.StatusCode)
	}
	if !strings.Contains(apiErr.Body, "not found") {
		t.Fatalf("expected body to contain 'not found', got %q", apiErr.Body)
	}

	// 验证 Error() 方法实现
	if msg := err.Error(); !strings.Contains(msg, "404") {
		t.Fatalf("expected Error() to contain 404, got %q", msg)
	}
}

func ExampleClient() {
	// 这是 example test，会出现在 godoc 里
	// 不实际运行（因为没有真实服务器）

	// client := NewClient("https://api.example.com", "my-api-key")
	// todos, _ := client.ListTodos(context.Background())
	// fmt.Println(len(todos))

	fmt.Println("see ListTodos / CreateTodo / GetTodo / DeleteTodo")
	// Output: see ListTodos / CreateTodo / GetTodo / DeleteTodo
}
