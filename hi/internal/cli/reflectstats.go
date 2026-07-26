package cli

import (
	"context"
	"fmt"
	"hi/internal/memory"
	"hi/spec"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
)

func CmdReflectStats() *cobra.Command {
	return &cobra.Command{
		Use:   "reflect-stats",
		Short: "Show reflection statistics",
		RunE:  runReflectStats,
	}
}

func runReflectStats(cmd *cobra.Command, args []string) error {
	home, _ := os.UserHomeDir()
	hiDir := filepath.Join(home, ".hi")
	memStore := memory.NewMemoryStore(filepath.Join(hiDir, "memories"), 0)

	deferredDir := filepath.Join(hiDir, "deferred-candidates")
	if _, err := os.Stat(deferredDir); os.IsNotExist(err) {
		fmt.Println("No deferred candidates directory found")
	} else {
		entries, _ := os.ReadDir(deferredDir)
		fmt.Printf("Deferred candidates: %d\n", len(entries))
	}

	pinned, err := memStore.ListPinned(context.Background())
	if err != nil {
		return err
	}
	fmt.Printf("Pinned memories: %d\n", len(pinned))

	all, err := memStore.List(spec.MemoryScopeUser)
	if err != nil {
		return err
	}
	fmt.Printf("Total memories: %d\n", len(all))
	return nil
}
