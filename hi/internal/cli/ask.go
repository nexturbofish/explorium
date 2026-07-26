package cli

import (
	"context"
	"fmt"
	"hi/internal/appstate"
	"strings"

	"github.com/spf13/cobra"
)

func CmdAsk() *cobra.Command {
	return &cobra.Command{
		Use:   "ask <prompt>",
		Short: "Single-turn conversation (non-interactive)",
		Args:  cobra.MinimumNArgs(1),
		RunE:  runAsk,
	}
}

func runAsk(cmd *cobra.Command, args []string) error {
	prompt := strings.Join(args, " ")
	a, err := appstate.InitApp(context.Background())
	if err != nil {
		return err
	}
	defer a.Cleanup()

	session := appstate.NewSession(a.Config.DefaultProvider)
	a.Turn.SetSession(session)
	result, err := a.Turn.RunTurn(context.Background(), prompt)
	if err != nil {
		return err
	}
	fmt.Println(result.Response)
	return a.Store.Save(session)
}
