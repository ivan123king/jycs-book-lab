package main

import (
	"embed"

	"github.com/wailsapp/wails/v2"
	wailsmenu "github.com/wailsapp/wails/v2/pkg/menu"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
	"github.com/wailsapp/wails/v2/pkg/runtime"
)

//go:embed all:frontend/dist
var assets embed.FS

//go:embed wails.json
var wailsConfig []byte

func main() {
	// Create an instance of the app structure
	app := NewApp()

	// Create application with options
	err := wails.Run(&options.App{
		Title:  "jiayi-book-lab",
		Width:  1024,
		Height: 768,
		AssetServer: &assetserver.Options{
			Assets: assets,
		},
		BackgroundColour: &options.RGBA{R: 27, G: 38, B: 54, A: 1},
		OnStartup:        app.startup,
		Menu:             createApplicationMenu(app),
		Bind: []interface{}{
			app,
		},
	})

	if err != nil {
		println("Error:", err.Error())
	}
}

// createApplicationMenu 创建窗口原生菜单，关于菜单通过事件交给 Vue 弹框渲染。
func createApplicationMenu(app *App) *wailsmenu.Menu {
	applicationMenu := wailsmenu.NewMenu()
	helpMenu := applicationMenu.AddSubmenu("帮助")
	helpMenu.AddText("关于佳依拆书", nil, func(_ *wailsmenu.CallbackData) {
		if app.ctx == nil {
			return
		}
		runtime.EventsEmit(app.ctx, "app:open-about-dialog")
	})
	return applicationMenu
}
