package cli

import (
	"bufio"
	"context"
	"embed"
	"encoding/json"
	"fmt"
	"io/fs"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"slices"

	"github.com/go-chi/chi"

	"cdr.dev/slog"
	"cdr.dev/slog/sloggers/sloghuman"
	"github.com/coder/preview/types"
	"github.com/coder/preview/web"
	"github.com/coder/serpent"
	"github.com/coder/websocket"
)

//go:embed static/*
var static embed.FS

type responseRecorder struct {
	http.ResponseWriter
	headerWritten bool
	logger        slog.Logger
}

// Implement Hijacker interface for WebSocket support
func (r *responseRecorder) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	if hijacker, ok := r.ResponseWriter.(http.Hijacker); ok {
		return hijacker.Hijack()
	}
	return nil, nil, fmt.Errorf("responseRecorder does not implement http.Hijacker")
}

// Wrap your handler
func debugMiddleware(logger slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			recorder := &responseRecorder{
				ResponseWriter: w,
				logger:         logger,
			}
			next.ServeHTTP(recorder, r)
		})
	}
}

func (r *RootCmd) WebsocketServer() *serpent.Command {
	var (
		address string
	)

	cmd := &serpent.Command{
		Use:   "web",
		Short: "Runs a websocket for interactive form inputs.",
		Options: serpent.OptionSet{
			{
				Name:        "Address",
				Description: "Address to listen on.",
				Required:    false,
				Flag:        "addr",
				Default:     "0.0.0.0:8100",
				Value:       serpent.StringOf(&address),
			},
		},
		// This command is mainly for developing the preview tool.
		Hidden: true,
		Handler: func(i *serpent.Invocation) error {
			ctx := i.Context()
			logger := slog.Make(sloghuman.Sink(i.Stderr)).Leveled(slog.LevelDebug)

			mux := chi.NewMux()

			mux.Use(debugMiddleware(logger))
			
			// Add CORS middleware
			mux.Use(func(next http.Handler) http.Handler {
				return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					w.Header().Set("Access-Control-Allow-Origin", "*")
					w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
					w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
					
					if r.Method == "OPTIONS" {
						w.WriteHeader(http.StatusOK)
						return
					}
					
					next.ServeHTTP(w, r)
				})
			})

			mux.HandleFunc("/users/{dir}", func(rw http.ResponseWriter, r *http.Request) {
				dirFS := os.DirFS(chi.URLParam(r, "dir"))
				availableUsers, err := availableUsers(dirFS)
				if err != nil {
					http.Error(rw, err.Error(), http.StatusInternalServerError)
					return
				}
				_ = json.NewEncoder(rw).Encode(availableUsers)
			})
			mux.HandleFunc("/directories", func(rw http.ResponseWriter, r *http.Request) {
				entries, err := os.ReadDir(".")
				if err != nil {
					http.Error(rw, "Could not read directory", http.StatusInternalServerError)
					return
				}

				var dirs []string
				for _, entry := range entries {
					if entry.IsDir() {
						subentries, err := os.ReadDir(entry.Name())
						if err != nil {
							continue
						}
						if !slices.ContainsFunc(subentries, func(entry fs.DirEntry) bool {
							return filepath.Ext(entry.Name()) == ".tf"
						}) {
							continue
						}
						dirs = append(dirs, entry.Name())
					}
				}
				rw.Header().Set("Content-Type", "application/json")
				rw.WriteHeader(http.StatusOK)
				_ = json.NewEncoder(rw).Encode(dirs)
			})
			mux.HandleFunc("/ws/{dir}", websocketHandler(logger))

			staticDir, err := fs.Sub(static, "static")
			if err != nil {
				return err
			}
			mux.NotFound(http.FileServer(http.FS(staticDir)).ServeHTTP)

			srv := &http.Server{
				Addr:    address,
				Handler: mux,
				BaseContext: func(listener net.Listener) context.Context {
					return ctx
				},
			}

			logger.Info(ctx, "Starting server", slog.F("address", address))
			return srv.ListenAndServe()
		},
	}

	return cmd
}

func websocketHandler(logger slog.Logger) func(rw http.ResponseWriter, r *http.Request) {
	return func(rw http.ResponseWriter, r *http.Request) {

		logger.Debug(r.Context(), "WebSocket connection attempt",
			slog.F("remote_addr", r.RemoteAddr),
			slog.F("path", r.URL.Path),
			slog.F("query", r.URL.RawQuery))

		// Validate all parameters BEFORE upgrading the connection
		dir := chi.URLParam(r, "dir")
		logger.Debug(r.Context(), "Directory parameter", slog.F("dir", dir))

		dinfo, err := os.Stat(dir)
		if err != nil {
			logger.Error(r.Context(), "Directory validation failed",
				slog.Error(err),
				slog.F("dir", dir))
			http.Error(rw, "Could not stat directory: "+err.Error(), http.StatusBadRequest)
			return
		}

		if !dinfo.IsDir() {
			http.Error(rw, "Not a directory", http.StatusBadRequest)
			return
		}

		// Log before WebSocket upgrade
		logger.Debug(r.Context(), "Attempting WebSocket upgrade")

		// Create WebSocket options with proper origin check
		options := &websocket.AcceptOptions{
			OriginPatterns: []string{
				"*",
			},
		}

		conn, err := websocket.Accept(rw, r, options)
		if err != nil {
			logger.Error(r.Context(), "WebSocket upgrade failed", slog.Error(err))
			http.Error(rw, "Could not accept websocket connection: "+err.Error(), http.StatusInternalServerError)
			return
		}
		logger.Debug(r.Context(), "WebSocket connection established")

		var owner types.WorkspaceOwner
		dirFS := os.DirFS(dir)
		planPath := r.URL.Query().Get("plan")
		user := r.URL.Query().Get("user")
		if user != "" {
			available, err := availableUsers(dirFS)
			if err != nil {
				_ = conn.Close(websocket.StatusInternalError, err.Error())
				return
			}

			var ok bool
			owner, ok = available[user]
			if !ok {
				_ = conn.Close(websocket.StatusInternalError, err.Error())
				return
			}
		}

		session := web.NewSession(logger, dirFS, web.SessionInputs{
			PlanPath: planPath,
			User:     owner,
		})
		session.Listen(r.Context(), conn)
	}
}

func availableUsers(dirFS fs.FS) (map[string]types.WorkspaceOwner, error) {
	entries, err := fs.ReadDir(dirFS, ".")
	if err != nil {
		return nil, fmt.Errorf("could not read directory: %w", err)
	}

	idx := slices.IndexFunc(entries, func(entry fs.DirEntry) bool {
		return entry.Name() == "users.json"
	})
	if idx == -1 {
		return map[string]types.WorkspaceOwner{}, nil
	}

	file, err := dirFS.Open(entries[idx].Name())
	if err != nil {
		return nil, fmt.Errorf("could not open users file: %w", err)
	}
	defer file.Close()

	var users map[string]types.WorkspaceOwner
	if err := json.NewDecoder(file).Decode(&users); err != nil {
		return nil, fmt.Errorf("could not decode users file: %w", err)
	}

	return users, nil
}
