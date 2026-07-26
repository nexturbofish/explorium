package tools

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/schema"
)

type WebFetchTool struct{ ctx *WebToolsContext }

func (t *WebFetchTool) Info(ctx context.Context) (*schema.ToolInfo, error) {
	return &schema.ToolInfo{
		Name: "web_fetch",
		Desc: "Fetch a URL and return its content as text (max 100KB).",
		ParamsOneOf: schema.NewParamsOneOfByParams(map[string]*schema.ParameterInfo{
			"url": {Type: "string", Desc: "The URL to fetch", Required: true},
		}),
	}, nil
}

func (t *WebFetchTool) InvokableRun(ctx context.Context, argumentsInJSON string, opts ...tool.Option) (string, error) {
	var params struct{ URL string }
	if err := json.Unmarshal([]byte(argumentsInJSON), &params); err != nil {
		return "", fmt.Errorf("web_fetch: invalid arguments: %w", err)
	}
	if params.URL == "" {
		return "", fmt.Errorf("web_fetch: url is required")
	}
	return t.ctx.Fetch(ctx, params.URL)
}
