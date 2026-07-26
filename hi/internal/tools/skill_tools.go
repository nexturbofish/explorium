package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"hi/internal/skills"
	"hi/spec"
	"strings"

	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/schema"
)

type SkillListTool struct{ store spec.SkillStore }

func (t *SkillListTool) Info(ctx context.Context) (*schema.ToolInfo, error) {
	return &schema.ToolInfo{
		Name: "skill_list",
		Desc: "List all installed skills with their descriptions and active status.",
		ParamsOneOf: schema.NewParamsOneOfByParams(map[string]*schema.ParameterInfo{
			"always_active": {Type: "string", Desc: "Filter: 'yes' for always-active skills only, 'no' for non-always, empty for all"},
		}),
	}, nil
}

func (t *SkillListTool) InvokableRun(ctx context.Context, argumentsInJSON string, opts ...tool.Option) (string, error) {
	if t.store == nil {
		return "", fmt.Errorf("skill_list: skill store not available")
	}
	var params struct{ AlwaysActive string }
	json.Unmarshal([]byte(argumentsInJSON), &params)

	all, err := t.store.List(ctx)
	if err != nil {
		return "", fmt.Errorf("skill_list: %w", err)
	}
	var filtered []*spec.LoadedSkill
	for _, s := range all {
		switch params.AlwaysActive {
		case "yes":
			if !s.Frontmatter.AlwaysActive {
				continue
			}
		case "no":
			if s.Frontmatter.AlwaysActive {
				continue
			}
		}
		filtered = append(filtered, s)
	}
	if len(filtered) == 0 {
		return "(no skills)", nil
	}
	var sb strings.Builder
	for _, s := range filtered {
		always := ""
		if s.Frontmatter.AlwaysActive {
			always = " [always]"
		}
		sb.WriteString(fmt.Sprintf("- %s: %s%s\n", s.Slug, s.Frontmatter.Description, always))
	}
	return sb.String(), nil
}

type SkillReadTool struct{ store spec.SkillStore }

func (t *SkillReadTool) Info(ctx context.Context) (*schema.ToolInfo, error) {
	return &schema.ToolInfo{
		Name: "skill_read",
		Desc: "Read the full content of an installed skill by slug.",
		ParamsOneOf: schema.NewParamsOneOfByParams(map[string]*schema.ParameterInfo{
			"slug": {Type: "string", Desc: "Skill slug", Required: true},
		}),
	}, nil
}

func (t *SkillReadTool) InvokableRun(ctx context.Context, argumentsInJSON string, opts ...tool.Option) (string, error) {
	if t.store == nil {
		return "", fmt.Errorf("skill_read: skill store not available")
	}
	var params struct{ Slug string }
	if err := json.Unmarshal([]byte(argumentsInJSON), &params); err != nil {
		return "", fmt.Errorf("skill_read: invalid arguments: %w", err)
	}
	if params.Slug == "" {
		return "", fmt.Errorf("skill_read: slug is required")
	}
	skill, err := t.store.Get(ctx, params.Slug)
	if err != nil {
		return "", fmt.Errorf("skill_read: %w", err)
	}
	return skill.Body, nil
}

type SkillInstallTool struct{ store spec.SkillStore }

func (t *SkillInstallTool) Info(ctx context.Context) (*schema.ToolInfo, error) {
	return &schema.ToolInfo{
		Name: "skill_install",
		Desc: "Install a skill from GitHub. Use format: owner/repo@skill-name or github:owner/repo/path.",
		ParamsOneOf: schema.NewParamsOneOfByParams(map[string]*schema.ParameterInfo{
			"source": {Type: "string", Desc: "GitHub source (e.g. 'owner/repo@skill-name')", Required: true},
		}),
	}, nil
}

func (t *SkillInstallTool) InvokableRun(ctx context.Context, argumentsInJSON string, opts ...tool.Option) (string, error) {
	if t.store == nil {
		return "", fmt.Errorf("skill_install: skill store not available")
	}
	var params struct{ Source string }
	if err := json.Unmarshal([]byte(argumentsInJSON), &params); err != nil {
		return "", fmt.Errorf("skill_install: invalid arguments: %w", err)
	}
	if params.Source == "" {
		return "", fmt.Errorf("skill_install: source is required")
	}
	skill, err := t.store.Install(ctx, params.Source)
	if err != nil {
		return "", fmt.Errorf("skill_install: %w", err)
	}
	return fmt.Sprintf("installed skill %q (%s)", skill.Slug, skill.Frontmatter.Description), nil
}

type SkillDeleteTool struct{ store spec.SkillStore }

func (t *SkillDeleteTool) Info(ctx context.Context) (*schema.ToolInfo, error) {
	return &schema.ToolInfo{
		Name: "skill_delete",
		Desc: "Delete an installed skill by slug. Bundled skills cannot be deleted.",
		ParamsOneOf: schema.NewParamsOneOfByParams(map[string]*schema.ParameterInfo{
			"slug": {Type: "string", Desc: "Skill slug", Required: true},
		}),
	}, nil
}

func (t *SkillDeleteTool) InvokableRun(ctx context.Context, argumentsInJSON string, opts ...tool.Option) (string, error) {
	if t.store == nil {
		return "", fmt.Errorf("skill_delete: skill store not available")
	}
	var params struct{ Slug string }
	if err := json.Unmarshal([]byte(argumentsInJSON), &params); err != nil {
		return "", fmt.Errorf("skill_delete: invalid arguments: %w", err)
	}
	if params.Slug == "" {
		return "", fmt.Errorf("skill_delete: slug is required")
	}
	for _, b := range skills.BundledSkills {
		if strings.EqualFold(params.Slug, b) {
			return "", fmt.Errorf("skill_delete: cannot delete bundled skill %q", params.Slug)
		}
	}
	if err := t.store.Delete(ctx, params.Slug); err != nil {
		return "", fmt.Errorf("skill_delete: %w", err)
	}
	return fmt.Sprintf("deleted skill %q", params.Slug), nil
}
