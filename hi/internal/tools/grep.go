package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/schema"
)

type GrepTool struct{ workspace string }

func (t *GrepTool) Info(ctx context.Context) (*schema.ToolInfo, error) {
	return &schema.ToolInfo{
		Name: "grep",
		Desc: "Search file contents using a regular expression.",
		ParamsOneOf: schema.NewParamsOneOfByParams(map[string]*schema.ParameterInfo{
			"pattern": {Type: "string", Desc: "Regular expression to search for", Required: true},
			"include": {Type: "string", Desc: "File glob pattern to filter (e.g. *.go)"},
		}),
	}, nil
}

func (t *GrepTool) InvokableRun(ctx context.Context, argumentsInJSON string, opts ...tool.Option) (string, error) {
	var params struct {
		Pattern string `json:"pattern"`
		Include string `json:"include"`
	}
	if err := json.Unmarshal([]byte(argumentsInJSON), &params); err != nil {
		return "", fmt.Errorf("grep: invalid arguments: %w", err)
	}
	if params.Pattern == "" {
		return "", fmt.Errorf("grep: pattern is required")
	}
	re, err := regexp.Compile(params.Pattern)
	if err != nil {
		return "", fmt.Errorf("grep: invalid regex: %w", err)
	}
	var matches []string
	walkFn := func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d == nil || d.IsDir() {
			return nil
		}
		if params.Include != "" {
			matched, err := filepath.Match(params.Include, d.Name())
			if err != nil {
				return err
			}
			if !matched {
				return nil
			}
		}
		data, err := os.ReadFile(path)
		if err != nil {
			return nil // skip unreadable files silently
		}
		for _, line := range strings.Split(string(data), "\n") {
			if re.MatchString(line) {
				rel, err := filepath.Rel(t.workspace, path)
				if err != nil {
					return err
				}
				matches = append(matches, fmt.Sprintf("%s: %s", rel, line))
			}
		}
		return nil
	}
	if err := filepath.WalkDir(t.workspace, walkFn); err != nil {
		return "", fmt.Errorf("grep: %w", err)
	}
	if len(matches) == 0 {
		return "(no matches)", nil
	}
	return strings.Join(matches, "\n"), nil
}
