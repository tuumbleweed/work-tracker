package worktracker

import (
	"maps"
	"time"

	tl "github.com/tuumbleweed/tintlog/logger"
	"github.com/tuumbleweed/tintlog/palette"
	"github.com/tuumbleweed/xerr"
)

func InitializeTrackerApp(appId, windowTitle, workDir, tasksFilePath string, tickInterval, flushInterval time.Duration) (trackerApp *TrackerApp, e *xerr.Error) {
	tl.Log(
		tl.Important, palette.BlueBold,
		"%s tracker app. App id: '%s', window title: '%s', work dir: '%s', tick interval: %s, flush interval: '%s'",
		"Initializing", appId, windowTitle, workDir, tickInterval, flushInterval,
	)

	trackerApp, e = initializeInterface(appId, windowTitle, tasksFilePath)
	if e != nil {
		return trackerApp, e
	}

	// determine current file path
	trackerApp.Workdir = workDir
	trackerApp.CurrentDateID = dateID(time.Now())
	trackerApp.CurrentFilePath = dayFilePath(trackerApp.Workdir, trackerApp.CurrentDateID)

	// get information about total duration and active time
	trackerApp.WorkedToday, trackerApp.ActiveToday, trackerApp.TimeByTask, e = loadFileActivityAndDuration(trackerApp.CurrentFilePath)
	if e != nil {
		return trackerApp, e
	}
	trackerApp.WorkedTodayBeforeStartingThisRun = trackerApp.WorkedToday
	// copy by entry
	trackerApp.TimeByTaskBeforeStartingThisRun = make(map[string]time.Duration, len(trackerApp.TimeByTask))
	maps.Copy(trackerApp.TimeByTaskBeforeStartingThisRun, trackerApp.TimeByTask)

	// initialize tickers
	trackerApp.TickInterval = tickInterval
	trackerApp.FlushInterval = flushInterval
	trackerApp.Ticker = time.NewTicker(trackerApp.TickInterval)
	trackerApp.FlushTicker = time.NewTicker(trackerApp.FlushInterval)
	trackerApp.done = make(chan struct{})
	trackerApp.LastTickStart = time.Now()

	tl.Log(
		tl.Important1, palette.GreenBold,
		"%s tracker app. App id: '%s', window title: '%s', work dir: '%s', tick interval: %s, flush interval: '%s'",
		"Initialized", appId, windowTitle, workDir, tickInterval, flushInterval,
	)
	tl.Log(
		tl.Notice, palette.CyanBold,
		"\nWorkDir: '%s'\nCurrendDateID: '%s'\nCurrentFilePath: '%s'\nWorkedToday: '%s'\nActiveToday: '%s'",
		trackerApp.Workdir, trackerApp.CurrentDateID, trackerApp.CurrentFilePath,
		trackerApp.WorkedToday, trackerApp.ActiveToday,
	)
	return trackerApp, nil
}
