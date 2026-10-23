package main

import (
	"embed"
	"flag"
	"log/slog"
	"os"

	"github.com/wailsapp/wails/v3/pkg/application"

	"github.com/fza/sddviz/internal/graph"
)

//go:embed all:frontend/dist
var assets embed.FS

func main() {
	flag.Parse()

	// The first argument names a repository to open at once, which is the only
	// way in where no window system offers a directory picker.
	app := application.New(application.Options{
		Name:        "sddviz",
		Description: "A read-only viewer for SDD decision graphs",
		Services: []application.Service{
			application.NewService(graph.NewService(flag.Arg(0), graph.UserConfigPath())),
		},
		Server: application.ServerOptions{Host: "localhost", Port: 8080},
		Assets: application.AssetOptions{Handler: application.AssetFileServerFS(assets)},
		Mac: application.MacOptions{
			ApplicationShouldTerminateAfterLastWindowClosed: true,
		},
	})

	app.Window.NewWithOptions(application.WebviewWindowOptions{
		Title:            "sddviz",
		Width:            1100,
		Height:           720,
		MinWidth:         720,
		MinHeight:        480,
		BackgroundColour: application.NewRGB(24, 26, 33),
		URL:              "/",
		Mac: application.MacWindow{
			TitleBar:                application.MacTitleBarHiddenInset,
			InvisibleTitleBarHeight: 52,
			// Refusing the full-screen mode leaves the window's zoom control doing
			// what it does on a window that cannot take one: growing to the
			// screen it sits on and staying a window, rather than taking a space
			// of its own. A reader keeps every other window beside this one.
			CollectionBehavior: application.MacWindowCollectionBehaviorFullScreenNone,
		},
	})

	err := app.Run()
	if err != nil {
		slog.Error("running the sddviz application", "error", err)
		os.Exit(1)
	}
}
