package main

import (
	"embed"
	"io/fs"
	"log/slog"
	"net/http"

	"github.com/tiredkangaroo/display64/api"
	"github.com/tiredkangaroo/display64/env"
	"github.com/tiredkangaroo/display64/providers"
)

//go:embed dist
var embedFS embed.FS

func main() {
	// environment
	if err := env.Init(); err != nil {
		slog.Error("initializing environment (fatal)", "error", err)
		return
	}

	// providers
	providers := new(providers.Providers)
	if err := providers.Init(); err != nil {
		slog.Error("initializing providers (fatal)", "error", err)
		return
	}

	// web server
	distFS, err := fs.Sub(embedFS, "dist")
	if err != nil {
		slog.Error("sub filesystem (fatal)", "error", err)
		return
	}

	http.Handle("/api/", http.StripPrefix("/api", api.UseHandler(providers)))
	http.Handle("/", http.FileServerFS(distFS))
	go http.ListenAndServe(":9000", nil)

	if err := providers.Start(providers.NoProvider); err != nil {
		slog.Error("starting no provider (fatal)", "error", err)
	}
	select {}
}
