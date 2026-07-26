package skills

import (
	"context"
	_ "embed"
	"fmt"
	"hi/internal/lib"
	"hi/spec"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

//go:embed bundled/skill-creator/SKILL.md
var bundledSkillCreator string

//go:embed bundled/find-skills/SKILL.md
var bundledFindSkills string

type SkillStore struct {
	rootDir string
}

func NewSkillStore(rootDir string) *SkillStore {
	return &SkillStore{rootDir: rootDir}
}

// List implements [spec.SkillStore].
func (s *SkillStore) List(ctx context.Context) ([]*spec.LoadedSkill, error) {
	entries, err := os.ReadDir(s.rootDir)
	if err != nil {
		return nil, fmt.Errorf("list skills: %w", err)
	}
	var result []*spec.LoadedSkill
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		skill, err := s.Get(ctx, e.Name())
		if err != nil {
			continue
		}
		result = append(result, skill)
	}
	return result, nil
}

// Get implements [spec.SkillStore].
func (s *SkillStore) Get(ctx context.Context, slug string) (*spec.LoadedSkill, error) {
	path := filepath.Join(s.rootDir, slug, "SKILL.md")
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("get skill %q: %w", slug, err)
	}
	fm, body, err := lib.ParseFrontmatter(string(data))
	if err != nil {
		return nil, fmt.Errorf("parse skill %q: %w", slug, err)
	}
	return &spec.LoadedSkill{
		Frontmatter: spec.SkillFrontmatter{
			Name:         lib.GetString(fm, "name"),
			Description:  lib.GetString(fm, "description"),
			Triggers:     lib.GetStrings(fm, "triggers"),
			AlwaysActive: lib.GetBool(fm, "always_active"),
		},
		Body: body,
		Slug: slug,
	}, nil
}

// Search implements [spec.SkillStore].
func (s *SkillStore) Search(ctx context.Context, query string, limit int) ([]spec.ScoredSkill, error) {
	all, err := s.List(ctx)
	if err != nil {
		return nil, err
	}
	q := strings.ToLower(query)
	var scored []spec.ScoredSkill
	for _, sk := range all {
		score := float64(0)
		name := strings.ToLower(sk.Frontmatter.Name)
		desc := strings.ToLower(sk.Frontmatter.Description)
		body := strings.ToLower(sk.Body)
		for _, token := range strings.Fields(q) {
			if strings.Contains(name, token) {
				score += 0.5
			}
			if strings.Contains(desc, token) {
				score += 0.3
			}
			if strings.Contains(body, token) {
				score += 0.2
			}
		}
		if score > 0 {
			scored = append(scored, spec.ScoredSkill{
				Skill: sk,
				Score: score,
			})
		}
	}
	sort.Slice(scored, func(i, j int) bool {
		return scored[i].Score > scored[j].Score
	})
	if len(scored) > limit {
		scored = scored[:limit]
	}
	return scored, nil
}

// Install implements [spec.SkillStore].
func (s *SkillStore) Install(ctx context.Context, source string) (*spec.LoadedSkill, error) {
	out, err := InstallFromGitHub(ctx, source, s.rootDir)
	if err != nil {
		return nil, err
	}
	return s.Get(ctx, out.Name)
}

// Delete implements [spec.SkillStore].
func (s *SkillStore) Delete(ctx context.Context, slug string) error {
	_, err := DeleteSkill(ctx, s.rootDir, slug)
	return err
}

func (s *SkillStore) AlwaysActive(ctx context.Context) ([]*spec.LoadedSkill, error) {
	all, _ := s.List(ctx)
	active := make([]*spec.LoadedSkill, 0)
	for _, skill := range all {
		if skill.Frontmatter.AlwaysActive {
			active = append(active, skill)
		}
	}
	return active, nil
}

// EnsureBundled checks that all bundled skills exist, installing any that are missing.
// Should be called once at startup after the root dir is known.
func (s *SkillStore) EnsureBundled(ctx context.Context) {
	for _, slug := range BundledSkills {
		if _, err := s.Get(ctx, slug); err == nil {
			continue
		}
		switch slug {
		case "skill-creator":
			s.installBundled(ctx, slug, bundledSkillCreator)
		case "find-skills":
			s.installBundled(ctx, slug, bundledFindSkills)
		case "memory-palace":
			body := `# Memory Palace Protocol

Your memories are organized into zones. The palace index (zone map) is in the system prompt.

## Zones
- core — stable user identity, preferences, principles
- work — current focus, recent activity
- project:<name> — per-project conventions
- episode — session summaries
- general — uncategorized (default)

## Navigation
1. Check the palace index to see what zones exist
2. Use palace_read_zone to load a specific zone's content
3. Don't guess — load the zone before answering questions about preferences or conventions

## Saving
When using memory_save, set the zone parameter:
- User preferences, identity → core
- Current tasks, recent decisions → work
- Project-specific → project:<name>
- Everything else → general`
			s.installBundled(ctx, slug, body)
		}
	}
}

// installBundled writes a bundled skill's SKILL.md into the store.
func (s *SkillStore) installBundled(ctx context.Context, slug, raw string) {
	fm, body, err := lib.ParseFrontmatter(raw)
	if err != nil {
		return
	}
	skillDir := filepath.Join(s.rootDir, slug)
	if err := os.MkdirAll(skillDir, 0755); err != nil {
		return
	}
	finalPath := filepath.Join(skillDir, "SKILL.md")
	tmp, err := os.CreateTemp(skillDir, "SKILL.md.tmp.*")
	if err != nil {
		return
	}
	tmpPath := tmp.Name()
	if err := lib.WriteFrontmatter(tmp, fm, body); err != nil {
		tmp.Close()
		os.Remove(tmpPath)
		return
	}
	tmp.Close()
	if err := os.Rename(tmpPath, finalPath); err != nil {
		os.Remove(tmpPath)
	}
}

var _ (spec.SkillStore) = (*SkillStore)(nil)
