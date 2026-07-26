package reflect

import (
	"context"
	"fmt"
	"hi/spec"
	"strings"

	"github.com/cloudwego/eino/components/model"
	"github.com/cloudwego/eino/schema"
)

const fullReflectSystem = `You are a reflection module for a self-evolving agent. After each completed session you are given the full transcript plus the agent's current skill / memory inventory. Your job is to identify three things, and only when each case truly meets the bar:

1. SKILL CANDIDATES — reusable procedures the agent worked out, with clear triggers and a self-contained body of instructions in markdown. Only propose if you would genuinely want the same procedure applied next time the same situation appears. Skip anything that was a one-shot exploration.

2. MEMORY CANDIDATES — durable facts, conventions, preferences, or constraints the agent discovered that should persist across sessions. One claim per memory. Default 'zone' to 'general'.

3. CONFLICTS — when a new memory candidate contradicts, duplicates, or subsumes an existing memory, report a conflict referencing the existing memory id and proposing resolution.

CRITICAL — the agent's user sees every candidate and must decide. Spammy proposals erode trust. Default to empty arrays. Prefer false negatives over false positives. Confidence = "high" should be rare.

Reply with EXACTLY ONE JSON object. No markdown fences. No commentary.

{
  "summary": "<one sentence summarising what the session accomplished>",
  "skills": [
    {
      "name": "kebab-case-name",
      "description": "one-line description",
      "triggers": ["keyword", "phrase"],
      "body": "## Title\n\nFull markdown instructions.",
      "rationale": "why this is reusable",
      "confidence": "low" | "medium" | "high"
    }
  ],
  "memorise": [
    {
      "content": "one short statement; one fact per memory",
      "tags": ["tag"],
      "zone": "general",
      "confidence": "low" | "medium" | "high",
      "rationale": "why this should persist",
      "supersedes": ["mem_xxx"]
    }
  ],
  "conflicts": [
    {
      "existing_id": "mem_xxx",
      "new_content": "the new fact",
      "explanation": "what the disagreement is",
      "resolution": "keep_new | keep_old | merge"
    }
  ]
}
`

// FullReflect 完整反思：用整个 session 生成反思输出。
func FullReflect(ctx context.Context, session *spec.Session, m model.ToolCallingChatModel, recentOutcomes string) (*spec.ReflectionOutput, error) {
	var sb strings.Builder
	sb.WriteString(fullReflectSystem)

	modules := []struct {
		title string
		body  string
	}{
		{"Recent reflection outcomes", recentOutcomes},
		{"Session transcript", formatSession(session)},
	}
	for _, mod := range modules {
		if mod.body == "" {
			continue
		}
		sb.WriteString("\n\n=== ")
		sb.WriteString(mod.title)
		sb.WriteString(" ===\n")
		sb.WriteString(mod.body)
	}

	sb.WriteString("\n\nNow produce the reflection JSON. Default to empty arrays.\n")

	msgs := []*schema.Message{
		{Role: schema.System, Content: sb.String()},
	}

	resp, err := m.Generate(ctx, msgs)
	if err != nil {
		return nil, fmt.Errorf("full reflect: %w", err)
	}

	var output spec.ReflectionOutput
	if err := robustParse(resp.Content, &output); err != nil {
		return nil, fmt.Errorf("full reflect: %w", err)
	}
	return &output, nil
}

// formatSession 将 session 消息格式化为文本用于提示词。
func formatSession(session *spec.Session) string {
	var b fmtBuilder
	for _, m := range session.Messages {
		switch m.Role {
		case schema.User:
			b.Linef("[User] %s", truncate(m.Content, 1000))
		case schema.Assistant:
			if m.Content != "" {
				b.Linef("[Assistant] %s", truncate(m.Content, 1000))
			}
			for _, tc := range m.ToolCalls {
				b.Linef("[Assistant tool_use] %s(%s)", tc.Function.Name, truncate(string(tc.Function.Arguments), 200))
			}
		case schema.Tool:
			b.Linef("[Tool] %s: %s", m.ToolName, truncate(m.Content, 400))
		case schema.System:
			// skip system messages in transcript
		}
	}
	return b.String()
}

func truncate(s string, max int) string {
	runes := []rune(s)
	if len(runes) <= max {
		return s
	}
	return string(runes[:max]) + "…"
}
