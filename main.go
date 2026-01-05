package main

import (
	"embed"
	_ "embed"
	"grpc-gui/internal/consts"
	"grpc-gui/internal/utils"
	"log"
	"path/filepath"

	"github.com/wailsapp/wails/v3/pkg/application"
)

//go:embed all:frontend/dist
var assets embed.FS

func init() {
	// Register a custom event whose associated data type is string.
	// This is not required, but the binding generator will pick up registered events
	// and provide a strongly typed JS/TS API for them.
	application.RegisterEvent[string]("time")
}

func main() {

	appConfigDir, err := utils.GetAppConfigDir()
	if err != nil {
		log.Fatalf("failed to get app config dir: %v", err)
	}

	dbPath := filepath.Join(appConfigDir, consts.AppDbName)
	appService := NewApp(dbPath)

	app := application.New(application.Options{
		Name:        "grpc-gui",
		Description: "gRPC GUI",
		Services: []application.Service{
			application.NewService(appService),
		},
		Assets: application.AssetOptions{
			Handler: application.AssetFileServerFS(assets),
		},
		Mac: application.MacOptions{
			ApplicationShouldTerminateAfterLastWindowClosed: true,
		},
	})

	app.Window.NewWithOptions(application.WebviewWindowOptions{
		Title: "grpc-gui",
		Mac: application.MacWindow{
			InvisibleTitleBarHeight: 50,
			Backdrop:                application.MacBackdropTranslucent,
			TitleBar:                application.MacTitleBarHiddenInset,
		},
		BackgroundColour: application.NewRGB(27, 38, 54),
		URL:              "/",
		Frameless:        true,
		Width:            1244,
		Height:           700,
	})

	err = app.Run()
	if err != nil {
		log.Fatal(err)
	}
}
