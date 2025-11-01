package worktracker

import (
	"maps"
	"time"

	er "work-tracker/src/pkg/error"
	"work-tracker/src/pkg/logger"
)

func InitializeTrackerApp(appId, windowTitle, workDir string, tickInterval, flushInterval time.Duration) (trackerApp *TrackerApp, e *er.Error) {
	logger.Log(
		logger.Important, logger.BoldBlueColor,
		"%s tracker app. App id: '%s', window title: '%s', work dir: '%s', tick interval: %s, flush interval: '%s'",
		"Initializing", appId, windowTitle, workDir, tickInterval, flushInterval,
	)

	trackerApp ,e = initializeInterface(appId, windowTitle)
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
