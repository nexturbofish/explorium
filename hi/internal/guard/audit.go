package guard

import (
	"encoding/json"
	"hi/internal/lib"
	"path/filepath"
	"time"
)

type AuditEntry struct {
	Timestamp   time.Time      `json:"timestamp"`
	Tool        string         `json:"tool"`
	Args        map[string]any `json:"args,omitempty"`
	Allowed     bool           `json:"allowed"`
	DenyReason  string         `json:"deny_reason,omitempty"`
	Duration    time.Duration  `json:"duration_ms"`
	TokenUsage  int            `json:"token_usage,omitempty"`
	UserConfirm bool           `json:"user_confirm,omitempty"`
}

type AuditLog struct {
	writer *lib.RotatingFileWriter
}

func NewAuditLog(dir string) (*AuditLog, error) {
	w, err := lib.NewRotatingFileWriter(filepath.Join(dir, "audit.jsonl"), 10*1024*1024)
	if err != nil {
		return nil, err
	}
	return &AuditLog{writer: w}, nil
}

func (a *AuditLog) Log(entry AuditEntry) error {
	// Duration in milliseconds as float
	data, err := json.Marshal(entry)
	if err != nil {
		return err
	}
	return a.writer.WriteLine(data)
}

func (a *AuditLog) Close() error {
	return a.writer.Close()
}

// Query reads the last N audit entries.
func (a *AuditLog) Query(n int) ([]AuditEntry, error) {
	data, err := a.writer.Tail(n)
	if err != nil {
		return nil, err
	}
	var entries []AuditEntry
	for _, line := range data {
		var e AuditEntry
		if err := json.Unmarshal(line, &e); err != nil {
			continue
		}
		entries = append(entries, e)
	}
	return entries, nil
}
