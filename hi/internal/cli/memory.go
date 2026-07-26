package cli

import (
	"fmt"
	"hi/internal/memory"
	"hi/spec"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
)

func CmdMemory() *cobra.Command {
	cmd := &cobra.Command{Use: "memory", Short: "Manage memories", Aliases: []string{"memories"}}
	cmd.AddCommand(
		&cobra.Command{Use: "list", Short: "List memories", RunE: runMemoryList},
		&cobra.Command{Use: "show <id>", Short: "Show memory details", Args: cobra.ExactArgs(1), RunE: runMemoryShow},
		&cobra.Command{Use: "delete <id>", Short: "Delete a memory", Args: cobra.ExactArgs(1), RunE: runMemoryDelete},
		&cobra.Command{Use: "pin <id>", Short: "Pin a memory", Args: cobra.ExactArgs(1), RunE: runMemoryPin},
		&cobra.Command{Use: "unpin <id>", Short: "Unpin a memory", Args: cobra.ExactArgs(1), RunE: runMemoryUnpin},
	)
	return cmd
}

func runMemoryList(cmd *cobra.Command, args []string) error {
	home, _ := os.UserHomeDir()
	mem := memory.NewMemoryStore(filepath.Join(home, ".hi", "memories"), 0)
	list, err := mem.List(spec.MemoryScopeUser)
	if err != nil {
		return err
	}
	for _, m := range list {
		pin := " "
		if m.Frontmatter.Pinned {
			pin = "*"
		}
		body := strings.ReplaceAll(m.Content, "\n", " ")
		if len(body) > 60 {
			body = body[:60] + "..."
		}
		fmt.Printf("%s %s  %s\n", pin, m.Frontmatter.ID[:8], body)
	}
	return nil
}

func runMemoryShow(cmd *cobra.Command, args []string) error {
	home, _ := os.UserHomeDir()
	mem := memory.NewMemoryStore(filepath.Join(home, ".hi", "memories"), 0)
	m, err := mem.Load(args[0])
	if err != nil {
		return err
	}
	fmt.Printf("ID: %s\n", m.Frontmatter.ID)
	fmt.Printf("Content:\n%s\n", m.Content)
	return nil
}

func runMemoryDelete(cmd *cobra.Command, args []string) error {
	home, _ := os.UserHomeDir()
	mem := memory.NewMemoryStore(filepath.Join(home, ".hi", "memories"), 0)
	return mem.Delete(args[0])
}

func runMemoryPin(cmd *cobra.Command, args []string) error {
	home, _ := os.UserHomeDir()
	mem := memory.NewMemoryStore(filepath.Join(home, ".hi", "memories"), 0)
	return mem.Pin(args[0])
}

func runMemoryUnpin(cmd *cobra.Command, args []string) error {
	home, _ := os.UserHomeDir()
	mem := memory.NewMemoryStore(filepath.Join(home, ".hi", "memories"), 0)
	return mem.Unpin(args[0])
}
