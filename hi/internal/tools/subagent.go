package tools

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/schema"
)

type SubagentTool struct{}

func (t *SubagentTool) Info(ctx context.Context) (*schema.ToolInfo, error) {
	return &schema.ToolInfo{
		Name: "subagent",
		Desc: "Spawn a sub-agent for an independent task. The sub-agent has access to all tools.",
		ParamsOneOf: schema.NewParamsOneOfByParams(map[string]*schema.ParameterInfo{
			"task": {Type: "string", Desc: "The task description for the sub-agent", Required: true},
		}),
	}, nil
}

func (t *SubagentTool) InvokableRun(ctx context.Context, argumentsInJSON string, opts ...tool.Option) (string, error) {
	var params struct{ Task string }
	if err := json.Unmarshal([]byte(argumentsInJSON), &params); err != nil {
		return "", fmt.Errorf("subagent: invalid arguments: %w", err)
	}
	if params.Task == "" {
		return "", fmt.Errorf("subagent: task is required")
	}
	return params.Task, nil
}
