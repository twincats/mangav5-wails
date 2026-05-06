package main

import (
	"embed"
	_ "embed"
	"log"
	"mangav5/internal/db"
	"mangav5/internal/menu"
	"mangav5/internal/repo"
	"mangav5/internal/util"
	"mangav5/services"
	"time"

	"github.com/wailsapp/wails/v3/pkg/application"
	"github.com/wailsapp/wails/v3/pkg/events"
)

// Wails uses Go's `embed` package to embed the frontend files into the binary.
// Any files in the frontend/dist folder will be embedded into the binary and
// made available to the frontend.
// See https://pkg.go.dev/embed for more information.

//go:embed all:frontend/dist
var assets embed.FS

func init() {
	// Register a custom event whose associated data type is string.
	// This is not required, but the binding generator will pick up registered events
	// and provide a strongly typed JS/TS API for them.
	application.RegisterEvent[string]("time")
}

// main function serves as the application's entry point. It initializes the application, creates a window,
// and starts a goroutine that emits a time-based event every second. It subsequently runs the application and
// logs any error that might occur.
func main() {

	// Create a new Wails application by providing the necessary options.
	// Variables 'Name' and 'Description' are for application metadata.
	// 'Assets' configures the asset server with the 'FS' variable pointing to the frontend files.
	// 'Bind' is a list of Go struct instances. The frontend has access to the methods of these instances.
	// 'Mac' options tailor the application when running an macOS.

	// Determine database path and migrate if necessary
	dbPath, err := util.GetAndMigrateDatabasePath()
	if err != nil {
		log.Fatal("Failed to setup database path:", err)
	}

	// Initialize Database
	database, err := db.Open(dbPath)
	if err != nil {
		log.Fatal("Failed to open database:", err)
	}
	defer database.Close()

	// Run Migrations
	if err := db.Migrate(database); err != nil {
		log.Fatal("Failed to migrate database:", err)
	}

	// Initialize Repositories
	repos := repo.NewRepositories(database)

	// Initialize Services
	databaseService := services.NewDatabaseService(repos)
	browserService := services.NewBrowserService()
	defer browserService.Cleanup()
	scraperService := services.NewScraperService(browserService)
	fileService := services.NewFileService(databaseService)
	imageService := services.NewImageService()

	app := application.New(application.Options{
		Name:        "mangav5-wails3",
		Description: "Manga Reader, Downloader and overall manga manager",
		Services: []application.Service{
			application.NewService(browserService),
			application.NewService(scraperService),
			application.NewService(services.NewDownloadService()),
			application.NewService(databaseService),
			application.NewService(imageService),
			application.NewServiceWithOptions(fileService, application.ServiceOptions{
				Route: "/filemanga",
			}),
		},
		Assets: application.AssetOptions{
			Handler: application.AssetFileServerFS(assets),
		},
		Mac: application.MacOptions{
			ApplicationShouldTerminateAfterLastWindowClosed: true,
		},
	})

	app.OnShutdown(func() {
		browserService.Cleanup()
	})

	cm := menu.NewContextMenu()
	hcm := cm.GetHomeContextMenu()
	rcm := cm.GetReadContextMenu()
	app.ContextMenu.Add("home-menu", hcm)
	app.ContextMenu.Add("read-menu", rcm)

	// Create a new window with the necessary options.
	// 'Title' is the title of the window.
	// 'Mac' options tailor the window when running on macOS.
	// 'BackgroundColour' is the background colour of the window.
	// 'URL' is the URL that will be loaded into the webview.
	const minWindowWidth = 1200
	const minWindowHeight = 720
	window := app.Window.NewWithOptions(application.WebviewWindowOptions{
		Title:     "Mangav5",
		Width:     minWindowWidth,
		Height:    minWindowHeight,
		MinWidth:  minWindowWidth,
		MinHeight: minWindowHeight,
		Mac: application.MacWindow{
			InvisibleTitleBarHeight: 50,
			Backdrop:                application.MacBackdropTranslucent,
			TitleBar:                application.MacTitleBarHiddenInset,
		},
		BackgroundColour: application.NewRGB(16, 16, 20),
		URL:              "/",
	})

	// WORKAROUND (Windows / Wails v3):
	// Pada beberapa versi Wails v3, setelah window Maximise lalu kembali Restore/UnMaximise/UnMinimise,
	// constraint MinWidth/MinHeight bisa "lepas" dan window dapat di-resize di bawah minimum.
	// Referensi: issue https://github.com/wailsapp/wails/issues/4593
	// Kandidat fix upstream (masih draft saat ditambahkan):
	// - https://github.com/wailsapp/wails/pull/5194
	// - https://github.com/wailsapp/wails/pull/5204
	// Hapus blok listener ini jika kamu sudah upgrade ke versi Wails yang sudah mengandung fix tersebut.
	window.OnWindowEvent(events.Common.WindowUnMaximise, func(_ *application.WindowEvent) {
		window.SetMinSize(minWindowWidth, minWindowHeight)
	})
	window.OnWindowEvent(events.Common.WindowUnMinimise, func(_ *application.WindowEvent) {
		window.SetMinSize(minWindowWidth, minWindowHeight)
	})
	window.OnWindowEvent(events.Common.WindowRestore, func(_ *application.WindowEvent) {
		window.SetMinSize(minWindowWidth, minWindowHeight)
	})

	// Create a goroutine that emits an event containing the current time every second.
	// The frontend can listen to this event and update the UI accordingly.
	go func() {
		for {
			now := time.Now().Format(time.RFC1123)
			app.Event.Emit("time", now)
			time.Sleep(time.Second)
		}
	}()

	// Run the application. This blocks until the application has been exited.
	err = app.Run()

	// Ensure cleanup is called even if OnShutdown was missed
	browserService.Cleanup()

	// If an error occurred while running the application, log it and exit.
	if err != nil {
		log.Fatal(err)
	}
}
