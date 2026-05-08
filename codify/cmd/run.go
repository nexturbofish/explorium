package cmd

import "github.com/spf13/cobra"

func NewRunCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use: "run",
	}

	return cmd
}
