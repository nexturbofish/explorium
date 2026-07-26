package skills

import (
	"context"
	"fmt"
	"hi/internal/lib"
	"os"
	"path/filepath"
	"strings"
)

// CreateSkillOpts 本地创建技能的参数。
type CreateSkillOpts struct {
	Name        string
	Description string
	Triggers    []string
	AlwaysActive bool
	Body        string
	Version     string
	Author      string
}

// CreateSkill 在本地创建或编辑一个技能文件。
// 如果技能已存在且 overwrite 为 false，返回错误。
func CreateSkill(ctx context.Context, rootDir string, opts CreateSkillOpts, overwrite bool) error {
	if opts.Name == "" {
		return fmt.Errorf("skill name is required")
	}
	if opts.Description == "" {
		return fmt.Errorf("skill description is required")
	}
	if opts.Body == "" {
		return fmt.Errorf("skill body is required")
	}

	slug := opts.Name
	skillDir := filepath.Join(rootDir, filepath.Clean(slug))
	if !strings.HasPrefix(filepath.Clean(skillDir), filepath.Clean(rootDir)+string(filepath.Separator)) && skillDir != filepath.Clean(rootDir) {
		return fmt.Errorf("skill name %q escapes base directory", slug)
	}

	finalPath := filepath.Join(skillDir, "SKILL.md")
	if !overwrite {
		if _, err := os.Stat(finalPath); err == nil {
			return fmt.Errorf("skill %q already exists (use overwrite=true to replace)", slug)
		}
	}

	if err := os.MkdirAll(skillDir, 0755); err != nil {
		return fmt.Errorf("create skill dir: %w", err)
	}

	fm := map[string]any{
		"name":        opts.Name,
		"description": opts.Description,
	}
	if len(opts.Triggers) > 0 {
		fm["triggers"] = opts.Triggers
	}
	if opts.AlwaysActive {
		fm["always_active"] = true
	}
	if opts.Version != "" {
		fm["version"] = opts.Version
	}
	if opts.Author != "" {
		fm["author"] = opts.Author
	}

	tmpFile, err := os.CreateTemp(skillDir, "SKILL.md.tmp.*")
	if err != nil {
		return fmt.Errorf("create temp file: %w", err)
	}
	tmpPath := tmpFile.Name()

	if err := lib.WriteFrontmatter(tmpFile, fm, opts.Body); err != nil {
		tmpFile.Close()
		os.Remove(tmpPath)
		return fmt.Errorf("write SKILL.md: %w", err)
	}
	if err := tmpFile.Close(); err != nil {
		os.Remove(tmpPath)
		return fmt.Errorf("close temp file: %w", err)
	}

	if err := os.Rename(tmpPath, finalPath); err != nil {
		os.Remove(tmpPath)
		return fmt.Errorf("rename %s -> %s: %w", tmpPath, finalPath, err)
	}

	return nil
}
