package agent

import (
	"context"
	"hi/spec"
	"hi/internal/tools"
	"hi/internal/turn"

	"github.com/cloudwego/eino/components/model"
)

type AgentRunner struct {
	chatModel        model.ToolCallingChatModel
	tools            *tools.BuiltinToolSet
	contextAssembler *turn.ContextAssembler
	turnConfig       turn.TurnConfig
	sessionStore     spec.SessionStore
}

func NewAgentRunner(
	chatModel model.ToolCallingChatModel,
	tools *tools.BuiltinToolSet,
	ca *turn.ContextAssembler,
	cfg turn.TurnConfig,
	store spec.SessionStore,
) *AgentRunner {
	return &AgentRunner{
		chatModel:        chatModel,
		tools:            tools,
		contextAssembler: ca,
		turnConfig:       cfg,
		sessionStore:     store,
	}
}

func (r *AgentRunner) Run(ctx context.Context, goal string) (<-chan AgentEvent, error) {
	return RunAgent(ctx, AgentConfig{
		Goal:             goal,
		TurnConfig:       r.turnConfig,
		ContextAssembler: r.contextAssembler,
		SessionStore:     r.sessionStore,
	})
}
