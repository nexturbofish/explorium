package reflect

import (
	"context"
	"encoding/json"
	"fmt"
	"hi/internal/lib"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type DeferredCandidate struct {
	Type      string // "skill" | "memory"
	Data      any
	CreatedAt time.Time
	Score     float64 // 累积分数
	TimesSeen int
}

type DeferredQueue struct {
	dir string // ~/.hi/deferred-candidates/
}

func NewDeferredQueue(dir string) *DeferredQueue {
	return &DeferredQueue{dir: dir}
}

func (q *DeferredQueue) Add(candidate *DeferredCandidate) error {
	data, err := json.Marshal(candidate)
	if err != nil {
		return err
	}
	path := filepath.Join(q.dir, fmt.Sprintf("%s_%d.json", candidate.Type, time.Now().UnixNano()))
	return lib.AtomicWriteFile(path, data)
}

func (q *DeferredQueue) Promote(ctx context.Context, threshold int) []*DeferredCandidate {
	entries, _ := os.ReadDir(q.dir)
	var promoted []*DeferredCandidate
	for _, e := range entries {
		if !strings.HasSuffix(e.Name(), ".json") {
			continue
		}
		data, err := os.ReadFile(filepath.Join(q.dir, e.Name()))
		if err != nil {
			continue
		}
		var c DeferredCandidate
		if err := json.Unmarshal(data, &c); err != nil {
			continue
		}
		c.TimesSeen++
		if c.TimesSeen >= threshold {
			promoted = append(promoted, &c)
			os.Remove(filepath.Join(q.dir, e.Name()))
		} else {
			q.Add(&c)
		}
	}
	return promoted
}

func (q *DeferredQueue) Purge(maxAge time.Duration) error {
	entries, err := os.ReadDir(q.dir)
	if err != nil {
		return err
	}
	cutoff := time.Now().Add(-maxAge)
	for _, e := range entries {
		info, err := e.Info()
		if err != nil {
			continue
		}
		if info.ModTime().Before(cutoff) {
			os.Remove(filepath.Join(q.dir, e.Name()))
		}
	}
	return nil
}
