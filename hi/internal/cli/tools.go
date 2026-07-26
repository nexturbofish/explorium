package cli

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/spf13/cobra"
)

func CmdTools() *cobra.Command {
	cmd := &cobra.Command{Use: "tools", Short: "View tool call audit logs"}
	cmd.AddCommand(
		cmdToolsLog(),
	)
	return cmd
}

func cmdToolsLog() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "log",
		Short: "Show recent tool call audit log entries",
		Args:  cobra.NoArgs,
		RunE:  runToolsLogCmd,
	}
	cmd.Flags().IntP("n", "n", 10, "number of entries to show")
	return cmd
}

type auditEntry struct {
	Timestamp   time.Time      `json:"timestamp"`
	Tool        string         `json:"tool"`
	Args        map[string]any `json:"args,omitempty"`
	Allowed     bool           `json:"allowed"`
	DenyReason  string         `json:"deny_reason,omitempty"`
	DurationMs  float64        `json:"duration_ms"`
	TokenUsage  int            `json:"token_usage,omitempty"`
	UserConfirm bool           `json:"user_confirm,omitempty"`
}

func runToolsLogCmd(cmd *cobra.Command, args []string) error {
	n, _ := cmd.Flags().GetInt("n")
	s, err := formatAuditLog(n)
	if err != nil {
		return err
	}
	fmt.Print(s)
	return nil
}

func runToolsLog(n int) (string, error) {
	return formatAuditLog(n)
}

func formatAuditLog(n int) (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	logPath := filepath.Join(home, ".hi", "logs", "audit.jsonl")

	data, err := os.ReadFile(logPath)
	if err != nil {
		return "", fmt.Errorf("reading audit log: %w", err)
	}

	var entries []auditEntry
	for _, line := range bytesLines(data) {
		if len(line) == 0 {
			continue
		}
		var e auditEntry
		if err := json.Unmarshal(line, &e); err != nil {
			continue
		}
		entries = append(entries, e)
	}

	start := 0
	if len(entries) > n {
		start = len(entries) - n
	}

	var b strings.Builder
	for i := start; i < len(entries); i++ {
		e := entries[i]
		status := "✓"
		if !e.Allowed {
			status = "✗"
		}
		fmt.Fprintf(&b, "%s  %s  %s", e.Timestamp.Format("15:04:05"), status, e.Tool)
		if len(e.Args) > 0 {
			fmt.Fprintf(&b, "  args=%v", e.Args)
		}
		if e.DenyReason != "" {
			fmt.Fprintf(&b, "  reason=%s", e.DenyReason)
		}
		b.WriteString("\n")
	}
	return b.String(), nil
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
