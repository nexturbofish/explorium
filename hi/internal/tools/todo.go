package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/schema"
)

// Todo 表示一个待办项
type Todo struct {
	ID        string    `json:"id"`
	Content   string    `json:"content"`
	Status    string    `json:"status"` // pending, done, cancelled
	CreatedAt time.Time `json:"created_at"`
}

// TodoStore 管理待办项，持久化到 JSON 文件
type TodoStore struct {
	path  string
	mu    sync.RWMutex
	items []Todo
}

func NewTodoStore(path string) *TodoStore {
	s := &TodoStore{path: path}
	s.load() // 启动时加载已有数据
	return s
}

func (s *TodoStore) load() {
	data, err := os.ReadFile(s.path)
	if err != nil {
		return // 首次使用文件不存在，静默忽略
	}
	json.Unmarshal(data, &s.items)
}

func (s *TodoStore) save() error {
	data, err := json.MarshalIndent(s.items, "", "  ")
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(s.path), 0755); err != nil {
		return err
	}
	return os.WriteFile(s.path, data, 0644)
}

// TodoWriteTool 创建或更新待办项
type TodoWriteTool struct{ store *TodoStore }

func (t *TodoWriteTool) Info(ctx context.Context) (*schema.ToolInfo, error) {
	return &schema.ToolInfo{
		Name: "todo_write",
		Desc: "Create or update a todo item. Use id=empty to create, or provide an id to update.",
		ParamsOneOf: schema.NewParamsOneOfByParams(map[string]*schema.ParameterInfo{
			"id":      {Type: "string", Desc: "Todo ID (leave empty to create new)"},
			"content": {Type: "string", Desc: "Todo content", Required: true},
			"status":  {Type: "string", Desc: "Status: pending, done, or cancelled"},
		}),
	}, nil
}

func (t *TodoWriteTool) InvokableRun(ctx context.Context, argumentsInJSON string, opts ...tool.Option) (string, error) {
	var params struct {
		ID      string `json:"id"`
		Content string `json:"content"`
		Status  string `json:"status"`
	}
	if err := json.Unmarshal([]byte(argumentsInJSON), &params); err != nil {
		return "", fmt.Errorf("todo_write: invalid arguments: %w", err)
	}
	if params.Content == "" {
		return "", fmt.Errorf("todo_write: content is required")
	}
	if params.Status == "" {
		params.Status = "pending"
	}
	t.store.mu.Lock()
	defer t.store.mu.Unlock()
	if params.ID == "" {
		params.ID = fmt.Sprintf("todo-%d", time.Now().UnixNano())
		t.store.items = append(t.store.items, Todo{
			ID: params.ID, Content: params.Content,
			Status: params.Status, CreatedAt: time.Now(),
		})
	} else {
		found := false
		for i := range t.store.items {
			if t.store.items[i].ID == params.ID {
				if params.Content != "" {
					t.store.items[i].Content = params.Content
				}
				t.store.items[i].Status = params.Status
				found = true
				break
			}
		}
		if !found {
			return "", fmt.Errorf("todo_write: todo %q not found", params.ID)
		}
	}
	if err := t.store.save(); err != nil {
		return "", fmt.Errorf("todo_write: %w", err)
	}
	return fmt.Sprintf("todo %s: %s", params.ID, params.Status), nil
}

// TodoListTool 列出待办项
type TodoListTool struct{ store *TodoStore }

func (t *TodoListTool) Info(ctx context.Context) (*schema.ToolInfo, error) {
	return &schema.ToolInfo{
		Name: "todo_list",
		Desc: "List all todo items, optionally filtered by status.",
		ParamsOneOf: schema.NewParamsOneOfByParams(map[string]*schema.ParameterInfo{
			"status": {Type: "string", Desc: "Filter by status: pending, done, cancelled (empty = all)"},
		}),
	}, nil
}

func (t *TodoListTool) InvokableRun(ctx context.Context, argumentsInJSON string, opts ...tool.Option) (string, error) {
	var params struct{ Status string }
	if err := json.Unmarshal([]byte(argumentsInJSON), &params); err != nil {
		return "", fmt.Errorf("todo_list: invalid arguments: %w", err)
	}
	t.store.mu.RLock()
	defer t.store.mu.RUnlock()
	var sb strings.Builder
	count := 0
	for _, item := range t.store.items {
		if params.Status != "" && item.Status != params.Status {
			continue
		}
		sb.WriteString(fmt.Sprintf("- [%s] %s (id=%s)\n", item.Status, item.Content, item.ID))
		count++
	}
	if count == 0 {
		return "(no todos)", nil
	}
	return sb.String(), nil
}
