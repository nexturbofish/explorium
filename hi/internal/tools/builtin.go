package tools

import (
	"context"
	"fmt"
	"hi/spec"
	"path/filepath"
	"sort"

	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/compose"
	"github.com/cloudwego/eino/schema"
)

type BuiltinToolSet struct {
	workspace   string
	todo        *TodoStore
	memoryStore spec.MemoryStore
	skillStore  spec.SkillStore
	webCtx      *WebToolsContext
	tools       map[string]tool.InvokableTool // name → tool
}

func NewBuiltinToolSet(workspace string) *BuiltinToolSet {
	webCtx := NewWebToolsContext()
	webCtx.WithSearcher(NewDuckDuckGoSearcher())
	s := &BuiltinToolSet{
		workspace: workspace,
		todo:      NewTodoStore(filepath.Join(".hi", "todos.json")),
		webCtx:    webCtx,
		tools:     make(map[string]tool.InvokableTool),
	}
	for _, t := range []tool.InvokableTool{
		&ReadTool{workspace: workspace},
		&WriteTool{workspace: workspace},
		&EditTool{workspace: workspace},
		&BashTool{workspace: workspace},
		&GlobTool{workspace: workspace},
		&GrepTool{workspace: workspace},
		&GitTool{workspace: workspace},
		&TodoListTool{store: s.todo},
		&TodoWriteTool{store: s.todo},
		&ThinkTool{},
		&WebFetchTool{ctx: s.webCtx},
		&WebSearchTool{ctx: s.webCtx},
		&SubagentTool{},
	} {
		info, err := t.Info(context.Background())
		if err != nil {
			panic(fmt.Sprintf("tool %T info: %v", t, err))
		}
		s.tools[info.Name] = t
	}
	return s
}

func (s *BuiltinToolSet) WithMemoryStore(store spec.MemoryStore) *BuiltinToolSet {
	s.memoryStore = store
	if store != nil {
		for _, t := range []tool.InvokableTool{
			&MemorySearchTool{store: store},
			&MemorySaveTool{store: store},
			&MemoryDeleteTool{store: store},
			&PalaceZonesTool{store: store},
			&PalaceReadZoneTool{store: store},
			&PalaceRecallTool{store: store},
		} {
			info, err := t.Info(context.Background())
			if err != nil {
				panic(fmt.Sprintf("tool %T info: %v", t, err))
			}
			s.tools[info.Name] = t
		}
	}
	return s
}

func (s *BuiltinToolSet) WithSkillStore(store spec.SkillStore) *BuiltinToolSet {
	s.skillStore = store
	if store != nil {
		for _, t := range []tool.InvokableTool{
			&SkillListTool{store: store},
			&SkillReadTool{store: store},
			&SkillInstallTool{store: store},
			&SkillDeleteTool{store: store},
			&SkillCreateTool{Store: store},
			&ProposeSkillTool{store: store},
		} {
			info, err := t.Info(context.Background())
			if err != nil {
				panic(fmt.Sprintf("tool %T info: %v", t, err))
			}
			s.tools[info.Name] = t
		}
	}
	return s
}

func (s *BuiltinToolSet) List() []string {
	names := make([]string, 0, len(s.tools))
	for n := range s.tools {
		names = append(names, n)
	}
	sort.Strings(names)
	return names
}

func (s *BuiltinToolSet) Get(name string) (tool.InvokableTool, bool) {
	t, ok := s.tools[name]
	return t, ok
}

// Executor adapts a BuiltinToolSet to the turn.ToolExecutor interface, with
// optional MCP tool fallback — MCP tools are wrapped as InvokableTool via a lookup map.
func (s *BuiltinToolSet) Executor(mcpTools map[string]tool.InvokableTool) *ToolSetExecutor {
	return &ToolSetExecutor{builtin: s, mcp: mcpTools}
}

// --- ToolSetExecutor ---

// ToolSetExecutor combines builtin and MCP tools into a single ToolExecutor.
type ToolSetExecutor struct {
	builtin *BuiltinToolSet
	mcp     map[string]tool.InvokableTool
}

func (e *ToolSetExecutor) Execute(ctx context.Context, name string, argsJSON string) (string, error) {
	if t, ok := e.builtin.Get(name); ok {
		return t.InvokableRun(ctx, argsJSON)
	}
	if t, ok := e.mcp[name]; ok {
		return t.InvokableRun(ctx, argsJSON)
	}
	return "", fmt.Errorf("tool %q not found", name)
}

// Info returns the ToolInfo for all tools (used by turn.TurnConfig.Tools).
func (e *ToolSetExecutor) ToolInfos(ctx context.Context) ([]*schema.ToolInfo, error) {
	var infos []*schema.ToolInfo
	for _, name := range e.builtin.List() {
		t, ok := e.builtin.Get(name)
		if !ok {
			continue
		}
		info, err := t.Info(ctx)
		if err != nil {
			return nil, fmt.Errorf("builtin tool %q info: %w", name, err)
		}
		infos = append(infos, info)
	}
	for name, t := range e.mcp {
		info, err := t.Info(ctx)
		if err != nil {
			return nil, fmt.Errorf("mcp tool %q info: %w", name, err)
		}
		infos = append(infos, info)
	}
	return infos, nil
}

func (s *BuiltinToolSet) AsToolsNodeConfig() *compose.ToolsNodeConfig {
	tools := make([]tool.BaseTool, 0, len(s.tools))
	for _, t := range s.tools {
		tools = append(tools, t)
	}
	return &compose.ToolsNodeConfig{Tools: tools}
}
