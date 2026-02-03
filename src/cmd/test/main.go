package main

import (
	"log"
	"path/filepath"

	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/driver/desktop"
	"fyne.io/fyne/v2"
)

func main() {
	a := app.New()
	w := a.NewWindow("Work Tracker")

	var desk desktop.App
	var ok bool
	desk, ok = a.(desktop.App)
	if !ok {
		// Not running on desktop driver (or tray not supported here).
		w.ShowAndRun()
		return
	}

	var icon fyne.Resource
	var e error

	// Path is relative to where you run the binary from.
	icon, e = fyne.LoadResourceFromPath(filepath.FromSlash("./pictures/tray-clock-blue-24.png"))
	if e != nil {
		log.Fatalf("load tray icon: %v", e)
	}

	desk.SetSystemTrayIcon(icon)

	m := fyne.NewMenu("Work Tracker",
		fyne.NewMenuItem("Show", func() { w.Show() }),
		fyne.NewMenuItem("Hide", func() { w.Hide() }),
		fyne.NewMenuItemSeparator(),
		fyne.NewMenuItem("Quit", func() { a.Quit() }),
	)
	desk.SetSystemTrayMenu(m)

	w.ShowAndRun()
}
