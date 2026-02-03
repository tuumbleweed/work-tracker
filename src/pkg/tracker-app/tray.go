package trackerapp

import (
	"os"
	"path/filepath"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/driver/desktop"

	tl "github.com/tuumbleweed/tintlog/logger"
	"github.com/tuumbleweed/tintlog/palette"
	"github.com/tuumbleweed/xerr"
)

// initTray sets up system tray (icons + menu). No-op if driver doesn't support tray.
func (t *TrackerApp) initTray() (e *xerr.Error) {
	if t == nil {
		return nil
	}
	if t.App == nil {
		return nil
	}
	if t.Window == nil {
		return nil
	}

	deskApp, ok := t.App.(desktop.App)
	if !ok {
		tl.Log(tl.Notice1, palette.YellowBold, "%s", "System tray not supported on this platform/driver")
		return nil
	}
	t.DeskApp = deskApp

	// Resolve icon paths relative to the executable (not CWD)
	exePath, err := os.Executable()
	if err != nil {
		return xerr.NewError(err, "Unable to resolve executable path", exePath)
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
		return e
	}
	green, e := loadIcon("pictures/tray-clock-green-24.png")
	if e != nil {
		return e
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
			// Call cleanup path, not just Quit, so tray gets cleared, tickers stop, etc.
			t.onClose()
		}),
	)
	t.DeskApp.SetSystemTrayMenu(m)

	return nil
}
