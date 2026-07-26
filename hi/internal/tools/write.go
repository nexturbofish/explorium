package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"hi/internal/lib"
	"os"
	"path/filepath"

	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/schema"
)

type WriteTool struct{ workspace string }

func (t *WriteTool) Info(ctx context.Context) (*schema.ToolInfo, error) {
	return &schema.ToolInfo{
		Name: "write",
		Desc: "Write content to a file. Creates parent directories if needed.",
		ParamsOneOf: schema.NewParamsOneOfByParams(map[string]*schema.ParameterInfo{
			"path":    {Type: "string", Desc: "File path relative to workspace", Required: true},
			"content": {Type: "string", Desc: "Content to write", Required: true},
		}),
	}, nil
}

func (t *WriteTool) InvokableRun(ctx context.Context, argumentsInJSON string, opts ...tool.Option) (string, error) {
	var params struct {
		Path    string `json:"path"`
		Content string `json:"content"`
	}
	if err := json.Unmarshal([]byte(argumentsInJSON), &params); err != nil {
		return "", fmt.Errorf("write: invalid arguments: %w", err)
	}
	if params.Path == "" {
		return "", fmt.Errorf("write: path is required")
	}
	safePath, err := lib.SafePath(t.workspace, params.Path)
	if err != nil {
		return "", fmt.Errorf("write: %w", err)
	}
	if err := os.MkdirAll(filepath.Dir(safePath), 0755); err != nil {
		return "", fmt.Errorf("write: mkdir: %w", err)
	}
	if err := os.WriteFile(safePath, []byte(params.Content), 0644); err != nil {
		return "", fmt.Errorf("write: %w", err)
	}
	return fmt.Sprintf("wrote %d bytes to %s", len(params.Content), params.Path), nil
}
