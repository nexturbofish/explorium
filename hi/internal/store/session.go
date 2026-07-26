package store

import (
	"encoding/json"
	"fmt"
	"hi/spec"
	"os"
	"path/filepath"
	"strings"

	"github.com/cloudwego/eino/schema"
)

type SessionStore struct {
	dir string
}

func NewSessionStore(dir string) *SessionStore {
	return &SessionStore{dir: dir}
}

// Save implements [spec.SessionStore].
func (s *SessionStore) Save(session *spec.Session) error {
	path := filepath.Join(s.dir, session.Meta.ID+".jsonl")
	f, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("store: %w", err)
	}
	defer f.Close()

	// 第一行: Meta
	metaLine, _ := json.Marshal(session.Meta)
	fmt.Fprintln(f, string(metaLine))

	// 后续行：每条消息一行 json
	for _, msg := range session.Messages {
		line, _ := json.Marshal(msg)
		fmt.Fprintln(f, string(line))
	}

	// 最后一行：Usage
	usage := map[string]int{
		"total_input_tokens":  session.TotalInputToken,
		"total_output_tokens": session.TotalOutputToken,
	}
	ussageLine, _ := json.Marshal(usage)
	fmt.Fprintln(f, string(ussageLine))

	return nil
}

// Load implements [spec.SessionStore].
func (s *SessionStore) Load(id string) (*spec.Session, error) {
	path := filepath.Join(s.dir, id+".jsonl")
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("store: %w", err)
	}

	var session spec.Session
	lines := strings.Split(string(data), "\n")
	for i, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		if session.Meta.ID == "" {
			if err := json.Unmarshal([]byte(line), &session.Meta); err != nil {
				return nil, fmt.Errorf("store: parse meta on line %d: %w", i, err)
			}
		}
		var msg schema.Message
		if err := json.Unmarshal([]byte(line), &msg); err == nil {
			session.Messages = append(session.Messages, &msg)
			continue
		}

		var usage struct {
			TotalInputTokens  int `json:"total_input_tokens"`
			TotalOutputTokens int `json:"total_output_tokens"`
		}
		if err := json.Unmarshal([]byte(line), &usage); err != nil {
			return nil, fmt.Errorf("store: unparsable line %d: %w", i, err)
		}
		session.TotalInputToken = usage.TotalInputTokens
		session.TotalOutputToken = usage.TotalOutputTokens
	}

	if session.Meta.ID == "" {
		return nil, fmt.Errorf("store: no metadata found in %q", id+".jsonl")
	}

	return &session, nil
}

// List implements [spec.SessionStore].
func (s *SessionStore) List() ([]spec.SessionMeta, error) {
	entries, err := os.ReadDir(s.dir)
	if err != nil {
		return nil, fmt.Errorf("store: %w", err)
	}
	var metas []spec.SessionMeta
	for _, e := range entries {
		if !strings.HasSuffix(e.Name(), ".jsonl") {
			continue
		}
		data, err := os.ReadFile(filepath.Join(s.dir, e.Name()))
		if err != nil {
			continue
		}

		// 第一行是 Meta
		firstLine, _, _ := strings.Cut(string(data), "\n")
		var meta spec.SessionMeta
		if err := json.Unmarshal([]byte(firstLine), &meta); err == nil {
			metas = append(metas, meta)
		}
	}

	return metas, nil
}

// Delete implements [spec.SessionStore].
func (s *SessionStore) Delete(id string) error {
	path := filepath.Join(s.dir, id+".jsonl")
	if err := os.Remove(path); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("store: %w", err)
	}
	return nil
}

var _ spec.SessionStore = (*SessionStore)(nil)
