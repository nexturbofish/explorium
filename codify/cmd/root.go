package cmd

import "github.com/spf13/cobra"

func NewCodifyCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use: "codify",
	}

	cmd.AddCommand(
		NewRunCommand(),
	)

	return cmd
}
