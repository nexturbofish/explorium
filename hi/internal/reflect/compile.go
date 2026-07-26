package reflect

import (
	"context"
	"fmt"
	"hi/internal/lib"
	"hi/spec"
	"strings"
)

func CompileProfile(ctx context.Context, store spec.MemoryStore) (string, error) {
	all, err := store.ListActive(ctx)
	if err != nil {
		return "", err
	}
	var sb strings.Builder
	sb.WriteString("# Learning Profile\n\n")
	zones := make(map[string][]string)
	for _, m := range all {
		zones[m.Frontmatter.Zone] = append(zones[m.Frontmatter.Zone], m.Content)
	}
	for zone, entries := range zones {
		sb.WriteString(fmt.Sprintf("## %s (%d)\n", zone, len(entries)))
		for _, e := range entries {
			sb.WriteString(fmt.Sprintf("- %s\n", lib.Truncate(e, 120)))
		}
		sb.WriteString("\n")
	}
	return sb.String(), nil
}

func CompilePalaceIndex(ctx context.Context, store spec.MemoryStore) (string, error) {
	all, err := store.ListActive(ctx)
	if err != nil {
		return "", err
	}
	var sb strings.Builder
	sb.WriteString("# Palace Index\n\n")
	for _, m := range all {
		sb.WriteString(fmt.Sprintf("- [%s] %s\n", m.Frontmatter.Zone, lib.Truncate(m.Content, 80)))
	}
	return sb.String(), nil
}

func CompileZoneSummary(ctx context.Context, store spec.MemoryStore, zone string) (string, error) {
	all, err := store.List(spec.MemoryScope(""))
	if err != nil {
		return "", err
	}
	var zoneMems []*spec.LoadedMemory
	for _, m := range all {
		if m.Frontmatter.Zone == zone {
			zoneMems = append(zoneMems, m)
		}
	}
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("# Zone: %s (%d entries)\n\n", zone, len(zoneMems)))
	for _, m := range zoneMems {
		sb.WriteString(fmt.Sprintf("- %s\n", lib.Truncate(m.Content, 120)))
	}
	return sb.String(), nil
}
