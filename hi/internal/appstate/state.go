package appstate

import (
	"context"
	"fmt"
	"hi/internal/agent"
	"hi/internal/config"
	"hi/internal/guard"
	"hi/internal/mcp"
	"hi/internal/memory"
	"hi/internal/permission"
	"hi/internal/reflect"
	"hi/internal/skills"
	"hi/internal/store"
	"hi/internal/tools"
	"hi/internal/turn"
	"hi/spec"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/cloudwego/eino/components/model"
	"github.com/google/uuid"
)

// ReflectOptions re-exported for CLI use.
type ReflectOptions = reflect.ReflectOptions

type AppState struct {
	sync.RWMutex
	Config    *config.Config
	Store     *store.SessionStore
	ChatModel model.ToolCallingChatModel
	Turn      *turn.Runner
	Tools     *tools.BuiltinToolSet
	MCP       *mcp.McpManager
	Memory    spec.MemoryStore
	Skills    *skills.SkillStore
	Reflect   *reflect.Reflector
	Agent     *agent.AgentRunner
}

// InitApp initializes a full AppState from default config.
func InitApp(ctx context.Context) (*AppState, error) {
	cfg, err := config.LoadDefault()
	if err != nil {
		return nil, err
	}
	return initAppWithConfig(ctx, cfg)
}

// InitAppWithWorkspace initializes AppState with an optional workspace root override.
func InitAppWithWorkspace(ctx context.Context, workspaceRoot string) (*AppState, error) {
	cfg, err := config.LoadDefault()
	if err != nil {
		return nil, err
	}
	if workspaceRoot != "" {
		cfg.Workspace.Root = workspaceRoot
	}
	return initAppWithConfig(ctx, cfg)
}

func initAppWithConfig(ctx context.Context, cfg *config.Config) (*AppState, error) {

	chatModel, err := config.SetupChatModelFromConfig(ctx, cfg)
	if err != nil {
		return nil, err
	}

	home, _ := os.UserHomeDir()
	hiDir := filepath.Join(home, ".hi")

	sessionStore := store.NewSessionStore(filepath.Join(hiDir, "sessions"))
	memoryStore := memory.NewMemoryStore(filepath.Join(hiDir, "memories"), cfg.Limits.ActiveWindowDays)
	skillStore := skills.NewSkillStore(filepath.Join(hiDir, "skills"))
	skillStore.EnsureBundled(ctx)

	builtinTools := tools.NewBuiltinToolSet(cfg.Workspace.Root).
		WithMemoryStore(memoryStore).
		WithSkillStore(skillStore)

	mcpCfg, err := mcp.LoadMcpConfig()
	if err != nil {
		return nil, err
	}
	mcpCfg.AutoDetect()
	mcpMgr := mcp.NewMcpManager(mcpCfg)
	if err := mcpMgr.Start(ctx); err != nil {
		return nil, err
	}

	ca := &turn.ContextAssembler{
		MemoryStore: memoryStore,
		SkillStore:  skillStore,
		Config:      cfg.Limits,
	}

	// Build combined tool executor (builtin + MCP).
	executor := builtinTools.Executor(mcpMgr.ToolMap())
	allTools, err := executor.ToolInfos(ctx)
	if err != nil {
		return nil, err
	}

	// Initialize guard components: audit log → PolicyEngine.
	logDir := filepath.Join(hiDir, "logs")
	auditLog, err := guard.NewAuditLog(logDir)
	if err != nil {
		return nil, fmt.Errorf("audit log: %w", err)
	}
	permChecker := permission.NewPermissionChecker(cfg.Permissions)
	var rateLimits map[string]guard.RateLimitConfig
	if cfg.RateLimits.GlobalMaxPerMin > 0 {
		rateLimits = map[string]guard.RateLimitConfig{
			"default": {Tokens: cfg.RateLimits.GlobalMaxPerMin, Window: time.Minute},
		}
	}
	rateLimiter := guard.NewRateLimiter(rateLimits)
	confirm := guard.NewConfirmationQueue(time.Duration(cfg.Guard.ConfirmTimeout) * time.Second)
	policyEngine := guard.NewPolicyEngine(permChecker, rateLimiter, confirm, auditLog)

	turnRunner := turn.NewRunner(turn.TurnConfig{
		Model:            chatModel,
		Store:            sessionStore,
		Tools:            allTools,
		ToolExecutor:     executor,
		Policy:           policyEngine,
		ContextAssembler: ca,
		MaxToolRounds:    cfg.Limits.MaxToolRounds,
	})

	agentRunner := agent.NewAgentRunner(chatModel, builtinTools, ca, turn.TurnConfig{}, sessionStore)

	reflector := reflect.NewReflector(reflect.ReflectConfig{
		Model:                chatModel,
		MemoryStore:          memoryStore,
		SkillStore:           skillStore,
		MinTurns:             cfg.Reflect.MinTurns,
		AutoAcceptMemories:   cfg.Reflect.AutoAcceptMemory,
		AutoAcceptConfidence: cfg.Reflect.AutoAcceptMinConfidence,
	})

	return &AppState{
		Config:    cfg,
		Store:     sessionStore,
		ChatModel: chatModel,
		Turn:      turnRunner,
		Tools:     builtinTools,
		MCP:       mcpMgr,
		Memory:    memoryStore,
		Skills:    skillStore,
		Reflect:   reflector,
		Agent:     agentRunner,
		RWMutex:   sync.RWMutex{},
	}, nil
}

// NewSession creates a new session for the given provider.
func NewSession(provider string) *spec.Session {
	return &spec.Session{
		Meta: spec.SessionMeta{
			ID:        uuid.New().String(),
			CreatedAt: time.Now(),
			Model:     provider,
		},
	}
}

// Cleanup stops MCP manager etc.
func (a *AppState) Cleanup() {
	a.MCP.StopAll()
}
