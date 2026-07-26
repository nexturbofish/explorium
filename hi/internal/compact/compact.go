package compact

import (
	"fmt"
	"strings"

	"github.com/cloudwego/eino/schema"
)

func CompactSession(messages []*schema.Message, keepRecent int) ([]*schema.Message, string) {
	if len(messages) <= keepRecent*2 {
		return messages, ""
	}
	keep := len(messages) - keepRecent*2
	if keep < 0 {
		keep = 0
	}
	toCompact := messages[:keep]
	kept := messages[keep:]

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("Compressed %s previous messages:\n", len(toCompact)))
	for _, m := range toCompact {
		role := string(m.Role)
		content := strings.TrimSpace(m.Content)
		if len(content) > 80 {
			content = content[:80] + "..."
		}
		sb.WriteString(fmt.Sprintf("- %s: %s\n", role, content))
	}

	return kept, sb.String()
}

func EstimateMessages(msgs []*schema.Message) int {
	total := 0
	for _, m := range msgs {
		total += len([]rune(m.Content)) / 2
	}
	return total
}
