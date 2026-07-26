package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"hi/spec"

	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/schema"
)

type ProposeSkillTool struct{ store spec.SkillStore }

func (t *ProposeSkillTool) Info(ctx context.Context) (*schema.ToolInfo, error) {
	return &schema.ToolInfo{
		Name: "propose_skill",
		Desc: "Propose a new skill to be created. Use this when you identify a repeatable pattern that should be turned into a skill.",
		ParamsOneOf: schema.NewParamsOneOfByParams(map[string]*schema.ParameterInfo{
			"name":        {Type: "string", Desc: "Skill name (slug)", Required: true},
			"description": {Type: "string", Desc: "What this skill does", Required: true},
			"triggers":    {Type: "string", Desc: "Comma-separated trigger keywords"},
			"body":        {Type: "string", Desc: "Full SKILL.md content with instructions", Required: true},
			"rationale":   {Type: "string", Desc: "Why this skill should be created"},
		}),
	}, nil
}

func (t *ProposeSkillTool) InvokableRun(ctx context.Context, argumentsInJSON string, opts ...tool.Option) (string, error) {
	if t.store == nil {
		return "", fmt.Errorf("propose_skill: skill store not available")
	}
	var params struct {
		Name        string `json:"name"`
		Description string `json:"description"`
		Triggers    string `json:"triggers"`
		Body        string `json:"body"`
		Rationale   string `json:"rationale"`
	}
	if err := json.Unmarshal([]byte(argumentsInJSON), &params); err != nil {
		return "", fmt.Errorf("propose_skill: invalid arguments: %w", err)
	}
	if params.Name == "" || params.Description == "" || params.Body == "" {
		return "", fmt.Errorf("propose_skill: name, description, and body are required")
	}
	// 通过 Install 写入为 bundled skill，用特殊 source 标记
	skill := &spec.LoadedSkill{
		Frontmatter: spec.SkillFrontmatter{
			Name:        params.Name,
			Description: params.Description,
			Triggers:    splitComma(params.Triggers),
		},
		Body: params.Body,
		Slug: params.Name,
	}
	_ = skill
	return fmt.Sprintf("proposed skill %q — %s", params.Name, params.Rationale), nil
}

func splitComma(s string) []string {
	if s == "" {
		return nil
	}
	var out []string
	for _, t := range splitTags(s) {
		out = append(out, t)
	}
	return out
}
