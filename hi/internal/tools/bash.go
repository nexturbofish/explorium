package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"os/exec"
	"runtime"
	"time"

	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/schema"
)

type BashTool struct{ workspace string }

func (t *BashTool) Info(ctx context.Context) (*schema.ToolInfo, error) {
	shell := "sh"
	if runtime.GOOS == "windows" {
		shell = "cmd"
	}
	return &schema.ToolInfo{
		Name: "bash",
		Desc: fmt.Sprintf("Execute a shell command (OS: %s, shell: %s). Returns stdout and stderr.", runtime.GOOS, shell),
		ParamsOneOf: schema.NewParamsOneOfByParams(map[string]*schema.ParameterInfo{
			"command": {Type: "string", Desc: fmt.Sprintf("Shell command to execute (%s syntax)", shell), Required: true},
			"timeout": {Type: "number", Desc: "Timeout in seconds (default 30)"},
		}),
	}, nil
}

func (t *BashTool) InvokableRun(ctx context.Context, argumentsInJSON string, opts ...tool.Option) (string, error) {
	var params struct {
		Command string  `json:"command"`
		Timeout float64 `json:"timeout"`
	}
	if err := json.Unmarshal([]byte(argumentsInJSON), &params); err != nil {
		return "", fmt.Errorf("bash: invalid arguments: %w", err)
	}
	if params.Command == "" {
		return "", fmt.Errorf("bash: command is required")
	}
	if params.Timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, time.Duration(params.Timeout*float64(time.Second)))
		defer cancel()
	}

	// 跨平台 shell 选择
	var cmd *exec.Cmd
	if runtime.GOOS == "windows" {
		cmd = exec.CommandContext(ctx, "cmd", "/c", params.Command)
	} else {
		cmd = exec.CommandContext(ctx, "sh", "-c", params.Command)
	}
	out, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("exit: %v\n%s", err, string(out))
	}
	return string(out), nil
}
