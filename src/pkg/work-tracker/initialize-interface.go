package worktracker

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"

	"work-tracker/src/pkg/logger"
)

// initializeInterface sets up the Fyne app/window and constructs the UI widgets.
// It does NOT wire handlers, lay out content, or start tickers.
// Call t.initUI() later to compose these widgets into the window.
func initializeInterface(appId, windowTitle string) *TrackerApp {
	logger.Log(logger.Notice, logger.BoldBlueColor, "%s for '%s'", "Initializing interface", windowTitle)

	// set up the app
	t := &TrackerApp{}
	t.App = app.NewWithID(appId)
	// Apply a slightly larger theme, light theme
	currentTheme := t.App.Settings().Theme()
	t.App.Settings().SetTheme(scaledTheme{
		base:   currentTheme, // or theme.LightTheme()/DarkTheme()
		factor: 1.30,
	})

	// Start large + fullscreen
	t.Window = t.App.NewWindow(windowTitle)
	t.Window.Resize(fyne.NewSize(1280, 720))   // initial size (before FS)
	// t.Window.SetFullScreen(true)               // launch fullscreen

	// title widget
	t.Title = canvas.NewText("Today", theme.Color(theme.ColorNameForeground))
	t.Title.Alignment = fyne.TextAlignCenter
	t.Title.TextStyle = fyne.TextStyle{Bold: true}
	t.Title.TextSize  = theme.TextSize() * 2.0   // 2x normal

	// clock widget
	t.Clock = canvas.NewText("00:00:00", theme.Color(theme.ColorNameForeground))
	t.Clock.Alignment = fyne.TextAlignCenter
	t.Clock.TextStyle = fyne.TextStyle{Monospace: true, Bold: true}
	t.Clock.TextSize  = theme.TextSize() * 3.2   // really big

	// activity bars
	t.AverageActivityBar = NewActivityBar("Average activity")
	t.CurrentActivityBar = NewActivityBar("Current activity")

	// start button
	t.Button = widget.NewButtonWithIcon("Start", theme.MediaPlayIcon(), nil)

	logger.Log(logger.Notice1, logger.BoldGreenColor, "%s for '%s'", "Initialized interface", windowTitle)
	return t
}
