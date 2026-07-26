package agent

import (
	"context"
	"fmt"
	"hi/internal/compact"
	"hi/spec"
	"hi/internal/turn"
	"strings"
	"time"

	"github.com/cloudwego/eino/schema"
)

type AgentConfig struct {
	Goal              string
	MaxIterations     int // 默认 50
	TurnConfig        turn.TurnConfig
	ContextModelLimit int     // 默认 128000
	ContextHeadroom   float64 // 默认 0.18
	KeepRecentTurns   int     // 默认 4
	ContextAssembler  *turn.ContextAssembler
	SessionStore      spec.SessionStore
}

type AgentEvent struct {
	Type      string          `json:"type"`
	Iteration int             `json:"iteration,omitempty"`
	Max       int             `json:"max,omitempty"`
	TurnEvent *turn.TurnEvent `json:"turn_event,omitempty"`
	Compacted int             `json:"compacted,omitempty"`
	Summary   string          `json:"summary,omitempty"`
	Reason    string          `json:"reason,omitempty"`
}

type AgentOutput struct {
	Messages   []*schema.Message
	TotalUsage *schema.TokenUsage
	Summary    string
}

// RunAgent runs an autonomous agent loop.
//
// Each iteration creates a fresh turn.Runner that handles system prompt assembly,
// tool-use loops, permission checks, and session saving. After each turn, the
// agent checks for [goal_complete] or [goal_failed] markers in the response.
//
// Context compression happens between iterations when the message buffer exceeds
// the configured limit (ContextModelLimit * (1 - ContextHeadroom)).
func RunAgent(ctx context.Context, config AgentConfig) (<-chan AgentEvent, error) {
	if config.MaxIterations == 0 {
		config.MaxIterations = 50
	}
	if config.ContextModelLimit == 0 {
		config.ContextModelLimit = 128000
	}
	if config.KeepRecentTurns == 0 {
		config.KeepRecentTurns = 4
	}

	ch := make(chan AgentEvent, 16)
	go func() {
		defer close(ch)

		session := &spec.Session{
			Meta: spec.SessionMeta{
				ID: fmt.Sprintf("agent-%d", time.Now().Unix()),
			},
			Messages: make([]*schema.Message, 0),
		}

		if config.ContextAssembler != nil {
			ctxData, err := config.ContextAssembler.Assemble(ctx, config.Goal)
			if err != nil {
				ch <- AgentEvent{Type: "error", Reason: err.Error()}
				return
			}
			session.Messages = append(session.Messages, &schema.Message{
				Role:    schema.System,
				Content: ctxData.System,
			})
		}

		session.Messages = append(session.Messages, &schema.Message{
			Role:    schema.System,
			Content: `You are an autonomous agent. You have a goal to achieve.

You have access to tools. Use them to make progress toward your goal.

When you have achieved the goal, include "[goal_complete]" in your response.
If the goal is impossible or you must stop, include "[goal_failed]" in your response.`,
		})

		for i := 0; i < config.MaxIterations; i++ {
			ch <- AgentEvent{Type: "iteration", Iteration: i + 1, Max: config.MaxIterations}

			// 上下文压缩：超过限制时压缩早期消息
			totalTokens := compact.EstimateMessages(session.Messages)
			limit := int(float64(config.ContextModelLimit) * (1 - config.ContextHeadroom))
			if totalTokens > limit {
				compressed, summary := compact.CompactSession(session.Messages, config.KeepRecentTurns)
				ch <- AgentEvent{Type: "context_compacted", Compacted: len(session.Messages) - len(compressed), Summary: summary}
				session.Messages = compressed
				session.Messages = append(session.Messages, &schema.Message{
					Role:    schema.System,
					Content: fmt.Sprintf("[Previous context compressed: %s]", summary),
				})
				session.Messages = append(session.Messages, &schema.Message{
					Role:    schema.User,
					Content: config.Goal,
				})
			}

			// 每次迭代创建一个 turn.Runner，共享同一个 session
			turnCfg := config.TurnConfig
			turnCfg.Session = session
			turnCfg.ContextAssembler = nil // system prompt 已由 agent 管理
			turnRunner := turn.NewRunner(turnCfg)

			// 判断本轮的消息：如果是第一次或有压缩后补充的 goal，用 goal；否则让模型继续
			input := config.Goal
			if i > 0 {
				input = "Continue working toward the goal."
			}

			result, err := turnRunner.RunTurn(ctx, input)
			if err != nil {
				ch <- AgentEvent{Type: "error", Reason: err.Error()}
				return
			}

			ch <- AgentEvent{
				Type:    "turn_complete",
				Summary: result.Response,
			}

			// 检测目标完成/失败
			content := strings.ToLower(result.Response)
			if strings.Contains(content, "[goal_complete]") {
				ch <- AgentEvent{Type: "goal_complete", Summary: result.Response}
				return
			}
			if strings.Contains(content, "[goal_failed]") {
				ch <- AgentEvent{Type: "goal_failed", Summary: result.Response}
				return
			}
		}

		ch <- AgentEvent{Type: "goal_failed", Reason: "max iterations reached"}
	}()

	return ch, nil
}
