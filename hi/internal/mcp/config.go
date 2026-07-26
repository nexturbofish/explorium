package mcp

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
)

type McpConfig struct {
	Servers map[string]ServerSpec `json:"servers"`
}

type ServerSpec struct {
	Transport string `json:"transport"`
	// Stdio
	Command  string            `json:"command,omitempty"`
	Args     []string          `json:"args,omitempty"`
	Env      map[string]string `json:"env,omitempty"`
	URL      string            `json:"url,omitempty"`
	Disabled bool              `json:"disabled,omitempty"`
}

func (s *ServerSpec) IsDisabled() bool { return s.Disabled }

func LoadMcpConfig() (*McpConfig, error) {
	path := DefaultMcpConfigPath()
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return &McpConfig{Servers: make(map[string]ServerSpec)}, nil
		}
	}
	var cfg McpConfig
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}
	if cfg.Servers == nil {
		cfg.Servers = make(map[string]ServerSpec)
	}
	return &cfg, nil
}

// AutoDetect 自动检测已知 MCP server（如 officecli），如果未配置则添加
func (c *McpConfig) AutoDetect() {
	if _, ok := c.Servers[officecliKey]; !ok {
		if _, err := exec.LookPath(officecliKey); err == nil {
			c.Servers[officecliKey] = ServerSpec{
				Transport: "stdio",
				Command:   "officecli",
				Args:      []string{"mcp"},
			}
		}
	}
}

func DefaultMcpConfigPath() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".hi", "mcp.json")
}

const (
	officecliKey = "officecli"
)
