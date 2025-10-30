package worktracker

import (
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"

	er "my-project/src/pkg/error"
	"my-project/src/pkg/logger"
)

func NewTrackerApp(appId, windowTitle, workDir string, uiInterval, chunkInterval time.Duration) (trackerApp *TrackerApp, e *er.Error) {
	trackerApp = &TrackerApp{}
	trackerApp.WorktrackerApp = app.NewWithID(appId)
	trackerApp.WorktrackerWindow = trackerApp.WorktrackerApp.NewWindow(windowTitle)
	trackerApp.WorktrackerWindow.Resize(fyne.NewSize(420, 220))

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
	trackerApp.TotalDuration, trackerApp.TotalActiveTime, e = loadFileActivityAndDuration(trackerApp.CurrentFilePath)
	if e != nil {
		return trackerApp, e
	}

	logger.Log(
		logger.Notice, logger.CyanColor,
		"\nWorkDir: '%s'\nCurrendDateID: '%s'\nCurrentFilePath: '%s'\nTotalDuration: '%s'\nTotalActiveTime: '%s'",
		trackerApp.Workdir, trackerApp.CurrentDateID, trackerApp.CurrentFilePath,
		trackerApp.TotalDuration, trackerApp.TotalActiveTime,
	)
	return trackerApp, nil
}
