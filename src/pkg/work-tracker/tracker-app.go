package worktracker

import (
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"

	er "work-tracker/src/pkg/error"
	"work-tracker/src/pkg/logger"
)

func InitializeTrackerApp(appId, windowTitle, workDir string, tickInterval, flushInterval time.Duration) (trackerApp *TrackerApp, e *er.Error) {
	logger.Log(
		logger.Important, logger.BoldBlueColor,
		"%s tracker app. App id: '%s', window title: '%s', work dir: '%s', tick interval: %s, flush interval: '%s'",
		"Initializing", appId, windowTitle, workDir, tickInterval, flushInterval,
	)

	trackerApp = &TrackerApp{}
	trackerApp.App = app.NewWithID(appId)
	trackerApp.Window = trackerApp.App.NewWindow(windowTitle)
	trackerApp.Window.Resize(fyne.NewSize(420, 220))

	// clock widget
	trackerApp.Clock = widget.NewLabel("00:00:00")
	trackerApp.Clock.Alignment = fyne.TextAlignCenter
	trackerApp.Clock.TextStyle = fyne.TextStyle{Monospace: true, Bold: true}
	trackerApp.Clock.Wrapping = fyne.TextWrapOff
	trackerApp.Clock.Importance = widget.MediumImportance

	// status widget
	trackerApp.Status = widget.NewLabel("stopped")
	trackerApp.Status.Alignment = fyne.TextAlignCenter

	// start button
	trackerApp.Button = widget.NewButtonWithIcon("Start", theme.MediaPlayIcon(), nil)

	// determine current file path
	trackerApp.Workdir = workDir
	trackerApp.CurrentDateID = dateID(time.Now())
	trackerApp.CurrentFilePath = dayFilePath(trackerApp.Workdir, trackerApp.CurrentDateID)

	// get information about total duration and active time
	trackerApp.WorkedToday, trackerApp.ActiveToday, e = loadFileActivityAndDuration(trackerApp.CurrentFilePath)
	if e != nil {
		return trackerApp, e
	}
	trackerApp.WorkedTodayBeforeStartingThisRun = trackerApp.WorkedToday

	// initialize tickers
	trackerApp.TickInterval = tickInterval
	trackerApp.FlushInterval = flushInterval
	trackerApp.Ticker = time.NewTicker(trackerApp.TickInterval)
	trackerApp.FlushTicker = time.NewTicker(trackerApp.FlushInterval)
	trackerApp.done = make(chan struct{})

	logger.Log(
		logger.Important1, logger.BoldGreenColor,
		"%s tracker app. App id: '%s', window title: '%s', work dir: '%s', tick interval: %s, flush interval: '%s'",
		"Initialized", appId, windowTitle, workDir, tickInterval, flushInterval,
	)
	logger.Log(
		logger.Notice, logger.BoldCyanColor,
		"\nWorkDir: '%s'\nCurrendDateID: '%s'\nCurrentFilePath: '%s'\nWorkedToday: '%s'\nActiveToday: '%s'",
		trackerApp.Workdir, trackerApp.CurrentDateID, trackerApp.CurrentFilePath,
		trackerApp.WorkedToday, trackerApp.ActiveToday,
	)
	return trackerApp, nil
}
