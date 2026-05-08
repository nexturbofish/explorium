package cmd

import (
	"explorium/codify/internal/server"
	"io"
	"os"

	"github.com/charmbracelet/x/term"
	"github.com/spf13/cobra"
)

func NewCodifyCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "codify",
		Short: "A terminal-first AI assistant for software development",
		Example: `
# Run non-interactively
codify run "Guess my 5 favorite Pokémon"

# Run a non-interactively with pipes and redirection
cat README.md | codify run "make this more glamorous" > GLAMOROUS_README.md

# Run with debug logging in a specific directory
codify --debug --cwd /path/to/project

# Run in yolo mode (auto-accept all permissions; use with care)
codify --yolo

# Run with custom data directory
codify --data-dir /path/to/custom/.codify

# Continue a previous session
codify --session {session-id}

# Continue the most recent session
codify --continue

`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return nil
		},
	}

	cmd.PersistentFlags().StringP("cwd", "c", "", "Current working directory")
	cmd.PersistentFlags().StringP("data-dir", "D", "", "Custom codify data directory")
	cmd.PersistentFlags().StringP("host", "H", server.DefaultHost(), "Connect to codify specified server (for advanced users)")
	cmd.PersistentFlags().BoolP("debug", "d", false, "Debug")

	cmd.Flags().StringP("session", "s", "", "Continue a previous session by ID")
	cmd.Flags().BoolP("continue", "C", false, "Continue the most recent session")
	cmd.Flags().BoolP("yolo", "y", false, "Automatically accepts all permissions (dangerous mode)")
	cmd.MarkFlagsMutuallyExclusive("continue", "session")
	cmd.AddCommand(
		NewRunCommand(),
	)

	return cmd
}

func MaybePrependStdin(prompt string) (string, error) {
	if term.IsTerminal(os.Stdin.Fd()) {
		return prompt, nil
	}

	fi, err := os.Stdin.Stat()
	if err != nil {
		return prompt, err
	}

	if fi.Mode()&os.ModeNamedPipe == 0 && !fi.Mode().IsRegular() {
		return prompt, nil
	}

	bytes, err := io.ReadAll(os.Stdin)
	if err != nil {
		return prompt, err
	}

	return string(bytes) + "\n\n" + prompt, nil
}
