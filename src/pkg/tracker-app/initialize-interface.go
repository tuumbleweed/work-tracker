package trackerapp

import (
	"os"
	"path/filepath"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/driver/desktop"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"

	tl "github.com/tuumbleweed/tintlog/logger"
	"github.com/tuumbleweed/tintlog/palette"
	"github.com/tuumbleweed/xerr"
)

// initializeInterface sets up the Fyne app/window and constructs the UI widgets.
// It does NOT wire handlers, lay out content, or start tickers.
// Call t.initUI() later to compose these widgets into the window.
func initializeInterface(appId, windowTitle, tasksFilePath string) (t *TrackerApp, e *xerr.Error) {
	tl.Log(tl.Notice, palette.BlueBold, "%s for '%s'", "Initializing interface", windowTitle)

	// set up the app
	t = &TrackerApp{}
	t.App = app.NewWithID(appId)
	// Apply a slightly larger theme, light theme
	currentTheme := t.App.Settings().Theme()
	t.App.Settings().SetTheme(scaledTheme{
		base:   currentTheme, // or theme.LightTheme()/DarkTheme()
		factor: 1.30,
	})

	// Start large + fullscreen
	t.Window = t.App.NewWindow(windowTitle)
	t.Window.Resize(fyne.NewSize(1280, 720)) // initial size (before FS)
	// t.Window.SetFullScreen(true)          // launch fullscreen

	// --- System tray (desktop only) ---
	deskApp, ok := t.App.(desktop.App)
	if ok {
		t.DeskApp = deskApp

		// Resolve icon paths relative to the executable (not CWD)
		exePath, err := os.Executable()
		if err != nil {
			return nil, xerr.NewError(err, "Unable to resolve executable path", exePath)
		}
		exeDir := filepath.Dir(exePath)

		loadIcon := func(relPath string) (r fyne.Resource, e *xerr.Error) {
			rel := filepath.FromSlash(relPath)

			abs := filepath.Join(exeDir, rel)
			res, err := fyne.LoadResourceFromPath(abs)
			if err == nil {
				return res, nil
			}

			// Fallback to working-directory relative (dev runs)
			cwd := filepath.FromSlash("./" + relPath)
			res, err2 := fyne.LoadResourceFromPath(cwd)
			if err2 != nil {
				return nil, xerr.NewError(err2, "Unable to load tray icon", abs)
			}
			return res, nil
		}

		blue, e := loadIcon("pictures/tray-clock-blue-24.png")
		if e != nil {
			return nil, e
		}
		green, e := loadIcon("pictures/tray-clock-green-24.png")
		if e != nil {
			return nil, e
		}

		t.TrayIconBlue = blue
		t.TrayIconGreen = green

		// Initial state: not running => blue
		t.Window.SetIcon(t.TrayIconBlue)
		t.DeskApp.SetSystemTrayIcon(t.TrayIconBlue)

		m := fyne.NewMenu("Work Tracker",
			fyne.NewMenuItem("Show", func() {
				t.Window.Show()
				t.Window.RequestFocus()
			}),
			fyne.NewMenuItem("Hide", func() {
				t.Window.Hide()
			}),
			fyne.NewMenuItemSeparator(),
			fyne.NewMenuItem("Quit", func() {
				t.App.Quit()
			}),
		)
		t.DeskApp.SetSystemTrayMenu(m)
	} else {
		tl.Log(tl.Notice1, palette.YellowBold, "%s", "System tray not supported on this platform/driver")
	}

	// title canvas
	t.Title = canvas.NewText("Today", theme.Color(theme.ColorNameForeground))
	t.Title.Alignment = fyne.TextAlignCenter
	t.Title.TextStyle = fyne.TextStyle{Bold: true}
	t.Title.TextSize = theme.TextSize() * 2.0 // 2x normal

	// task name canva
	t.TaskLabel = canvas.NewText("Current Task", theme.Color(theme.ColorNameForeground))
	t.TaskLabel.Alignment = fyne.TextAlignCenter
	t.TaskLabel.TextStyle = fyne.TextStyle{Bold: false}
	t.TaskLabel.TextSize = theme.TextSize() * 2.0 // 2x normal

	// clock widget
	t.Clock = canvas.NewText("00:00:00", theme.Color(theme.ColorNameForeground))
	t.Clock.Alignment = fyne.TextAlignCenter
	t.Clock.TextStyle = fyne.TextStyle{Monospace: true, Bold: true}
	t.Clock.TextSize = theme.TextSize() * 3.2 // really big

	// activity bars
	t.AverageActivityBar = NewActivityBar("Average activity")
	t.CurrentActivityBar = NewActivityBar("Current activity")

	// start button
	t.Button = widget.NewButtonWithIcon("Start", theme.MediaPlayIcon(), nil)
	t.Button.Importance = widget.MediumImportance

	// after you computed tickers & LastTickStart...
	tasks, e := loadTasks(tasksFilePath)
	if e != nil {
		return t, e
	}
	t.TasksContainer = t.makeTasksUI(tasks)

	tl.Log(tl.Notice1, palette.GreenBold, "%s for '%s'", "Initialized interface", windowTitle)
	return t, nil
}
