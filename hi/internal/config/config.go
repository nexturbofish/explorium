package config

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/cloudwego/eino-ext/components/model/claude"
	"github.com/cloudwego/eino-ext/components/model/openai"
	"github.com/cloudwego/eino/components/model"
	"gopkg.in/yaml.v3"
)

type Config struct {
	DefaultProvider string            `yaml:"default_provider"`
	Providers       Providers         `yaml:"providers"`
	Reflect         ReflectConfig     `yaml:"reflect"`
	Context         ContextConfig     `yaml:"context"`
	Limits          LimitsConfig      `yaml:"limits"`
	Permissions     PermissionsConfig `yaml:"permissions"`
	RateLimits      RateLimitsConfig  `yaml:"rate_limits"`
	Guard           GuardConfig       `yaml:"guard"`
	Workspace       WorkspaceConfig   `yaml:"workspace"`
	Web             WebConfig         `yaml:"web"`
	UI              UIConfig          `yaml:"ui"`
}

type Providers struct {
	Anthropic *ProviderConfig `yaml:"anthropic"`
	OpenAI    *ProviderConfig `yaml:"openai"`
}

type ProviderConfig struct {
	BaseURL   string `yaml:"base_url"`
	APIKey    string `yaml:"api_key"`
	Model     string `yaml:"model"`
	MaxTokens int    `yaml:"max_tokens"`
}

type ReflectConfig struct {
	MinTurns                int    `yaml:"min_turns"`
	AutoAcceptMemory        bool   `yaml:"auto_accept_memories"`
	AutoAcceptMinConfidence string `yaml:"auto_accept_min_confidences"`
}

type ContextConfig struct {
	ModelLimit      int     `yaml:"model_limit"`
	Headroom        float64 `yaml:"headroom"`
	KeepRecentTurns int     `yaml:"keep_recent_turns"`
}

type LimitsConfig struct {
	MaxToolRounds        int `yaml:"max_tool_rounds"`
	AgentMaxIterations   int `yaml:"agent_max_iterations"`
	ActiveMemoryIndexCap int `yaml:"active_memory_index_cap"`
	ActiveWindowDays     int `yaml:"active_window_days"` // 0 或 nil = 不过期
	SkillIndexCap        int `yaml:"skill_index_cap"`
	RelevantMemoryCap    int `yaml:"relevant_memory_cap"`
	TriggeredSkillCap    int `yaml:"triggered_skill_cap"`
}

type PermissionsConfig struct {
	Allow []string `yaml:"allow"`
	Deny  []string `yaml:"deny"`
}

// RateLimitsConfig defines per-tool rate limits.
type RateLimitsConfig struct {
	GlobalMaxPerMin int `yaml:"global_max_per_min"`
}

// GuardConfig configures policy engine behaviour.
type GuardConfig struct {
	ConfirmTimeout int `yaml:"confirm_timeout"` // seconds
}

type WorkspaceConfig struct {
	Root string `yaml:"root"`
}

type WebConfig struct {
	SearchBackend string `yaml:"serach_backend"`
	TavilyAPIKey  string `yaml:"tavily_api_key"`
	BraveAPIKey   string `yaml:"brave_api_key"`
	CacheTTLSecs  int    `yaml:"cache_ttl_secs"`
}

type UIConfig struct {
	Language string `yaml:"language"`
}

func DefaultConfig() *Config {
	return &Config{
		DefaultProvider: "openai",
		Providers: Providers{
			Anthropic: &ProviderConfig{
				BaseURL:   "https://api.anthropic.com",
				Model:     "claude-sonnet-4-20250514",
				MaxTokens: 8192,
			},
			OpenAI: &ProviderConfig{
				BaseURL:   "https://api.openai.com",
				Model:     "gpt-4o",
				MaxTokens: 8192,
			},
		},
		Context: ContextConfig{
			ModelLimit:      2000000,
			Headroom:        0.18,
			KeepRecentTurns: 10,
		},
		Limits: LimitsConfig{
			MaxToolRounds:      20,
			AgentMaxIterations: 30,
			RelevantMemoryCap:  5,
			TriggeredSkillCap:  3,
		},
		Permissions: PermissionsConfig{
			Allow: []string{"*"},
		},
		RateLimits: RateLimitsConfig{
			GlobalMaxPerMin: 30,
		},
		Guard: GuardConfig{
			ConfirmTimeout: 30,
		},
		UI: UIConfig{Language: "zh-CN"},
	}
}

func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read config: %w", err)
	}
	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("parse config: %w", err)
	}
	return &cfg, nil
}

func LoadDefault() (*Config, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("home dir: %w", err)
	}
	return Load(filepath.Join(home, ".hi", "config.yaml"))
}

func Save(path string, cfg *Config) error {
	data, err := yaml.Marshal(cfg)
	if err != nil {
		return fmt.Errorf("marshal config: %w", err)
	}
	return os.WriteFile(path, data, 0644)
}

func SetupChatModel(ctx context.Context, providerName string, cfg *ProviderConfig) (model.ToolCallingChatModel, error) {
	switch providerName {
	case "anthropic":
		return SetupClaudeChatModel(ctx, cfg)
	case "openai", "deepseek":
		return SetupOpenAIChatModel(ctx, cfg)
	default:
		return nil, fmt.Errorf("unsupported provider: %s", providerName)
	}
}

func SetupClaudeChatModel(ctx context.Context, cfg *ProviderConfig) (model.ToolCallingChatModel, error) {
	chatModel, err := claude.NewChatModel(ctx, &claude.Config{
		APIKey:    cfg.APIKey,
		Model:     cfg.Model,
		BaseURL:   &cfg.BaseURL,
		MaxTokens: cfg.MaxTokens,
	})
	if err != nil {
		return nil, fmt.Errorf("claude: %w", err)
	}
	return chatModel, nil
}

func SetupOpenAIChatModel(ctx context.Context, cfg *ProviderConfig) (model.ToolCallingChatModel, error) {
	chatModel, err := openai.NewChatModel(ctx, &openai.ChatModelConfig{
		APIKey:    cfg.APIKey,
		Model:     cfg.Model,
		BaseURL:   cfg.BaseURL,
		MaxTokens: &cfg.MaxTokens,
	})
	if err != nil {
		return nil, fmt.Errorf("openai: %w", err)
	}
	return chatModel, nil
}

func SetupChatModelFromConfig(ctx context.Context, cfg *Config) (model.ToolCallingChatModel, error) {
	providerCfg := cfg.Providers.Anthropic
	if cfg.DefaultProvider == "openai" || cfg.DefaultProvider == "deepseek" {
		providerCfg = cfg.Providers.OpenAI
	}
	if providerCfg == nil {
		return nil, fmt.Errorf("provider %q not configured", cfg.DefaultProvider)
	}
	return SetupChatModel(ctx, cfg.DefaultProvider, providerCfg)
}
