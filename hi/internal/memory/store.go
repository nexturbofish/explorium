package memory

import (
	"bytes"
	"context"
	"fmt"
	"hi/internal/lib"
	"hi/spec"
	"os"
	"path/filepath"
	"strings"
	"time"

	"go.yaml.in/yaml/v4"
)

type MemoryStore struct {
	dir              string // ~/.hi/memories/
	activeWindowDays int    // 0 = 不过期
}

func NewMemoryStore(dir string, activeWindowDays int) spec.MemoryStore {
	return &MemoryStore{dir: dir, activeWindowDays: activeWindowDays}
}

// Save implements [spec.MemoryStore].
func (s *MemoryStore) Save(memory *spec.LoadedMemory) error {
	path := filepath.Join(s.dir, memory.Frontmatter.ID+".md")
	fm := map[string]any{
		"id":           memory.Frontmatter.ID,
		"tags":         memory.Frontmatter.Tags,
		"zone":         memory.Frontmatter.Zone,
		"pinned":       memory.Frontmatter.Pinned,
		"source":       memory.Frontmatter.Source,
		"created_at":   memory.Frontmatter.CreatedAt,
		"accessed_at":  memory.Frontmatter.AccessedAt,
		"access_count": memory.Frontmatter.AccessCount,
	}
	return lib.AtomicWriteFile(path, formatMemoryFile(fm, memory.Content))
}

// Load implements [spec.MemoryStore].
func (s *MemoryStore) Load(id string) (*spec.LoadedMemory, error) {
	path := filepath.Join(s.dir, id+".md")
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, &MemoryStoreError{Op: "load", ID: id, Err: err}
	}
	fm, body, err := lib.ParseFrontmatter(string(data))
	if err != nil {
		return nil, &MemoryStoreError{Op: "load", ID: id, Err: err}
	}
	return &spec.LoadedMemory{
		Frontmatter: spec.MemoryFrontmatter{
			ID:          lib.GetString(fm, "id"),
			Tags:        lib.GetStrings(fm, "tags"),
			Zone:        lib.GetString(fm, "zone"),
			Pinned:      lib.GetBool(fm, "pinned"),
			Source:      spec.MemorySource(lib.GetString(fm, "source")),
			CreatedAt:   lib.GetTime(fm, "created_at"),
			AccessedAt:  lib.GetTime(fm, "accessed_at"),
			AccessCount: lib.GetInt(fm, "access_count"),
		},
		Content: body,
	}, nil
}

// Delete implements [spec.MemoryStore].
func (s *MemoryStore) Delete(id string) error {
	path := filepath.Join(s.dir, id+".md")
	if err := os.Remove(path); err != nil && !os.IsNotExist(err) {
		return &MemoryStoreError{Op: "delete", ID: id, Err: err}
	}
	return nil
}

// List implements [spec.MemoryStore].
func (s *MemoryStore) List(scope spec.MemoryScope) ([]*spec.LoadedMemory, error) {
	entries, err := os.ReadDir(s.dir)
	if err != nil {
		return nil, &MemoryStoreError{Op: "list", Err: err}
	}
	var result []*spec.LoadedMemory
	for _, e := range entries {
		if !strings.HasSuffix(e.Name(), ".md") {
			continue
		}
		m, err := s.Load(strings.TrimSuffix(e.Name(), ".md"))
		if err != nil {
			continue
		}
		if scope != "" && m.Frontmatter.Scope != scope {
			continue
		}
		result = append(result, m)
	}
	return result, nil
}

// Search implements [spec.MemoryStore].
func (s *MemoryStore) Search(ctx context.Context, query string, limit int) ([]spec.ScoredMemory, error) {
	all, err := s.List(spec.MemoryScope(""))
	if err != nil {
		return nil, err
	}
	return TfidfSearch(all, query, limit), nil
}

// Touch 更新记忆的访问时间和计数，应在每次 Load 后调用
func (s *MemoryStore) Touch(m *spec.LoadedMemory) error {
	m.Frontmatter.AccessedAt = time.Now()
	m.Frontmatter.AccessCount++
	return s.Save(m)
}

// ListPinned implements [spec.MemoryStore].
func (s *MemoryStore) ListPinned(ctx context.Context) ([]*spec.LoadedMemory, error) {
	all, err := s.List(spec.MemoryScope(""))
	if err != nil {
		return nil, err
	}
	var pinned []*spec.LoadedMemory
	for _, m := range all {
		if m.Frontmatter.Pinned {
			pinned = append(pinned, m)
		}
	}
	return pinned, nil
}

// ListActive implements [spec.MemoryStore].
func (s *MemoryStore) ListActive(ctx context.Context) ([]*spec.LoadedMemory, error) {
	all, err := s.List(spec.MemoryScope(""))
	if err != nil {
		return nil, err
	}
	var active []*spec.LoadedMemory
	now := time.Now()
	for _, m := range all {
		if m.Frontmatter.Pinned {
			active = append(active, m)
			continue
		}
		if m.Frontmatter.AccessCount == 0 {
			continue // 从未被访问过，不视为 active
		}
		if s.activeWindowDays > 0 {
			cutoff := now.Add(-time.Duration(s.activeWindowDays) * 24 * time.Hour)
			if m.Frontmatter.AccessedAt.Before(cutoff) {
				continue
			}
		}
		active = append(active, m)
	}
	return active, nil
}

// Pin implements [spec.MemoryStore].
func (s *MemoryStore) Pin(id string) error {
	m, err := s.Load(id)
	if err != nil {
		return err
	}
	m.Frontmatter.Pinned = true
	return s.Save(m)
}

// Unpin implements [spec.MemoryStore].
func (s *MemoryStore) Unpin(id string) error {
	m, err := s.Load(id)
	if err != nil {
		return err
	}
	m.Frontmatter.Pinned = false
	return s.Save(m)
}

// CheckConflict 检测新 memory body 是否与已有记忆冲突
// 使用 TfidfSearch 做相似度比较，threshold 默认 0.85
func (s *MemoryStore) CheckConflict(body string, threshold float64) (*spec.LoadedMemory, float64, error) {
	active, _ := s.ListActive(context.Background())
	scored := TfidfSearch(toLoadedMemories(active), body, 1)
	if len(scored) > 0 && scored[0].Score > threshold {
		return scored[0].Memory, scored[0].Score, &MemoryStoreError{
			Op:  "save",
			Err: fmt.Errorf("conflict with %s (similarity %.2f)", scored[0].Memory.Frontmatter.ID, scored[0].Score),
		}
	}
	return nil, 0, nil
}

func (s *MemoryStore) CheckConflictTFIDF(body string, threshold float64) (*spec.LoadedMemory, float64, error) {
	return s.CheckConflict(body, threshold)
}

var _ spec.MemoryStore = (*MemoryStore)(nil)

func formatMemoryFile(fm map[string]any, body string) []byte {
	var b bytes.Buffer
	b.WriteString("---\n")
	enc := yaml.NewEncoder(&b)
	enc.Encode(fm)
	enc.Close()
	b.WriteString("---\n")
	b.WriteString(body)
	return b.Bytes()
}

func toLoadedMemories(mems []*spec.LoadedMemory) []*spec.LoadedMemory { return mems }
