package app

import (
	"hi/internal/cli"

	"github.com/spf13/cobra"
)

// Run creates the root command and executes it.
func Run() error {
	root := &cobra.Command{
		Use:   "hi",
		Short: "Hi-inspired AI agent",
	}
	root.AddCommand(
		cli.CmdInit(),
		cli.CmdAsk(),
		cli.CmdChat(),
		cli.CmdAgentRun(),
		cli.CmdSession(),
		cli.CmdMCP(),
		cli.CmdSkills(),
		cli.CmdMemory(),
		cli.CmdReflectStats(),
		cli.CmdServe(),
		cli.CmdDoctor(),
		cli.CmdTools(),
	)
	return root.Execute()
}
