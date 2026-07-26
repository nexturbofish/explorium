package cli

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"hi/internal/appstate"
	"hi/internal/reflect"
	"hi/internal/skills"
	"hi/internal/turn"
	"hi/spec"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
	"golang.org/x/term"
)

var errQuit = errors.New("quit")

func CmdChat() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "chat",
		Short: "Interactive REPL session",
		RunE:  runChat,
	}
	cmd.Flags().StringP("workspace", "w", "", "override workspace root directory")
	return cmd
}

func runChat(cmd *cobra.Command, args []string) error {
	workspace, _ := cmd.Flags().GetString("workspace")

	var a *appstate.AppState
	var err error
	if workspace != "" {
		a, err = appstate.InitAppWithWorkspace(context.Background(), workspace)
	} else {
		a, err = appstate.InitApp(context.Background())
	}
	if err != nil {
		return err
	}
	defer a.Cleanup()

	session := appstate.NewSession(a.Config.DefaultProvider)
	a.Turn.SetSession(session)

	cmds := NewCmdRegistry()
	cmds.Add(CmdHandler{
		Name:    "quit",
		Aliases: []string{"exit"},
		Description: "Exit the REPL",
		Run:    func([]string) (string, error) { return "", errQuit },
	})
	cmds.Add(CmdHandler{
		Name:        "help",
		Aliases:     []string{"?"},
		Description: "Show available commands",
		Run: func([]string) (string, error) {
			return cmds.Help(), nil
		},
	})
	cmds.Add(CmdHandler{
		Name:        "tools",
		Aliases:     []string{"audit"},
		Description: "Show recent tool call audit log entries",
		Run: func(args []string) (string, error) {
			n := 10
			if len(args) > 0 {
				fmt.Sscanf(args[0], "%d", &n)
			}
			return runToolsLog(n)
		},
	})
	cmds.Add(CmdHandler{
		Name:        "skills",
		Aliases:     []string{"skill"},
		Description: "Manage skills: list, show <slug>, install <source>, create, delete <slug>",
		Run: func(args []string) (string, error) {
			return dispatchSkillCmd(a, args)
		},
	})
	cmds.Add(CmdHandler{
		Name:        "reflect",
		Aliases:     []string{"reflection", "refl"},
		Description: "Reflection commands: outcomes [n], deferred",
		Run: func(args []string) (string, error) {
			return dispatchReflectCmd(a, args)
		},
	})

	if !term.IsTerminal(int(os.Stdin.Fd())) {
		return pipedMode(a, session, cmds)
	}

	oldState, err := term.MakeRaw(int(os.Stdin.Fd()))
	if err != nil {
		return fmt.Errorf("make raw: %w", err)
	}
	defer term.Restore(int(os.Stdin.Fd()), oldState)

	return rawREPL(a, session, cmds, oldState)
}

func pipedMode(a *appstate.AppState, session *spec.Session, cmds *CmdRegistry) error {
	scanner := bufio.NewScanner(os.Stdin)
	fmt.Print("> ")
	for scanner.Scan() {
		input := scanner.Text()
		if handled, output, err := cmds.Dispatch(input); handled {
			if errors.Is(err, errQuit) {
				break
			}
			if output != "" {
				fmt.Print(output)
			}
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			}
			fmt.Print("> ")
			continue
		}
		var response strings.Builder
		result, err := a.Turn.RunTurn(context.Background(), input, func(evt turn.TurnEvent) {
			if evt.Type == "stream_chunk" {
				if s, ok := evt.Data.(string); ok {
					response.WriteString(s)
					fmt.Print(s)
				}
			}
		})
		if err != nil {
			fmt.Fprintf(os.Stderr, "\nError: %v\n", err)
			fmt.Print("> ")
			continue
		}
		if response.Len() == 0 {
			fmt.Print(result.Response)
		}
		fmt.Println()
		if a.Reflect != nil {
			a.Reflect.Reflect(context.Background(), session, reflect.ReflectOptions{Mode: "micro"})
		}
		fmt.Print("> ")
	}
	return a.Store.Save(session)
}

func rawREPL(a *appstate.AppState, session *spec.Session, cmds *CmdRegistry, oldState *term.State) error {
	reader := bufio.NewReader(os.Stdin)
	line := make([]rune, 0, 256)

	cookedPrint := func(s string) {
		term.Restore(int(os.Stdin.Fd()), oldState)
		fmt.Print(s)
		oldState2, _ := term.MakeRaw(int(os.Stdin.Fd()))
		if oldState2 != nil {
			oldState = oldState2
		}
	}

	submit := func() error {
		fmt.Print("\r\n")
		input := string(line)
		line = line[:0]

		var response strings.Builder
		if handled, output, cmdErr := cmds.Dispatch(input); handled {
			fmt.Print("\r\n")
			if errors.Is(cmdErr, errQuit) {
				return a.Store.Save(session)
			}
			if output != "" {
				s := strings.ReplaceAll(output, "\n", "\r\n")
				fmt.Print(s)
				if !strings.HasSuffix(output, "\n") {
					fmt.Print("\r\n")
				}
			}
			if cmdErr != nil {
				cookedPrint(fmt.Sprintf("Error: %v\r\n", cmdErr))
			}
			fmt.Print("> ")
			return nil
		}

		result, err := a.Turn.RunTurn(context.Background(), input, func(evt turn.TurnEvent) {
			if evt.Type == "stream_chunk" {
				if s, ok := evt.Data.(string); ok {
					response.WriteString(s)
					s = strings.ReplaceAll(s, "\n", "\r\n")
					fmt.Print(s)
				}
			}
		})
		if err != nil {
			cookedPrint(fmt.Sprintf("Error: %v\r\n", err))
		} else {
			if response.Len() == 0 {
				s := strings.ReplaceAll(result.Response, "\n", "\r\n")
				fmt.Print(s)
			}
			fmt.Print("\r\n")
			if a.Reflect != nil {
				a.Reflect.Reflect(context.Background(), session, reflect.ReflectOptions{Mode: "micro"})
			}
		}
		fmt.Print("> ")
		return nil
	}

	fmt.Print("> ")
	for {
		r, _, err := reader.ReadRune()
		if err != nil {
			return fmt.Errorf("read: %w", err)
		}

		switch r {
		case 0x03: // Ctrl+C
			cookedPrint("\n")
			return nil

		case 0x04: // Ctrl+D
			cookedPrint("\n")
			return nil

		case 0x1b: // ESC — Alt+Enter or arrow keys
			peek, err := reader.ReadByte()
			if err != nil {
				break
			}
			if peek == '\r' || peek == '\n' {
				if err := submit(); err != nil {
					return err
				}
				break
			}
			if peek == '[' {
				reader.ReadByte() // consume arrow key letter
				break
			}
			if peek == 0x1b {
				peek2, err := reader.ReadByte()
				if err == nil && (peek2 == '\r' || peek2 == '\n') {
					if err := submit(); err != nil {
						return err
					}
					break
				}
			}

		case '\t': // Tab completion for slash commands
			if len(line) > 0 && line[0] == '/' {
				prefix := string(line[1:])
				matches := cmds.Complete(prefix)
				if len(matches) == 1 {
					// Replace line with completed command + space
					rest := matches[0][len(prefix):]
					for _, r := range rest {
						line = append(line, r)
						fmt.Print(string(r))
					}
					line = append(line, ' ')
					fmt.Print(" ")
				} else if len(matches) > 1 {
					fmt.Print("\r\n")
					for _, m := range matches {
						fmt.Printf("  /%s\r\n", m)
					}
					fmt.Print("> " + string(line))
				}
			}

		case '\r', '\n':
			if len(line) > 0 && line[len(line)-1] == '\n' {
				if err := submit(); err != nil {
					return err
				}
				break
			}
			line = append(line, '\n')
			fmt.Print("\r\n")

		case 0x7f, 0x08: // Backspace
			if len(line) > 0 {
				last := line[len(line)-1]
				line = line[:len(line)-1]
				if last == '\n' {
					fmt.Print("\033[1A\033[K")
				} else {
					w := runeWidth(last)
					for i := 0; i < w; i++ {
						fmt.Print("\b \b")
					}
				}
			}

		default:
			line = append(line, r)
			fmt.Print(string(r))
		}
	}
}

func dispatchSkillCmd(a *appstate.AppState, args []string) (string, error) {
	if len(args) == 0 {
		return "Usage: /skills <subcommand> [args]\n\nSubcommands:\n" +
			"  list                  List installed skills\n" +
			"  show <slug>           Show skill details\n" +
			"  install <source>      Install a skill from GitHub\n" +
			"  create <name>         Create a new local skill\n" +
			"  delete <slug>         Delete a skill\n", nil
	}

	sub := args[0]
	subArgs := args[1:]
	home, _ := os.UserHomeDir()
	sk := skills.NewSkillStore(filepath.Join(home, ".hi", "skills"))

	switch sub {
	case "list", "ls":
		list, err := sk.List(context.Background())
		if err != nil {
			return "", err
		}
		var b strings.Builder
		for _, s := range list {
			always := ""
			if s.Frontmatter.AlwaysActive {
				always = " [always]"
			}
			fmt.Fprintf(&b, "%s  %s%s\n", s.Slug, s.Frontmatter.Description, always)
		}
		return b.String(), nil

	case "show", "view":
		if len(subArgs) < 1 {
			return "", fmt.Errorf("usage: /skills show <slug>")
		}
		skill, err := sk.Get(context.Background(), subArgs[0])
		if err != nil {
			return "", err
		}
		var b strings.Builder
		fmt.Fprintf(&b, "Name: %s\n", skill.Frontmatter.Name)
		fmt.Fprintf(&b, "Slug: %s\n", skill.Slug)
		fmt.Fprintf(&b, "Description: %s\n", skill.Frontmatter.Description)
		fmt.Fprintf(&b, "Triggers: %v\n", skill.Frontmatter.Triggers)
		fmt.Fprintf(&b, "Always Active: %v\n", skill.Frontmatter.AlwaysActive)
		return b.String(), nil

	case "install":
		if len(subArgs) < 1 {
			return "", fmt.Errorf("usage: /skills install <source>")
		}
		ctx := context.Background()
		skill, err := sk.Install(ctx, subArgs[0])
		if err != nil {
			return "", err
		}
		return fmt.Sprintf("Installed %q (%s)\n", skill.Slug, skill.Frontmatter.Description), nil

	case "create":
		if len(subArgs) < 1 {
			return "", fmt.Errorf("usage: /skills create <name>")
		}
		name := subArgs[0]
		opts := skills.CreateSkillOpts{
			Name:        name,
			Description: "TODO: add description",
			Body:        "# " + name + "\n\nTODO: add instructions.\n",
		}
		if err := skills.CreateSkill(context.Background(), filepath.Join(home, ".hi", "skills"), opts, false); err != nil {
			return "", err
		}
		return fmt.Sprintf("Created skill %q. Edit it at ~/.hi/skills/%s/SKILL.md\n", name, name), nil

	case "delete", "rm":
		if len(subArgs) < 1 {
			return "", fmt.Errorf("usage: /skills delete <slug>")
		}
		for _, b := range skills.BundledSkills {
			if strings.EqualFold(subArgs[0], b) {
				return "", fmt.Errorf("cannot delete bundled skill %q", subArgs[0])
			}
		}
		if err := sk.Delete(context.Background(), subArgs[0]); err != nil {
			return "", err
		}
		return fmt.Sprintf("Deleted skill %q\n", subArgs[0]), nil

	default:
		return "", fmt.Errorf("unknown skill subcommand: %s (try: list, show, install, create, delete)", sub)
	}
}

func dispatchReflectCmd(a *appstate.AppState, args []string) (string, error) {
	if len(args) == 0 {
		return "Usage: /reflect <subcommand> [args]\n\nSubcommands:\n" +
			"  outcomes [n]    Show last n reflection log entries\n" +
			"  deferred        List deferred candidates\n", nil
	}

	switch args[0] {
	case "outcomes", "log", "recent":
		n := 10
		if len(args) > 1 {
			fmt.Sscanf(args[1], "%d", &n)
		}
		home, _ := os.UserHomeDir()
		store := reflect.NewReflectLogStore(filepath.Join(home, ".hi"))
		entries, err := store.Recent(n)
		if err != nil {
			return "", err
		}
		if len(entries) == 0 {
			return "No reflection outcomes recorded yet.\n", nil
		}
		var b strings.Builder
		for _, e := range entries {
			fmt.Fprintf(&b, "[%s] %s %s: %s\n", e.At.Format("01-02 15:04"), e.Action, e.Kind, e.Label)
		}
		return b.String(), nil

	case "deferred", "pending":
		home, _ := os.UserHomeDir()
		deferredDir := filepath.Join(home, ".hi", "deferred-candidates")
		if _, err := os.Stat(deferredDir); os.IsNotExist(err) {
			return "No deferred candidates.\n", nil
		}
		entries, err := os.ReadDir(deferredDir)
		if err != nil {
			return "", err
		}
		if len(entries) == 0 {
			return "No deferred candidates.\n", nil
		}
		var b strings.Builder
		fmt.Fprintf(&b, "%d deferred candidates:\n", len(entries))
		for _, e := range entries {
			fmt.Fprintf(&b, "  %s\n", e.Name())
		}
		return b.String(), nil

	default:
		return "", fmt.Errorf("unknown reflect subcommand: %s (try: outcomes, deferred)", args[0])
	}
}

func runeWidth(r rune) int {
	if r >= 0x1100 &&
		(r <= 0x115F || r == 0x2329 || r == 0x232A ||
			(0x2E80 <= r && r <= 0xA4CF && r != 0x303F) ||
			(0xAC00 <= r && r <= 0xD7AF) ||
			(0xF900 <= r && r <= 0xFAFF) ||
			(0xFE10 <= r && r <= 0xFE19) ||
			(0xFE30 <= r && r <= 0xFE6F) ||
			(0xFF01 <= r && r <= 0xFF60) ||
			(0xFFE0 <= r && r <= 0xFFE6) ||
			(0x1B000 <= r && r <= 0x1B0FF) ||
			(0x1D167 <= r && r <= 0x1D169) ||
			(0x20000 <= r && r <= 0x2FFFF)) {
		return 2
	}
	return 1
}
