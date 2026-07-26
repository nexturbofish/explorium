package cli

import (
	"context"
	"fmt"
	"hi/internal/appstate"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"encoding/json"
	"hi/internal/config"
	"hi/spec"

	"github.com/spf13/cobra"
)

func startServer(a *appstate.AppState, addr string) error {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /api/v1/health", func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
	})
	mux.HandleFunc("GET /api/v1/sessions", func(w http.ResponseWriter, r *http.Request) {
		metas, err := a.Store.List()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		type summary struct {
			ID    string `json:"id"`
			Model string `json:"model"`
			Time  string `json:"time"`
		}
		var out []summary
		for _, m := range metas {
			out = append(out, summary{ID: m.ID[:8], Model: m.Model, Time: m.CreatedAt.Format(time.RFC3339)})
		}
		json.NewEncoder(w).Encode(out)
	})
	mux.HandleFunc("GET /api/v1/sessions/{id}", func(w http.ResponseWriter, r *http.Request) {
		s, err := a.Store.Load(r.PathValue("id"))
		if err != nil {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}
		json.NewEncoder(w).Encode(s)
	})
	mux.HandleFunc("GET /api/v1/skills", func(w http.ResponseWriter, r *http.Request) {
		list, err := a.Skills.List(r.Context())
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		json.NewEncoder(w).Encode(list)
	})
	mux.HandleFunc("GET /api/v1/memories", func(w http.ResponseWriter, r *http.Request) {
		list, err := a.Memory.List(spec.MemoryScopeUser)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		json.NewEncoder(w).Encode(list)
	})
	mux.HandleFunc("GET /api/v1/config", func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(a.Config)
	})
	mux.HandleFunc("PUT /api/v1/config", func(w http.ResponseWriter, r *http.Request) {
		var cfg config.Config
		if err := json.NewDecoder(r.Body).Decode(&cfg); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		home, _ := os.UserHomeDir()
		if err := config.Save(filepath.Join(home, ".hi", "config.yaml"), &cfg); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		json.NewEncoder(w).Encode(map[string]bool{"applied": true})
	})
	mux.HandleFunc("GET /api/v1/mcp", func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(a.MCP)
	})

	return http.ListenAndServe(addr, mux)
}

func CmdServe() *cobra.Command {
	return &cobra.Command{
		Use:   "serve",
		Short: "Start HTTP/WebSocket server (port 8765)",
		RunE:  runServe,
	}
}

func runServe(cmd *cobra.Command, args []string) error {
	a, err := appstate.InitApp(context.Background())
	if err != nil {
		return err
	}
	defer a.Cleanup()

	fmt.Println("Starting HTTP server on :8765...")
	return startServer(a, ":8765")
}
