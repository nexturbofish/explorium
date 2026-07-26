package mcp

import (
	"context"
	"fmt"
	"os"
	"sort"
	"sync"

	einomcp "github.com/cloudwego/eino-ext/components/tool/mcp"
	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/compose"
	mcpgo "github.com/mark3labs/mcp-go/client"
	mcpgotypes "github.com/mark3labs/mcp-go/mcp"
)

type McpManager struct {
	sync.Mutex
	cfg     *McpConfig
	clients map[string]*mcpgo.Client
	tools   []tool.BaseTool
}

func NewMcpManager(cfg *McpConfig) *McpManager {
	return &McpManager{
		cfg:     cfg,
		clients: make(map[string]*mcpgo.Client),
	}
}

func (m *McpManager) Start(ctx context.Context) error {
	for name, spec := range m.cfg.Servers {
		if spec.Disabled {
			continue
		}
		var cli *mcpgo.Client
		var err error

		switch spec.Transport {
		case "stdio":
			cli, err = mcpgo.NewStdioMCPClient(spec.Command, envSlice(spec.Env), spec.Args...)
			if err != nil {
				return fmt.Errorf("stdio client %q: %w", name, err)
			}
		case "sse", "http":
			if spec.URL == "" {
				return fmt.Errorf("sse client %q: url is required", name)
			}
			cli, err = mcpgo.NewSSEMCPClient(spec.URL)
			if err != nil {
				return fmt.Errorf("sse client %q: %w", name, err)
			}
			if err := cli.Start(ctx); err != nil {
				return fmt.Errorf("sse client %q start: %w", name, err)
			}

		default:
			return fmt.Errorf("unknown transport %q for server %q", spec.Transport, name)
		}

		// Initialize 握手
		initReq := mcpgotypes.InitializeRequest{}
		initReq.Params.ProtocolVersion = mcpgotypes.LATEST_PROTOCOL_VERSION
		if _, err := cli.Initialize(ctx, initReq); err != nil {
			return fmt.Errorf("init %q: %w", name, err)
		}
		m.clients[name] = cli

		// 获取工具列表
		tools, err := einomcp.GetTools(ctx, &einomcp.Config{Cli: cli})
		if err != nil {
			return fmt.Errorf("get tools %q: %w", name, err)
		}
		m.tools = append(m.tools, tools...)
	}
	return nil
}

func (m *McpManager) ToolsNodeConfig() *compose.ToolsNodeConfig {
	return &compose.ToolsNodeConfig{Tools: m.tools}
}

// ToolMap returns a name→InvokableTool map for use with turn.ToolExecutor.
func (m *McpManager) ToolMap() map[string]tool.InvokableTool {
	m.Lock()
	defer m.Unlock()
	mcp := make(map[string]tool.InvokableTool, len(m.tools))
	for _, t := range m.tools {
		if inv, ok := t.(tool.InvokableTool); ok {
			info, err := t.Info(context.Background())
			if err == nil {
				mcp[info.Name] = inv
			}
		}
	}
	return mcp
}

func (m *McpManager) StopAll() {
	m.Lock()
	defer m.Unlock()
	for name, cli := range m.clients {
		if err := cli.Close(); err != nil {
			fmt.Fprintf(os.Stderr, "error closing MCP client %q: %v\n", name, err)
		} else {
			fmt.Fprintf(os.Stderr, "stopped MCP client %q\n", name)
		}
	}
}

func (m *McpManager) List() []string {
	m.Lock()
	defer m.Unlock()
	names := make([]string, 0, len(m.clients))
	for n := range m.clients {
		names = append(names, n)
	}
	sort.Strings(names)
	return names
}

func (m *McpManager) PingAll(ctx context.Context) error {
	m.Lock()
	defer m.Unlock()
	for name, cli := range m.clients {
		if err := cli.Ping(ctx); err != nil {
			return fmt.Errorf("ping %q: %w", name, err)
		}
	}
	return nil
}

func envSlice(env map[string]string) []string {
	if env == nil {
		return nil
	}
	s := make([]string, 0, len(env))
	for k, v := range env {
		s = append(s, k+"="+v)
	}
	return s
}
