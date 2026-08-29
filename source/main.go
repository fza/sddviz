package main

import (
	"embed"
	"log/slog"
	"os"

	"github.com/wailsapp/wails/v3/pkg/application"

	"github.com/fza/sddviz/internal/graph"
)

//go:embed all:frontend/dist
var assets embed.FS

func main() {
	app := application.New(application.Options{
		Name:        "sddviz",
		Description: "A read-only viewer for SDD decision graphs",
		Services: []application.Service{
			application.NewService(graph.NewService()),
		},
		Server: application.ServerOptions{Host: "localhost", Port: 8080},
		Assets: application.AssetOptions{Handler: application.AssetFileServerFS(assets)},
	})

	err := app.Run()
	if err != nil {
		slog.Error("running the sddviz application", "error", err)
		os.Exit(1)
	}
}
