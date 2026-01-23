package trackerapp

import (
	"maps"
	"time"

	tl "github.com/tuumbleweed/tintlog/logger"
	"github.com/tuumbleweed/tintlog/palette"
	"github.com/tuumbleweed/xerr"

	"work-tracker/src/pkg/util"
)

func InitializeTrackerApp(appId, windowTitle, workDir, tasksFilePath string, uiTickInterval, activityTickInterval, flushInterval time.Duration) (trackerApp *TrackerApp, e *xerr.Error) {
	tl.Log(
		tl.Important, palette.BlueBold,
		"%s tracker app. App id: '%s', window title: '%s', work dir: '%s', UI tick interval: %s, activity tick interval: %s, flush tick interval: '%s'",
		"Initializing", appId, windowTitle, workDir, uiTickInterval, activityTickInterval, flushInterval,
	)

	trackerApp, e = initializeInterface(appId, windowTitle, tasksFilePath)
	if e != nil {
		return trackerApp, e
	}

	// determine current file path
	trackerApp.Workdir = workDir
	trackerApp.CurrentYear, trackerApp.CurrentMonth, trackerApp.CurrentDay = dateID(time.Now())
	trackerApp.CurrentDirPath, trackerApp.CurrentFilePath = dayFilePath(trackerApp.Workdir, trackerApp.CurrentYear, trackerApp.CurrentMonth, trackerApp.CurrentDay)
	e = util.EnsureDirExists(trackerApp.CurrentDirPath, 0755)
	if e != nil {
		return trackerApp, e
	}

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
	trackerApp.UITickInterval = uiTickInterval
	trackerApp.ActivityTickInterval = activityTickInterval
	trackerApp.FlushTickInterval = flushInterval
	trackerApp.ActivityTicker = time.NewTicker(trackerApp.ActivityTickInterval)
	trackerApp.UITicker = time.NewTicker(trackerApp.UITickInterval)
	trackerApp.FlushTicker = time.NewTicker(trackerApp.FlushTickInterval)
	trackerApp.done = make(chan struct{})
	trackerApp.LastActivityTickStart = time.Now()

	tl.Log(
		tl.Important1, palette.GreenBold,
		"%s tracker app. App id: '%s', window title: '%s', work dir: '%s', UI tick interval: %s, activity tick interval: %s, flush tick interval: '%s'",
		"Initialized", appId, windowTitle, workDir, uiTickInterval, activityTickInterval, flushInterval,
	)
	tl.Log(
		tl.Notice, palette.CyanBold,
		"\nWorkDir: '%s'\nCurrentYear: '%s'\nCurrentMonth: '%s'\nCurrentDay: '%s'\nCurrentFilePath: '%s'\nWorkedToday: '%s'\nActiveToday: '%s'",
		trackerApp.Workdir, trackerApp.CurrentYear, trackerApp.CurrentMonth, trackerApp.CurrentDay,
		trackerApp.CurrentFilePath, trackerApp.WorkedToday, trackerApp.ActiveToday,
	)
	return trackerApp, nil
}
