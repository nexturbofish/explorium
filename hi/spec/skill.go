package spec

import "context"

type SkillStore interface {
	List(ctx context.Context) ([]*LoadedSkill, error)
	Get(ctx context.Context, slug string) (*LoadedSkill, error)
	Search(ctx context.Context, query string, limit int) ([]ScoredSkill, error)
	Install(ctx context.Context, source string) (*LoadedSkill, error)
	Delete(ctx context.Context, slug string) error
	AlwaysActive(ctx context.Context) ([]*LoadedSkill, error)
}

type ScoredSkill struct {
	Skill     *LoadedSkill
	Score     float64
	MatchedOn string
}
