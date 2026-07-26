package turn

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"hi/spec"
	"io"

	"github.com/cloudwego/eino/components/model"
	"github.com/cloudwego/eino/schema"
)

const (
	defaultMaxToolRounds = 20
)

type TurnResult struct {
	Response string
	Rounds   int
}

type Runner struct {
	config TurnConfig
}

func NewRunner(config TurnConfig) *Runner {
	if config.MaxToolRounds == 0 {
		config.MaxToolRounds = defaultMaxToolRounds
	}
	return &Runner{config: config}
}

// SetSession sets the session for this runner. Must be called before RunTurn.
func (r *Runner) SetSession(session *spec.Session) {
	r.config.Session = session
}

func (r *Runner) RunTurn(ctx context.Context, userInput string, onEvent ...func(TurnEvent)) (*TurnResult, error) {
	// 1. 构建 system prompt
	if r.config.ContextAssembler != nil {
		ctxData, err := r.config.ContextAssembler.Assemble(ctx, userInput)
		if err != nil {
			return nil, fmt.Errorf("assemble context: %w", err)
		}
		// 移除上一次由 ContextAssembler 注入的 system message
		filtered := make([]*schema.Message, 0, len(r.config.Session.Messages))
		for _, m := range r.config.Session.Messages {
			if m.Role != schema.System {
				filtered = append(filtered, m)
			}
		}
		r.config.Session.Messages = filtered
		r.config.Session.Messages = append(r.config.Session.Messages, &schema.Message{
			Role:    schema.System,
			Content: ctxData.System,
		})
	}

	// 2. 添加用户消息
	r.config.Session.Messages = append(r.config.Session.Messages, &schema.Message{
		Role:    schema.User,
		Content: userInput,
	})

	// 3. 多轮 tool-use 循环
	var result TurnResult
	for round := 0; round < r.config.MaxToolRounds; round++ {
		opts := make([]model.Option, 0, 1)
		if len(r.config.Tools) > 0 {
			opts = append(opts, model.WithTools(r.config.Tools))
		}
		var onEventHandler func(TurnEvent)
		if len(onEvent) > 0 && onEvent[0] != nil {
			onEventHandler = onEvent[0]
		} else if r.config.OnEvent != nil {
			onEventHandler = r.config.OnEvent
		}
		resp, err := r.streamGenerate(ctx, r.config.Session.Messages, onEventHandler, opts...)
		if err != nil {
			return nil, fmt.Errorf("model generate: %w", err)
		}

		r.config.Session.Messages = append(r.config.Session.Messages, resp)

		// 检查 tool calls
		if len(resp.ToolCalls) == 0 {
			result.Response = resp.Content
			break
		}

		for _, tc := range resp.ToolCalls {
			if r.config.OnEvent != nil {
				r.config.OnEvent(TurnEvent{Type: "tool_call", Data: tc})
			}

			// Phase-11: Policy engine evaluation (rate limit + permission + confirm)
			if r.config.Policy != nil {
				args := extractToolArgs(tc)
				pResult := r.config.Policy.Evaluate(ctx, tc.Function.Name, args)
				if !pResult.Allowed {
					r.config.Session.Messages = append(r.config.Session.Messages, &schema.Message{
						Role:       schema.Tool,
						Content:    fmt.Sprintf("tool %q denied by policy: %s", tc.Function.Name, pResult.DenyReason),
						ToolCallID: tc.ID,
						ToolName:   tc.Function.Name,
					})
					if r.config.OnEvent != nil {
						r.config.OnEvent(TurnEvent{Type: "tool_result", Data: "denied"})
					}
					continue
				}
				if pResult.NeedConfirm {
					approved, err := r.config.Policy.Wait(pResult.ConfirmReq)
					if err != nil || !approved {
						r.config.Session.Messages = append(r.config.Session.Messages, &schema.Message{
							Role:       schema.Tool,
							Content:    fmt.Sprintf("tool %q rejected by user", tc.Function.Name),
							ToolCallID: tc.ID,
							ToolName:   tc.Function.Name,
						})
						if r.config.OnEvent != nil {
							r.config.OnEvent(TurnEvent{Type: "tool_result", Data: "rejected"})
						}
						continue
					}
				}
			}

			// 权限检查 (legacy, kept for backward compatibility)
			if r.config.Permissions != nil {
				perm := r.config.Permissions.Check(tc.Function.Name)
				if perm == PermissionDeny {
					r.config.Session.Messages = append(r.config.Session.Messages, &schema.Message{
						Role:       schema.Tool,
						Content:    fmt.Sprintf("tool %q not permitted", tc.Function.Name),
						ToolCallID: tc.ID,
						ToolName:   tc.Function.Name,
					})
					if r.config.OnEvent != nil {
						r.config.OnEvent(TurnEvent{Type: "tool_result", Data: "denied"})
					}
					continue
				}
			}

			// Execute tool call
			if r.config.ToolExecutor != nil {
				result, err := r.config.ToolExecutor.Execute(ctx, tc.Function.Name, tc.Function.Arguments)
				if err != nil {
					result = fmt.Sprintf("error: %v", err)
				}
				r.config.Session.Messages = append(r.config.Session.Messages, &schema.Message{
					Role:       schema.Tool,
					Content:    result,
					ToolCallID: tc.ID,
					ToolName:   tc.Function.Name,
				})
				if r.config.OnEvent != nil {
					r.config.OnEvent(TurnEvent{Type: "tool_result", Data: result})
				}
				// Audit tool execution outcome (success or failure).
				if r.config.Policy != nil {
					r.config.Policy.AuditToolResult(tc.Function.Name, extractToolArgs(tc), err)
				}
			}
		}
	}

	// 4. 保持 session
	if r.config.Store != nil {
		if err := r.config.Store.Save(r.config.Session); err != nil {
			return nil, fmt.Errorf("save session: %w", err)
		}
	}

	if r.config.OnEvent != nil {
		r.config.OnEvent(TurnEvent{Type: "turn_complete", Data: result})
	}

	return &result, nil
}

// streamGenerate calls Stream and assembles the full response, emitting
// stream_chunk events for each fragment.
func (r *Runner) streamGenerate(ctx context.Context, messages []*schema.Message, onEvent func(TurnEvent), opts ...model.Option) (*schema.Message, error) {
	stream, err := r.config.Model.Stream(ctx, messages, opts...)
	if err != nil {
		return nil, err
	}
	defer stream.Close()

	var content string
	var toolCalls []schema.ToolCall
	// Track the last tool call being built across chunks.
	var pendingToolCall *schema.ToolCall

	for {
		chunk, err := stream.Recv()
		if err != nil {
			if errors.Is(err, io.EOF) {
				break
			}
			return nil, err
		}

		content += chunk.Content
		if onEvent != nil && chunk.Content != "" {
			onEvent(TurnEvent{Type: "stream_chunk", Data: chunk.Content})
		}

		// Accumulate tool call fragments.
		for _, tc := range chunk.ToolCalls {
			if tc.Index == nil {
				toolCalls = append(toolCalls, tc)
				continue
			}
			// Streaming tool calls use index to correlate fragments.
			if pendingToolCall != nil && pendingToolCall.Index != nil && *pendingToolCall.Index == *tc.Index {
				pendingToolCall.Function.Name += tc.Function.Name
				pendingToolCall.Function.Arguments += tc.Function.Arguments
			} else {
				if pendingToolCall != nil {
					toolCalls = append(toolCalls, *pendingToolCall)
				}
				tc := tc
				pendingToolCall = &tc
			}
		}
	}
	if pendingToolCall != nil {
		toolCalls = append(toolCalls, *pendingToolCall)
	}

	return &schema.Message{
		Role:      schema.Assistant,
		Content:   content,
		ToolCalls: toolCalls,
	}, nil
}

// extractToolArgs converts tool call arguments JSON into a map for policy evaluation.
func extractToolArgs(tc schema.ToolCall) map[string]any {
	args := make(map[string]any)
	if err := json.Unmarshal([]byte(tc.Function.Arguments), &args); err != nil {
		return nil
	}
	return args
}
