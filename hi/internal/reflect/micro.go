package reflect

import (
	"context"
	"fmt"
	"hi/spec"
	"strings"

	"github.com/cloudwego/eino/components/model"
	"github.com/cloudwego/eino/schema"
)

const reflectInterval = 3

// ShouldMicroReflect 判断当前 turn 是否需要微反思。
// 两个触发条件：
//  1. 用户显式表达了教学意图（偏好/纠正/记忆）
//  2. 距离上次反思已达到 REFLECT_INTERVAL 轮
func ShouldMicroReflect(turnMessages []*schema.Message, turnsSinceLastReflect int) bool {
	if hasExplicitIntent(turnMessages) {
		return true
	}
	return turnsSinceLastReflect >= reflectInterval
}

// hasExplicitIntent 检测用户是否在教学中——说了"记住/以后/偏好"等关键词。
func hasExplicitIntent(msgs []*schema.Message) bool {
	var userText strings.Builder
	for _, m := range msgs {
		if m.Role != schema.User {
			continue
		}
		userText.WriteString(m.Content)
	}
	lower := strings.ToLower(userText.String())
	triggers := []string{
		"记住", "以后", "偏好", "总是", "不是", "不对", "错了",
		"remember", "always", "prefer", "actually", "don't", "wrong",
	}
	for _, t := range triggers {
		if strings.Contains(lower, t) {
			return true
		}
	}
	return false
}

const microReflectSystem = `You are a micro-reflection module. You just observed ONE turn of conversation (user request + assistant response). Decide if anything from this turn is worth persisting as a memory or skill, and whether any existing memory is now stale.

Rules:
- Default to empty arrays. Most turns produce nothing.
- Only propose a memory if the user stated a durable preference, convention, or fact.
- Only propose a skill if the assistant followed a multi-step procedure that would be reusable verbatim next time.
- Never propose more than 1 memory and 1 skill per micro-reflection.
- Confidence should be "low" or "medium" — never "high" for micro-reflection (that's reserved for explicit user requests caught by full reflection).
- If the conversation reveals that an existing memory is WRONG or OUTDATED, produce a memory_candidates entry with the corrected fact and set 'supersedes' to the old memory's id.

Reply with EXACTLY ONE JSON object. No markdown fences. No commentary.

{
  "summary": "<one sentence>",
  "skills": [],
  "memorise": [],
  "conflicts": []
}`

// MicroReflect 运行一次微反思，只看最近一轮对话。
func MicroReflect(ctx context.Context, m model.ToolCallingChatModel, session *spec.Session, recentOutcomes string) (*spec.ReflectionOutput, error) {
	recent := session.Messages
	if len(recent) > 6 {
		recent = recent[len(recent)-6:]
	}

	var sb strings.Builder
	sb.WriteString(microReflectSystem)
	if recentOutcomes != "" {
		sb.WriteString("\n\n=== Recent reflection outcomes ===\n")
		sb.WriteString(recentOutcomes)
	}
	sb.WriteString("\n\nNow produce the reflection JSON.\n")

	msgs := []*schema.Message{
		{Role: schema.System, Content: sb.String()},
	}
	msgs = append(msgs, recent...)

	resp, err := m.Generate(ctx, msgs)
	if err != nil {
		return nil, fmt.Errorf("micro reflect: %w", err)
	}

	var output spec.ReflectionOutput
	if err := robustParse(resp.Content, &output); err != nil {
		return nil, fmt.Errorf("micro reflect: %w", err)
	}
	return &output, nil
}
