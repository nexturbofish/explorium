package cli

import (
	"bufio"
	"fmt"
	"hi/internal/config"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"

	"github.com/spf13/cobra"
	"golang.org/x/term"
)

func CmdInit() *cobra.Command {
	return &cobra.Command{
		Use:   "init",
		Short: "Initialize ~/.hi/ directory structure",
		RunE:  runInit,
	}
}

func runInit(cmd *cobra.Command, args []string) error {
	home, _ := os.UserHomeDir()
	hiDir := filepath.Join(home, ".hi")
	configPath := filepath.Join(hiDir, "config.yaml")

	// 如果已存在，询问是否覆盖
	if _, err := os.Stat(configPath); err == nil {
		fmt.Printf("%s already exists. Overwrite? [y/N] ", configPath)
		var answer string
		fmt.Scanln(&answer)
		answer = strings.TrimSpace(strings.ToLower(answer))
		if answer != "y" && answer != "yes" {
			fmt.Println("Aborted.")
			return nil
		}
	}

	// 创建目录
	for _, d := range []string{
		filepath.Join(hiDir, "sessions"),
		filepath.Join(hiDir, "memories"),
		filepath.Join(hiDir, "skills"),
	} {
		if err := os.MkdirAll(d, 0755); err != nil {
			return fmt.Errorf("create %s: %w", d, err)
		}
	}

	// 默认 mcp.json
	mcpPath := filepath.Join(hiDir, "mcp.json")
	if _, err := os.Stat(mcpPath); os.IsNotExist(err) {
		os.WriteFile(mcpPath, []byte("{}\n"), 0644)
	}

	reader := bufio.NewReader(os.Stdin)

	// 1. Provider 选择
	fmt.Println()
	fmt.Println("Select a provider:")
	fmt.Println("  1) OpenAI (GPT)")
	fmt.Println("  2) Anthropic (Claude)")
	fmt.Print("Provider [1-2] (default 1): ")
	providerChoice, _ := reader.ReadString('\n')
	providerChoice = strings.TrimSpace(providerChoice)

	providerName := "openai"
	defaultBaseURL := "https://api.openai.com"
	defaultModel := "gpt-4o-mini"
	if providerChoice == "2" {
		providerName = "anthropic"
		defaultBaseURL = "https://api.anthropic.com"
		defaultModel = "claude-sonnet-4-20250514"
	}

	// 2. Base URL
	fmt.Printf("Base URL (default %s): ", defaultBaseURL)
	baseURL, _ := reader.ReadString('\n')
	baseURL = strings.TrimSpace(baseURL)
	if baseURL == "" {
		baseURL = defaultBaseURL
	}

	// 3. API key (hidden input)
	fmt.Print("API key (input hidden): ")
	apiKeyBytes, err := term.ReadPassword(int(syscall.Stdin))
	fmt.Println()
	if err != nil {
		return fmt.Errorf("read api key: %w", err)
	}
	apiKey := strings.TrimSpace(string(apiKeyBytes))

	// 4. Model
	fmt.Printf("Model (default %s): ", defaultModel)
	model, _ := reader.ReadString('\n')
	model = strings.TrimSpace(model)
	if model == "" {
		model = defaultModel
	}

	// 5. Max tokens
	fmt.Print("Max output tokens (default 16384): ")
	maxTokensStr, _ := reader.ReadString('\n')
	maxTokensStr = strings.TrimSpace(maxTokensStr)
	maxTokens := 16384
	if maxTokensStr != "" {
		if v, err := strconv.Atoi(maxTokensStr); err == nil && v > 0 {
			maxTokens = v
		}
	}

	// 构建配置
	cfg := config.DefaultConfig()
	cfg.DefaultProvider = providerName
	providerCfg := config.ProviderConfig{
		BaseURL:   baseURL,
		APIKey:    apiKey,
		Model:     model,
		MaxTokens: maxTokens,
	}
	if providerName == "anthropic" {
		cfg.Providers.Anthropic = &providerCfg
	} else {
		cfg.Providers.OpenAI = &providerCfg
	}

	if err := config.Save(configPath, cfg); err != nil {
		return fmt.Errorf("save config: %w", err)
	}

	// chmod 0600 保护 API key
	os.Chmod(configPath, 0600)

	fmt.Printf("\nConfig written to %s\n", configPath)
	fmt.Println("Done. Run `hi ask \"hello\"` to test.")
	return nil
}
