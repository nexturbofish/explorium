package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"hi/internal/memory"
	"hi/spec"
	"strings"

	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/schema"
)

type PalaceZonesTool struct{ store spec.MemoryStore }

func (t *PalaceZonesTool) Info(ctx context.Context) (*schema.ToolInfo, error) {
	return &schema.ToolInfo{
		Name: "palace_zones",
		Desc: "List all memory palace zones.",
	}, nil
}

func (t *PalaceZonesTool) InvokableRun(ctx context.Context, argumentsInJSON string, opts ...tool.Option) (string, error) {
	if t.store == nil {
		return "", fmt.Errorf("palace_zones: memory store not available")
	}
	palace := memory.NewPalace(t.store)
	zones, err := palace.Zones(ctx)
	if err != nil {
		return "", fmt.Errorf("palace_zones: %w", err)
	}
	if len(zones) == 0 {
		return "(no zones)", nil
	}
	return strings.Join(zones, "\n"), nil
}

type PalaceReadZoneTool struct{ store spec.MemoryStore }

func (t *PalaceReadZoneTool) Info(ctx context.Context) (*schema.ToolInfo, error) {
	return &schema.ToolInfo{
		Name: "palace_read_zone",
		Desc: "Read all memories in a palace zone.",
		ParamsOneOf: schema.NewParamsOneOfByParams(map[string]*schema.ParameterInfo{
			"zone": {Type: "string", Desc: "Zone name", Required: true},
		}),
	}, nil
}

func (t *PalaceReadZoneTool) InvokableRun(ctx context.Context, argumentsInJSON string, opts ...tool.Option) (string, error) {
	if t.store == nil {
		return "", fmt.Errorf("palace_read_zone: memory store not available")
	}
	var params struct{ Zone string }
	if err := json.Unmarshal([]byte(argumentsInJSON), &params); err != nil {
		return "", fmt.Errorf("palace_read_zone: invalid arguments: %w", err)
	}
	palace := memory.NewPalace(t.store)
	mems, err := palace.ReadZone(ctx, params.Zone)
	if err != nil {
		return "", fmt.Errorf("palace_read_zone: %w", err)
	}
	var sb strings.Builder
	for _, m := range mems {
		sb.WriteString(fmt.Sprintf("- %s: %s\n", m.Frontmatter.ID, m.Content))
	}
	return sb.String(), nil
}

type PalaceRecallTool struct{ store spec.MemoryStore }

func (t *PalaceRecallTool) Info(ctx context.Context) (*schema.ToolInfo, error) {
	return &schema.ToolInfo{
		Name: "palace_recall",
		Desc: "Search memories across all zones by topic similarity.",
		ParamsOneOf: schema.NewParamsOneOfByParams(map[string]*schema.ParameterInfo{
			"topic": {Type: "string", Desc: "Topic to search for", Required: true},
		}),
	}, nil
}

func (t *PalaceRecallTool) InvokableRun(ctx context.Context, argumentsInJSON string, opts ...tool.Option) (string, error) {
	if t.store == nil {
		return "", fmt.Errorf("palace_recall: memory store not available")
	}
	var params struct{ Topic string }
	if err := json.Unmarshal([]byte(argumentsInJSON), &params); err != nil {
		return "", fmt.Errorf("palace_recall: invalid arguments: %w", err)
	}
	palace := memory.NewPalace(t.store)
	return palace.Recall(ctx, params.Topic)
}
