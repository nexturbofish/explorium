package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"hi/spec"
	"strings"
	"time"

	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/schema"
)

type MemorySearchTool struct{ store spec.MemoryStore }

func (t *MemorySearchTool) Info(ctx context.Context) (*schema.ToolInfo, error) {
	return &schema.ToolInfo{
		Name: "memory_search",
		Desc: "Search memories by semantic similarity. Returns up to 5 relevant memories.",
		ParamsOneOf: schema.NewParamsOneOfByParams(map[string]*schema.ParameterInfo{
			"query": {Type: "string", Desc: "Search query", Required: true},
		}),
	}, nil
}

func (t *MemorySearchTool) InvokableRun(ctx context.Context, argumentsInJSON string, opts ...tool.Option) (string, error) {
	if t.store == nil {
		return "", fmt.Errorf("memory_search: memory store not available")
	}
	var params struct{ Query string }
	if err := json.Unmarshal([]byte(argumentsInJSON), &params); err != nil {
		return "", fmt.Errorf("memory_search: invalid arguments: %w", err)
	}
	if params.Query == "" {
		return "", fmt.Errorf("memory_search: query is required")
	}
	results, err := t.store.Search(ctx, params.Query, 5)
	if err != nil {
		return "", fmt.Errorf("memory_search: %w", err)
	}
	if len(results) == 0 {
		return "(no relevant memories)", nil
	}
	var sb strings.Builder
	for _, r := range results {
		sb.WriteString(fmt.Sprintf("- [%.2f] %s\n", r.Score, r.Memory.Content))
	}
	return sb.String(), nil
}

type MemorySaveTool struct{ store spec.MemoryStore }

func (t *MemorySaveTool) Info(ctx context.Context) (*schema.ToolInfo, error) {
	return &schema.ToolInfo{
		Name: "memory_save",
		Desc: "Save a new memory. Returns the memory ID.",
		ParamsOneOf: schema.NewParamsOneOfByParams(map[string]*schema.ParameterInfo{
			"content": {Type: "string", Desc: "Memory content", Required: true},
			"zone":    {Type: "string", Desc: "Memory zone (e.g. 'user-preferences', 'project-rules')"},
			"tags":    {Type: "string", Desc: "Comma-separated tags"},
		}),
	}, nil
}

func (t *MemorySaveTool) InvokableRun(ctx context.Context, argumentsInJSON string, opts ...tool.Option) (string, error) {
	if t.store == nil {
		return "", fmt.Errorf("memory_save: memory store not available")
	}
	var params struct {
		Content string `json:"content"`
		Zone    string `json:"zone"`
		Tags    string `json:"tags"`
	}
	if err := json.Unmarshal([]byte(argumentsInJSON), &params); err != nil {
		return "", fmt.Errorf("memory_save: invalid arguments: %w", err)
	}
	if params.Content == "" {
		return "", fmt.Errorf("memory_save: content is required")
	}
	id := fmt.Sprintf("mem-%d", time.Now().UnixNano())
	scope := spec.MemoryScopeUser
	if params.Zone != "" {
		for _, prefix := range []string{"project-", "project/"} {
			if strings.HasPrefix(params.Zone, prefix) {
				scope = spec.MemoryScopeProject
				break
			}
		}
	}
	mem := &spec.LoadedMemory{
		Frontmatter: spec.MemoryFrontmatter{
			ID: id, CreatedAt: time.Now(),
			Scope:  scope,
			Source: spec.MemorySourceUser,
			Zone:   params.Zone,
			Tags:   splitTags(params.Tags),
		},
		Content: params.Content,
	}
	if err := t.store.Save(mem); err != nil {
		return "", fmt.Errorf("memory_save: %w", err)
	}
	return fmt.Sprintf("saved memory %s", id), nil
}

type MemoryDeleteTool struct{ store spec.MemoryStore }

func (t *MemoryDeleteTool) Info(ctx context.Context) (*schema.ToolInfo, error) {
	return &schema.ToolInfo{
		Name: "memory_delete",
		Desc: "Delete a memory by ID.",
		ParamsOneOf: schema.NewParamsOneOfByParams(map[string]*schema.ParameterInfo{
			"id": {Type: "string", Desc: "Memory ID to delete", Required: true},
		}),
	}, nil
}

func (t *MemoryDeleteTool) InvokableRun(ctx context.Context, argumentsInJSON string, opts ...tool.Option) (string, error) {
	if t.store == nil {
		return "", fmt.Errorf("memory_delete: memory store not available")
	}
	var params struct{ ID string }
	if err := json.Unmarshal([]byte(argumentsInJSON), &params); err != nil {
		return "", fmt.Errorf("memory_delete: invalid arguments: %w", err)
	}
	if params.ID == "" {
		return "", fmt.Errorf("memory_delete: id is required")
	}
	if err := t.store.Delete(params.ID); err != nil {
		return "", fmt.Errorf("memory_delete: %w", err)
	}
	return fmt.Sprintf("deleted memory %s", params.ID), nil
}

func splitTags(s string) []string {
	if s == "" {
		return nil
	}
	var out []string
	for _, t := range strings.Split(s, ",") {
		t = strings.TrimSpace(t)
		if t != "" {
			out = append(out, t)
		}
	}
	return out
}
