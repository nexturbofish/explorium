package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"hi/internal/lib"
	"os"
	"strings"

	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/schema"
)

type EditTool struct{ workspace string }

func (t *EditTool) Info(ctx context.Context) (*schema.ToolInfo, error) {
	return &schema.ToolInfo{
		Name: "edit",
		Desc: "Apply a search-and-replace edit to a file. Uses exact string matching",
		ParamsOneOf: schema.NewParamsOneOfByParams(map[string]*schema.ParameterInfo{
			"path":       {Type: "string", Desc: "File path relative to workspace", Required: true},
			"old_string": {Type: "string", Desc: "Text to replace (must match exactly)", Required: true},
			"new_string": {Type: "string", Desc: "Replacement text", Required: true},
		}),
	}, nil
}

func (t *EditTool) InvokableRun(ctx context.Context, argumentsInJSON string, opts ...tool.Option) (string, error) {
	var params struct {
		Path      string `json:"path"`
		OldString string `json:"old_string"`
		NewString string `json:"new_string"`
	}
	if err := json.Unmarshal([]byte(argumentsInJSON), &params); err != nil {
		return "", fmt.Errorf("edit: invalid arguments: %w", err)
	}
	if params.Path == "" || params.OldString == "" {
		return "", fmt.Errorf("edit: path and old_string are required")
	}
	safePath, err := lib.SafePath(t.workspace, params.Path)
	if err != nil {
		return "", fmt.Errorf("edit: %w", err)
	}
	data, err := os.ReadFile(safePath)
	if err != nil {
		return "", fmt.Errorf("edit: %w", err)
	}
	content := string(data)
	if !strings.Contains(content, params.OldString) {
		return "", fmt.Errorf("edit: old_string not found in %s", params.Path)
	}
	newContent := strings.ReplaceAll(content, params.OldString, params.NewString)
	if err := os.WriteFile(safePath, []byte(newContent), 0644); err != nil {
		return "", fmt.Errorf("edit: %w", err)
	}
	return fmt.Sprintf("applied edit to %s", params.Path), nil
}
