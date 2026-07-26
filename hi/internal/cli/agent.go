package cli

import (
	"context"
	"fmt"
	"hi/internal/appstate"
	"strings"

	"github.com/spf13/cobra"
)

func CmdAgentRun() *cobra.Command {
	return &cobra.Command{
		Use:   "run <goal>",
		Short: "Run agent autonomously to achieve a goal",
		Args:  cobra.MinimumNArgs(1),
		RunE:  runAgentRun,
	}
}

func runAgentRun(cmd *cobra.Command, args []string) error {
	goal := strings.Join(args, " ")
	a, err := appstate.InitApp(context.Background())
	if err != nil {
		return err
	}
	defer a.Cleanup()

	events, err := a.Agent.Run(context.Background(), goal)
	if err != nil {
		return err
	}
	for evt := range events {
		switch evt.Type {
		case "iteration":
			fmt.Printf("[%d/%d] ", evt.Iteration, evt.Max)
		case "turn_complete":
			fmt.Println(evt.Summary)
		case "goal_complete":
			fmt.Printf("\nGoal complete: %s\n", evt.Summary)
		case "goal_failed":
			fmt.Printf("\nGoal failed: %s\n", evt.Reason)
			return fmt.Errorf("goal failed: %s", evt.Reason)
		case "error":
			return fmt.Errorf("agent error: %s", evt.Reason)
		}
	}
	return nil
}
