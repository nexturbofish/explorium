package reflect

import (
	"context"
	"fmt"
	"hi/spec"
	"strings"

	"github.com/cloudwego/eino/components/model"
	"github.com/cloudwego/eino/schema"
)

// FocusedReflect 聚焦反思：只针对特定主题分析 session。
func FocusedReflect(ctx context.Context, session *spec.Session, m model.ToolCallingChatModel, topic, recentOutcomes string) (*spec.ReflectionOutput, error) {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf(`You are a focused reflection module. Focus on the topic "%s" in this conversation.

Identify skills, memories, and conflicts relevant to this topic only.

Default to empty arrays. Reply with EXACTLY ONE JSON object. No markdown fences.`, topic))

	if recentOutcomes != "" {
		sb.WriteString("\n\n")
		sb.WriteString(recentOutcomes)
	}

	sb.WriteString("\n\n=== Session transcript ===\n")
	sb.WriteString(formatSession(session))
	sb.WriteString("\n\nProduce the reflection JSON for topic \"")
	sb.WriteString(topic)
	sb.WriteString("\". Default to empty arrays.\n")

	msgs := []*schema.Message{
		{Role: schema.System, Content: sb.String()},
	}

	resp, err := m.Generate(ctx, msgs)
	if err != nil {
		return nil, fmt.Errorf("focused reflect: %w", err)
	}

	var output spec.ReflectionOutput
	if err := robustParse(resp.Content, &output); err != nil {
		return nil, fmt.Errorf("focused reflect: %w", err)
	}
	return &output, nil
}
