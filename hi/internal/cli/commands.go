package cli

import (
	"fmt"
	"sort"
	"strings"
)

// CmdHandler is a command that can be registered in the chat REPL.
type CmdHandler struct {
	Name        string
	Aliases     []string
	Description string
	Run         func(args []string) (string, error) // returns output text and error
}

// CmdRegistry holds all registered chat commands.
type CmdRegistry struct {
	handlers []CmdHandler
}

func NewCmdRegistry() *CmdRegistry {
	return &CmdRegistry{}
}

func (r *CmdRegistry) Add(h CmdHandler) {
	r.handlers = append(r.handlers, h)
}

func (r *CmdRegistry) Dispatch(input string) (handled bool, output string, err error) {
	if !strings.HasPrefix(input, "/") {
		return false, "", nil
	}

	parts := strings.Fields(input)
	name := strings.TrimPrefix(parts[0], "/")
	args := parts[1:]

	for _, h := range r.handlers {
		if h.Name == name {
			out, err := h.Run(args)
			return true, out, err
		}
		for _, a := range h.Aliases {
			if a == name {
				out, err := h.Run(args)
				return true, out, err
			}
		}
	}

	return true, "", fmt.Errorf("unknown command: /%s", name)
}

// Complete returns all command names (including aliases) matching the given prefix.
func (r *CmdRegistry) Complete(prefix string) []string {
	var matches []string
	seen := map[string]bool{}
	for _, h := range r.handlers {
		names := append([]string{h.Name}, h.Aliases...)
		for _, n := range names {
			if strings.HasPrefix(n, prefix) {
				if !seen[n] {
					matches = append(matches, n)
					seen[n] = true
				}
			}
		}
	}
	sort.Strings(matches)
	return matches
}

func (r *CmdRegistry) Help() string {
	var b strings.Builder
	b.WriteString("Commands:\n")
	for _, h := range r.handlers {
		names := "/" + h.Name
		for _, a := range h.Aliases {
			names += ", /" + a
		}
		fmt.Fprintf(&b, "  %-20s %s\n", names, h.Description)
	}
	return b.String()
}
