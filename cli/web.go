package cli

import (
	"context"
	"embed"
	"encoding/json"
	"io/fs"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"slices"

	"github.com/go-chi/chi"

	"cdr.dev/slog"
	"cdr.dev/slog/sloggers/sloghuman"
	"github.com/coder/preview/web"
	"github.com/coder/serpent"
	"github.com/coder/websocket"
)

//go:embed static/*
var static embed.FS

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
		conn, err := websocket.Accept(rw, r, nil)
		if err != nil {
			http.Error(rw, "Could not accept websocket connection", http.StatusInternalServerError)
			return
		}

		dir := chi.URLParam(r, "dir")
		dinfo, err := os.Stat(dir)
		if err != nil {
			_ = conn.Close(websocket.StatusInternalError, "Could not stat directory")
			return
		}

		if !dinfo.IsDir() {
			_ = conn.Close(websocket.StatusInternalError, "Not a directory")
			return
		}

		dirFS := os.DirFS(dir)
		planPath := r.URL.Query().Get("plan")

		session := web.NewSession(logger, dirFS, planPath)
		session.Listen(r.Context(), conn)
	}
}
