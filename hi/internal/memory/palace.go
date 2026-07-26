package memory

import (
	"context"
	"fmt"
	"hi/internal/lib"
	"hi/spec"
	"sort"
	"strings"
)

// Palace 按 zone 组织记忆，生成索引工 system prompt 使用
type Palace struct {
	store spec.MemoryStore
}

func NewPalace(store spec.MemoryStore) *Palace {
	return &Palace{store: store}
}

func (p *Palace) Zones(ctx context.Context) ([]string, error) {
	all, err := p.store.List(spec.MemoryScope(""))
	if err != nil {
		return nil, err
	}
	seen := make(map[string]bool)
	for _, m := range all {
		if m.Frontmatter.Zone != "" {
			seen[m.Frontmatter.Zone] = true
		}
	}
	zones := make([]string, 0, len(seen))
	for z := range seen {
		zones = append(zones, z)
	}
	sort.Strings(zones)
	return zones, nil
}

func (p *Palace) ReadZone(ctx context.Context, zone string) ([]*spec.LoadedMemory, error) {
	all, err := p.store.List(spec.MemoryScope(""))
	if err != nil {
		return nil, err
	}
	var result []*spec.LoadedMemory
	for _, m := range all {
		if m.Frontmatter.Zone == zone {
			result = append(result, m)
		}
	}
	return result, nil
}

func (p *Palace) Recall(ctx context.Context, topic string) (string, error) {
	all, err := p.store.List(spec.MemoryScope(""))
	if err != nil {
		return "", err
	}
	scored := TfidfSearch(all, topic, 5)
	if len(scored) == 0 {
		return "(no relevant memories found)", nil
	}
	var sb strings.Builder
	for _, s := range scored {
		sb.WriteString(fmt.Sprintf("- [%.2f] %s\n", s.Score, s.Memory.Content))
	}
	return sb.String(), nil
}

func (p *Palace) BuildIndex(ctx context.Context) (string, error) {
	zones, err := p.Zones(ctx)
	if err != nil {
		return "", err
	}
	var sb strings.Builder
	sb.WriteString("# Memory Palace Index\n\n")
	for _, z := range zones {
		mems, _ := p.ReadZone(ctx, z)
		if len(mems) == 0 {
			continue
		}
		sb.WriteString(fmt.Sprintf("## %s (%d)\n", z, len(mems)))
		for _, m := range mems {
			sb.WriteString(fmt.Sprintf("- %s: %s\n", m.Frontmatter.ID, lib.Truncate(m.Content, 80)))

		}
		sb.WriteString("\n")
	}
	return sb.String(), nil
}
