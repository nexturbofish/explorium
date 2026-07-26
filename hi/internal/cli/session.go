package cli

import (
	"fmt"
	"hi/internal/store"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
)

func CmdSession() *cobra.Command {
	cmd := &cobra.Command{Use: "session", Short: "Manage sessions"}
	cmd.AddCommand(
		&cobra.Command{Use: "list", Short: "List all sessions", RunE: runSessionList},
		&cobra.Command{Use: "show <id>", Short: "Show session details", Args: cobra.ExactArgs(1), RunE: runSessionShow},
	)
	return cmd
}

func runSessionList(cmd *cobra.Command, args []string) error {
	home, _ := os.UserHomeDir()
	s := store.NewSessionStore(filepath.Join(home, ".hi", "sessions"))
	metas, err := s.List()
	if err != nil {
		return err
	}
	for _, m := range metas {
		fmt.Printf("%s  %s  %s\n", m.ID[:8], m.CreatedAt.Format("2006-01-02 15:04"), m.Model)
	}
	return nil
}

func runSessionShow(cmd *cobra.Command, args []string) error {
	home, _ := os.UserHomeDir()
	s := store.NewSessionStore(filepath.Join(home, ".hi", "sessions"))
	session, err := s.Load(args[0])
	if err != nil {
		return err
	}
	fmt.Printf("ID: %s\n", session.Meta.ID)
	fmt.Printf("Model: %s\n", session.Meta.Model)
	fmt.Printf("Messages: %d\n", len(session.Messages))
	return nil
}
