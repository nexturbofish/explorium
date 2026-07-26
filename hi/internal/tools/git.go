package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"os/exec"

	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/schema"
	"github.com/google/shlex"
)

type GitTool struct{ workspace string }

func (t *GitTool) Info(ctx context.Context) (*schema.ToolInfo, error) {
	return &schema.ToolInfo{
		Name: "git",
		Desc: "Run a git command in the workspace repository.",
		ParamsOneOf: schema.NewParamsOneOfByParams(map[string]*schema.ParameterInfo{
			"args": {Type: "string", Desc: "Git arguments (e.g. 'diff --stat')", Required: true},
		}),
	}, nil
}

func (t *GitTool) InvokableRun(ctx context.Context, argumentsInJSON string, opts ...tool.Option) (string, error) {
	var params struct{ Args string }
	if err := json.Unmarshal([]byte(argumentsInJSON), &params); err != nil {
		return "", fmt.Errorf("git: invalid arguments: %w", err)
	}
	if params.Args == "" {
		return "", fmt.Errorf("git: args is required")
	}
	args, err := shlex.Split(params.Args)
	if err != nil {
		return "", fmt.Errorf("git: invalid args: %w", err)
	}
	cmd := exec.CommandContext(ctx, "git", args...)
	cmd.Dir = t.workspace
	out, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("exit: %v\n%s", err, string(out))
	}
	return string(out), nil
}
