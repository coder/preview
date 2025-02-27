package cli

import (
	"context"
	"embed"
	"io/fs"
	"net"
	"net/http"
	"os"

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
			logger := slog.Make(sloghuman.Sink(i.Stderr))

			mux := chi.NewMux()
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
