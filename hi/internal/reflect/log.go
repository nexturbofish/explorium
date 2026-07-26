package reflect

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// ActionTaken 记录用户对反思候选的操作。
type ActionTaken string

const (
	ActionAccept    ActionTaken = "accepted"
	ActionAutoAuto  ActionTaken = "auto_accept"
	ActionReject    ActionTaken = "rejected"
	ActionDefer     ActionTaken = "deferred"
	ActionCancelled ActionTaken = "cancelled"
)

// CandidateKind 候选类型。
type CandidateKind string

const (
	CandidateSkill        CandidateKind = "skill"
	CandidateMemory       CandidateKind = "memory"
	CandidateConflict     CandidateKind = "conflict"
	CandidateOrphanConflict CandidateKind = "orphan_conflict"
)

// ReflectLogEntry 一条反思操作日志。
type ReflectLogEntry struct {
	At        time.Time     `json:"at"`
	SessionID string        `json:"session_id"`
	Kind      CandidateKind `json:"kind"`
	Action    ActionTaken   `json:"action"`
	Label     string        `json:"label"`
}

type reflectLogStore struct {
	path string
}

func newReflectLogStore(homeDir string) *reflectLogStore {
	return &reflectLogStore{
		path: filepath.Join(homeDir, ".hi", "logs", "reflect.jsonl"),
	}
}

// NewReflectLogStore creates a reflect log store rooted at hiDir (the .hi directory).
func NewReflectLogStore(hiDir string) *reflectLogStore {
	return &reflectLogStore{
		path: filepath.Join(hiDir, "logs", "reflect.jsonl"),
	}
}

func (s *reflectLogStore) Append(entry ReflectLogEntry) {
	data, err := json.Marshal(entry)
	if err != nil {
		return
	}
	if err := os.MkdirAll(filepath.Dir(s.path), 0755); err != nil {
		return
	}
	f, err := os.OpenFile(s.path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return
	}
	defer f.Close()
	f.Write(data)
	f.WriteString("\n")
}

// Recent 返回最近的 n 条日志。
func (s *reflectLogStore) Recent(n int) ([]ReflectLogEntry, error) {
	data, err := os.ReadFile(s.path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	var entries []ReflectLogEntry
	for _, line := range bytesLines(data) {
		if len(line) == 0 {
			continue
		}
		var e ReflectLogEntry
		if err := json.Unmarshal(line, &e); err != nil {
			continue
		}
		entries = append(entries, e)
	}
	if len(entries) > n {
		entries = entries[len(entries)-n:]
	}
	return entries, nil
}

// OutcomesSummary 返回最近日志的摘要，用于注入到反思提示词中。
func (s *reflectLogStore) OutcomesSummary(n int) string {
	entries, err := s.Recent(n)
	if err != nil || len(entries) == 0 {
		return ""
	}
	var b fmtBuilder
	b.Line("=== Recent reflection outcomes (learn from these) ===")
	for _, e := range entries {
		b.Linef("- [%s] %s: \"%s\"", e.Action, e.Kind, e.Label)
	}
	return b.String()
}

func bytesLines(data []byte) [][]byte {
	var lines [][]byte
	start := 0
	for i := 0; i < len(data); i++ {
		if data[i] == '\n' {
			lines = append(lines, data[start:i])
			start = i + 1
		}
	}
	if start < len(data) {
		lines = append(lines, data[start:])
	}
	return lines
}

// fmtBuilder 是 strings.Builder 的辅助封装。
type fmtBuilder struct {
	buf []byte
}

func (b *fmtBuilder) Line(s string) {
	b.buf = append(b.buf, s...)
	b.buf = append(b.buf, '\n')
}

func (b *fmtBuilder) Linef(format string, args ...any) {
	b.Line(fmt.Sprintf(format, args...))
}

func (b *fmtBuilder) String() string {
	return string(b.buf)
}
