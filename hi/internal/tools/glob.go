package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/schema"
)

type GlobTool struct{ workspace string }

func (t *GlobTool) Info(ctx context.Context) (*schema.ToolInfo, error) {
	return &schema.ToolInfo{
		Name: "glob",
		Desc: "List files matching a pattern (recursive). Uses gitignore-style rules.",
		ParamsOneOf: schema.NewParamsOneOfByParams(map[string]*schema.ParameterInfo{
			"pattern": {Type: "string", Desc: "Glob pattern (e.g. **/*.go)", Required: true},
		}),
	}, nil
}

func (t *GlobTool) InvokableRun(ctx context.Context, argumentsInJSON string, opts ...tool.Option) (string, error) {
	var params struct{ Pattern string }
	if err := json.Unmarshal([]byte(argumentsInJSON), &params); err != nil {
		return "", fmt.Errorf("glob: invalid arguments: %w", err)
	}
	if params.Pattern == "" {
		return "", fmt.Errorf("glob: pattern is required")
	}
	pattern := filepath.Join(t.workspace, params.Pattern)
	// filepath.WalkDir 支持 ** 递归匹配
	var sb strings.Builder
	matched := false
	walkFn := func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		ok, err := filepath.Match(pattern, path)
		if err != nil {
			return err
		}
		if !ok {
			// 尝试只用 pattern 的基本名部分匹配
			ok, _ = filepath.Match(filepath.Base(pattern), d.Name())
		}
		if ok {
			rel, err := filepath.Rel(t.workspace, path)
			if err != nil {
				return err
			}
			sb.WriteString(rel + "\n")
			matched = true
		}
		return nil
	}
	if err := filepath.WalkDir(t.workspace, walkFn); err != nil {
		return "", fmt.Errorf("glob: %w", err)
	}
	if !matched {
		return "(no matches)", nil
	}
	return sb.String(), nil
}
