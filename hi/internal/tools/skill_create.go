package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"hi/internal/skills"
	"hi/spec"
	"os"
	"path/filepath"

	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/schema"
)

type SkillCreateTool struct {
	Store spec.SkillStore
}

func (t *SkillCreateTool) Info(ctx context.Context) (*schema.ToolInfo, error) {
	return &schema.ToolInfo{
		Name: "skill_create",
		Desc: "Create a new skill or overwrite an existing one. Creates ~/.hi/skills/<name>/SKILL.md with frontmatter and body. Use this when the user asks to create a skill.",
		ParamsOneOf: schema.NewParamsOneOfByParams(map[string]*schema.ParameterInfo{
			"name":         {Type: "string", Desc: "Skill name (slug), used as directory name", Required: true},
			"description":  {Type: "string", Desc: "Short description of what this skill does", Required: true},
			"body":         {Type: "string", Desc: "Full Markdown body with instructions and workflow", Required: true},
			"triggers":     {Type: "string", Desc: "Comma-separated trigger keywords"},
			"always_active": {Type: "string", Desc: "Set to 'true' if this skill should always be active"},
			"version":      {Type: "string", Desc: "Version number (optional)"},
			"author":       {Type: "string", Desc: "Author name (optional)"},
			"overwrite":    {Type: "string", Desc: "Set to 'true' to overwrite an existing skill"},
		}),
	}, nil
}

func (t *SkillCreateTool) InvokableRun(ctx context.Context, argumentsInJSON string, opts ...tool.Option) (string, error) {
	var params struct {
		Name         string `json:"name"`
		Description  string `json:"description"`
		Body         string `json:"body"`
		Triggers     string `json:"triggers"`
		AlwaysActive string `json:"always_active"`
		Version      string `json:"version"`
		Author       string `json:"author"`
		Overwrite    string `json:"overwrite"`
	}
	if err := json.Unmarshal([]byte(argumentsInJSON), &params); err != nil {
		return "", fmt.Errorf("skill_create: invalid arguments: %w", err)
	}
	if params.Name == "" || params.Description == "" || params.Body == "" {
		return "", fmt.Errorf("skill_create: name, description, and body are required")
	}

	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("skill_create: cannot get home dir: %w", err)
	}
	rootDir := filepath.Join(home, ".hi", "skills")

	skOpts := skills.CreateSkillOpts{
		Name:        params.Name,
		Description: params.Description,
		Triggers:    splitComma(params.Triggers),
		AlwaysActive: params.AlwaysActive == "true" || params.AlwaysActive == "yes",
		Body:        params.Body,
		Version:     params.Version,
		Author:      params.Author,
	}
	overwrite := params.Overwrite == "true" || params.Overwrite == "yes"

	if err := skills.CreateSkill(ctx, rootDir, skOpts, overwrite); err != nil {
		return "", fmt.Errorf("skill_create: %w", err)
	}

	return fmt.Sprintf("Created skill %q (%s)", params.Name, params.Description), nil
}
