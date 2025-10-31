package worktracker

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"

	"work-tracker/src/pkg/logger"
)

// initializeInterface sets up the Fyne app/window and constructs the UI widgets.
// It does NOT wire handlers, lay out content, or start tickers.
// Call t.initUI() later to compose these widgets into the window.
func initializeInterface(appId, windowTitle string) *TrackerApp {
	logger.Log(logger.Notice, logger.BoldBlueColor, "%s for '%s'", "Initializing interface", windowTitle)

	// set up the app and window
	t := &TrackerApp{}
	t.App = app.NewWithID(appId)
	t.Window = t.App.NewWindow(windowTitle)
	t.Window.Resize(fyne.NewSize(420, 220))

	// title widget
	t.Title = widget.NewLabel("Today")
	t.Title.Alignment = fyne.TextAlignCenter
	t.Title.TextStyle = fyne.TextStyle{Bold: true}

	// clock widget
	t.Clock = widget.NewLabel("00:00:00")
	t.Clock.Alignment = fyne.TextAlignCenter
	t.Clock.TextStyle = fyne.TextStyle{Monospace: true, Bold: true}
	t.Clock.Wrapping = fyne.TextWrapOff
	t.Clock.Importance = widget.MediumImportance

	// status widget
	t.Status = widget.NewLabel("stopped")
	t.Status.Alignment = fyne.TextAlignCenter

	// start button
	t.Button = widget.NewButtonWithIcon("Start", theme.MediaPlayIcon(), nil)

	logger.Log(logger.Notice1, logger.BoldGreenColor, "%s for '%s'", "Initialized interface", windowTitle)
	return t
}
