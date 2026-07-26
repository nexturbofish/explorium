package cli

import (
	"fmt"
	"hi/internal/config"
	"hi/internal/mcp"
	"os"

	"github.com/spf13/cobra"
)

func CmdDoctor() *cobra.Command {
	return &cobra.Command{
		Use:   "doctor",
		Short: "Run diagnostics",
		RunE:  runDoctor,
	}
}

func runDoctor(cmd *cobra.Command, args []string) error {
	cfg, err := config.LoadDefault()
	if err != nil {
		fmt.Fprintf(os.Stderr, "config: %v\n", err)
	} else {
		model := "unknown"
		switch cfg.DefaultProvider {
		case "anthropic":
			if cfg.Providers.Anthropic != nil {
				model = cfg.Providers.Anthropic.Model
			}
		case "openai", "deepseek":
			if cfg.Providers.OpenAI != nil {
				model = cfg.Providers.OpenAI.Model
			}
		}
		fmt.Printf("config: OK (%s → %s)\n", cfg.DefaultProvider, model)
	}

	home, _ := os.UserHomeDir()
	hiDir := home + "/.hi"

	if _, err := os.Stat(hiDir); err == nil {
		fmt.Println("~/.hi/: OK")
	} else {
		fmt.Println("~/.hi/: missing (run `hi init`)")
	}

	mcpCfg, err := mcp.LoadMcpConfig()
	if err != nil {
		fmt.Fprintf(os.Stderr, "mcp config: %v\n", err)
	} else {
		fmt.Printf("MCP servers: %d configured\n", len(mcpCfg.Servers))
	}

	return nil
}
