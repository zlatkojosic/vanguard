package main

import (
	"embed"
	"log"
	"os"

	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
	"github.com/wailsapp/wails/v2/pkg/options/windows"
	"nts/vanguard/internal/upgrade"
)

//go:embed all:frontend/dist
var assets embed.FS

func main() {
	if upgrade.MaybeApply(os.Args[1:]) {
		return
	}
	if c := upgrade.OpenAppLog(); c != nil {
		defer c.Close()
	}
	app := NewApp()
	err := wails.Run(&options.App{
		Title:         "N POS Z",
		Width:         480,
		Height:        560,
		DisableResize: true,
		Frameless:     false,
		AssetServer: &assetserver.Options{
			Assets: assets,
		},
		BackgroundColour: &options.RGBA{R: 51, G: 51, B: 51, A: 255},
		OnStartup:        app.startup,
		OnShutdown:       app.shutdown,
		Bind:             []interface{}{app},
		Windows: &windows.Options{
			WebviewIsTransparent: false,
			WindowIsTranslucent:  false,
		},
	})
	if err != nil {
		log.Fatal(err)
	}
}
