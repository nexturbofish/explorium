package cmd

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/spf13/cobra"
)

func NewRunCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "run [prompt...]",
		Short: "Run a single non-interactive prompt",
		Example: `
# Run a simple prompt
codify run "Guess my 5 favorite Pokémon"

# Pipe input from stdin
curl https://charm.land | codify run "Summarize this website"

# Read from a file
codify run "What is this code doing?" <<< prrr.go

# Redirect output to a file
codify run "Generate a hot README for this project" > MY_HOT_README.md

# Run in quiet mode (hide the spinner)
codify run --quiet "Generate a README for this project"

# Run in verbose mode (show logs)
codify run --verbose "Generate a README for this project"

# Continue a previous session
codify run --session {session-id} "Follow up on your last response"

# Continue the most recent session
codify run --continue "Follow up on your last response"

  `,
		RunE: func(cmd *cobra.Command, args []string) error {

			_, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
			defer cancel()

			prompt := strings.Join(args, " ")

			prompt, err := MaybePrependStdin(prompt)
			if err != nil {
				return fmt.Errorf("failed to read prompt from stdin: %w", err)
			}

			if prompt == "" {
				return fmt.Errorf("no prompt provided")
			}

			fmt.Println(prompt)

			return nil
		},
	}

	return cmd
}
