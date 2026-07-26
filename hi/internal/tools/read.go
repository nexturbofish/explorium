package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"hi/internal/lib"
	"os"

	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/schema"
)

type ReadTool struct{ workspace string }

func (t *ReadTool) Info(ctx context.Context) (*schema.ToolInfo, error) {
	return &schema.ToolInfo{
		Name: "read",
		Desc: "Read the contents of a file",
		ParamsOneOf: schema.NewParamsOneOfByParams(map[string]*schema.ParameterInfo{
			"path": {Type: "string", Desc: "File path relative to workspace"},
		}),
	}, nil
}

func (t *ReadTool) InvokableRun(ctx context.Context, argumentsInJSON string, opts ...tool.Option) (string, error) {
	var params struct{ Path string }
	if err := json.Unmarshal([]byte(argumentsInJSON), &params); err != nil {
		return "", fmt.Errorf("read: invalid arguments: %w", err)
	}
	if params.Path == "" {
		return "", fmt.Errorf("read: path is required")
	}
	// 安全检查，防止路径穿越
	safePath, err := lib.SafePath(t.workspace, params.Path)
	if err != nil {
		return "", fmt.Errorf("read: %w", err)
	}
	data, err := os.ReadFile(safePath)
	if err != nil {
		return "", fmt.Errorf("read: %w", err)
	}
	return string(data), nil
}
