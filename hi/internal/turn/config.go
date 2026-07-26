package turn

import (
	"context"
	"hi/internal/guard"
	"hi/spec"

	"github.com/cloudwego/eino/components/model"
	"github.com/cloudwego/eino/schema"
)

// ToolExecutor executes a tool call and returns the result string.
type ToolExecutor interface {
	Execute(ctx context.Context, name string, argsJSON string) (string, error)
}

type TurnConfig struct {
	Model            model.ToolCallingChatModel
	Session          *spec.Session
	Store            spec.SessionStore
	Tools            []*schema.ToolInfo
	ToolExecutor     ToolExecutor           // executes tool calls during the tool-use loop
	ContextAssembler *ContextAssembler
	Permissions      *PermissionChecker
	Policy           *guard.PolicyEngine
	OnEvent          func(event TurnEvent)
	MaxToolRounds    int
}

type TurnEvent struct {
	Type string // "system_chunk", "tool_call", "tool_result", "permission_request", "turn_complete"
	Data any
}
