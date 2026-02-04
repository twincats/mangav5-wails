package menu

import (
	"fmt"

	"github.com/wailsapp/wails/v3/pkg/application"
)

type ContextMenu struct {
	// You can store menu item references here if you need to update them later
	// deleteMenuItem *application.MenuItem
}

func NewContextMenu() *ContextMenu {
	return &ContextMenu{}
}

func (f *ContextMenu) GetHomeContextMenu() *application.ContextMenu {
	home_cm := application.NewContextMenu("home-context-menu")

	home_cm.Add("Add Alternative Title").OnClick(func(ctx *application.Context) {
		fmt.Println("Add Alternative Title")
	})

	home_cm.AddSeparator()

	home_cm.Add("Convert Chapter Webp").OnClick(func(ctx *application.Context) {
		fmt.Println("Convert Chapter Webp")
	})

	home_cm.Add("Compress Manga Chapter").OnClick(func(ctx *application.Context) {
		fmt.Println("Compress Manga Chapter")
	})

	home_cm.AddSeparator()

	// ContextMenuData is available in the OnClick handler.
	// Currently, Wails 3 context menus are static definitions. To make the label dynamic
	// (e.g., "Delete [Manga Name]"), you would need to update the menu item before it opens.
	// Since the data is only passed on click/request, a static label like "Delete Manga" is recommended.
	home_cm.Add("Delete Manga").OnClick(func(ctx *application.Context) {
		// Get the data passed from the frontend (e.g. style="--custom-contextmenu-data: {ID}")
		data := ctx.ContextMenuData()
		fmt.Printf("Delete Manga: %s\n", data)
		// Perform delete logic here using 'data'
	})

	return home_cm
}

func (f *ContextMenu) GetReadContextMenu() *application.ContextMenu {
	read_cm := application.NewContextMenu("read-menu")
	read_cm.Add("Fullscreen").OnClick(func(ctx *application.Context) {
		fmt.Println("fullscreen")
	})
	read_cm.Add("2 Pages Mode").OnClick(func(ctx *application.Context) {
		fmt.Println("2 Pages Mode")
	})
	read_cm.Add("LTR").OnClick(func(ctx *application.Context) {
		fmt.Println("LTR")
	})
	read_cm.Add("RTL").OnClick(func(ctx *application.Context) {
		fmt.Println("RTL")
	})
	read_cm.Add("Full Width").OnClick(func(ctx *application.Context) {
		fmt.Println("Full Width")
	})
	read_cm.Add("Normal Width").OnClick(func(ctx *application.Context) {
		fmt.Println("Normal Width")
	})

	read_cm.AddSeparator()
	read_cm.Add("Previous Chapter").OnClick(func(ctx *application.Context) {
		fmt.Println("Previous Chapter")
	})
	read_cm.Add("Home").OnClick(func(ctx *application.Context) {
		fmt.Println("Home")
	})
	read_cm.Add("Next Chapter").OnClick(func(ctx *application.Context) {
		fmt.Println("Next Chapter")
	})

	read_cm.AddSeparator()
	read_cm.Add("Copy Image").OnClick(func(ctx *application.Context) {
		fmt.Println("Copy Image")
	})
	read_cm.Add("Image Link").OnClick(func(ctx *application.Context) {
		fmt.Println("Image Link")
	})
	read_cm.AddSeparator()
	read_cm.Add("Delete Chapter").OnClick(func(ctx *application.Context) {
		fmt.Println("Delete Chapter")
	})
	return read_cm
}
