package tools

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/schema"
)

type ThinkTool struct{}

func (t *ThinkTool) Info(ctx context.Context) (*schema.ToolInfo, error) {
	return &schema.ToolInfo{
		Name: "think",
		Desc: "Use this tool to reason step-by-step before answering. The model's chain-of-thought is preserved in the tool output.",
		ParamsOneOf: schema.NewParamsOneOfByParams(map[string]*schema.ParameterInfo{
			"thought": {Type: "string", Desc: "Your step-by-step reasoning", Required: true},
		}),
	}, nil
}

func (t *ThinkTool) InvokableRun(ctx context.Context, argumentsInJSON string, opts ...tool.Option) (string, error) {
	var params struct{ Thought string }
	if err := json.Unmarshal([]byte(argumentsInJSON), &params); err != nil {
		return "", fmt.Errorf("think: invalid arguments: %w", err)
	}
	return params.Thought, nil
}
