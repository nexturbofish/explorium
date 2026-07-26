package cli

import (
	"fmt"
	"hi/internal/mcp"

	"github.com/spf13/cobra"
)

func CmdMCP() *cobra.Command {
	cmd := &cobra.Command{Use: "mcp", Short: "Manage MCP servers"}
	cmd.AddCommand(
		&cobra.Command{Use: "list", Short: "List MCP servers", RunE: runMCPList},
		&cobra.Command{Use: "test [name]", Short: "Test MCP server connection", Args: cobra.MaximumNArgs(1), RunE: runMCPTest},
	)
	return cmd
}

func runMCPList(cmd *cobra.Command, args []string) error {
	cfg, err := mcp.LoadMcpConfig()
	if err != nil {
		return err
	}
	for name, spec := range cfg.Servers {
		status := "enabled"
		if spec.Disabled {
			status = "disabled"
		}
		fmt.Printf("%s  %s  %s\n", name, spec.Transport, status)
	}
	return nil
}

func runMCPTest(cmd *cobra.Command, args []string) error {
	cfg, err := mcp.LoadMcpConfig()
	if err != nil {
		return err
	}
	cfg.AutoDetect()
	mgr := mcp.NewMcpManager(cfg)
	defer mgr.StopAll()

	if len(args) > 0 {
		return nil
	}
	return nil
}
