package worktracker

import (
	"maps"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"

	"work-tracker/src/pkg/logger"
)

func (t *TrackerApp) Start() {
	logger.Log(logger.Notice, logger.BoldBlueColor, "%s", "Running work tracker app...")

	// set functions
	t.Button.OnTapped = func() {
		t.onButtonTapped()
		// when main button stopped - showStopped all the buttons
		t.Mutex.Lock()
		allTableRows := t.TableRows
		currentTaskName := t.CurrentTaskName
		isRunning := t.IsRunning
		t.Mutex.Unlock()
		if !isRunning {
			for _, taskRow := range allTableRows {
				showStopped(taskRow.Button)
			}
			logger.Log(logger.Info, logger.CyanColor, "%s. Task name: '%s'", "Stopping task", currentTaskName)
		} else {
			logger.Log(logger.Info, logger.CyanColor, "%s. Task name: '%s'", "Starting task", currentTaskName)
		}
	}
	t.Window.SetCloseIntercept(t.onClose)

	t.setContent()

	go t.tickLoop()
	go t.flushLoop()

	t.updateInterface() // initial
	t.Window.ShowAndRun()

	logger.Log(logger.Notice, logger.BoldGreenColor, "%s", "Closing work tracker app")
}

func (t *TrackerApp) setContent() {
	content := container.New(
		layout.NewVBoxLayout(),
		vgap(1, 10),
		t.Title,
		vgap(1, 10),
		t.TaskLabel,
		vgap(1, 10),
		t.Clock,
		vgap(1, 5),
		t.AverageActivityBar,
		t.CurrentActivityBar,
		vgap(1, 10),
		container.NewCenter(t.Button),
		vgap(1, 10),
		t.TasksContainer,
		vgap(1, 10),
	)
	t.Window.SetContent(container.NewPadded(content))
}

func (t *TrackerApp) onButtonTapped() {
	t.refreshState()
	t.flushChunkIfRunning()
	t.flipSwitch()
	t.updateInterface()
}

func (t *TrackerApp) onClose() {
	close(t.done)
	t.Ticker.Stop()
	t.FlushTicker.Stop()
	// flush current run if any (only works when t.IsRunning == true)
	t.flushChunkIfRunning()

	t.Window.Close()
}

func (t *TrackerApp) tickLoop() {
	for {
		select {
		case <-t.Ticker.C:
			t.refreshState()
			t.updateInterface()
		case <-t.done:
			return
		}
	}
}

func (t *TrackerApp) flushLoop() {
	for {
		select {
		case <-t.FlushTicker.C:
			t.flushChunkIfRunning()
		case <-t.done:
			return
		}
	}
}

/*
Update clock and activity labels.

This function does not change any TrackerApp values. Only updates the interface.
*/
func (t *TrackerApp) updateInterface() {
	logger.Log(logger.Verbose, logger.BlueColor, "%s", "Updating interface")

	now := time.Now()

	// get the data
	t.Mutex.Lock()
	isRunning := t.IsRunning
	workedToday := t.WorkedToday
	activeToday := t.ActiveToday
	lastTickActiveDuration := t.LastTickActiveDuration
	currentTaskName := t.CurrentTaskName
	tableRows := t.TableRows
	timeByTask := t.TimeByTask
	t.Mutex.Unlock()

	if currentTaskName == "" {
		if isRunning {
			currentTaskName = "Unassigned Task"
		} else {
			currentTaskName = "Not Tracking"
		}
	}

	activeToday = Clamp(activeToday, 0, workedToday)
	todayAverageActivityPercentage := getActivityPercentage(activeToday, workedToday)
	lastTickActivityPercentage := getActivityPercentage(lastTickActiveDuration, t.TickInterval)

	clockText := formatDuration(workedToday)

	titleText := now.Format("Monday, January 02, 15:04:05")

	fyne.Do(func() {
		// Update title
		t.Title.Text = titleText
		t.Title.Refresh()
		// update clock
		t.Clock.Text = clockText
		t.TaskLabel.Text = currentTaskName
		if t.IsRunning {
			t.Clock.Color = getActiveColor()
			t.TaskLabel.Color = getActiveColor()
		} else {
			t.Clock.Color = theme.Color(theme.ColorNameForeground)     // revert to default theme color
			t.TaskLabel.Color = theme.Color(theme.ColorNameForeground) // revert to default theme color
		}
		t.Clock.Refresh()
		t.TaskLabel.Refresh()
		// update activity bars
		t.AverageActivityBar.SetPercent(todayAverageActivityPercentage)
		t.CurrentActivityBar.SetPercent(lastTickActivityPercentage)

		// update button
		if isRunning {
			t.Button.SetText("Stop")
			showRunning(t.Button)
			// t.Button.Importance = widget.WarningImportance // Importance change needs a Refresh()
			// t.Button.SetIcon(theme.MediaPauseIcon()) // SetIcon calls Refresh
		} else {
			t.Button.SetText("Start")
			showStopped(t.Button)
			// t.Button.Importance = widget.MediumImportance
			// t.Button.SetIcon(theme.MediaPlayIcon())
		}

		// update table rows
		for taskName, tableRow := range tableRows {
			tableRow.TimeLabel.Text = formatDuration(timeByTask[taskName])
			tableRow.TimeLabel.Refresh()
		}
	})
	logger.Log(logger.Verbose1, logger.GreenColor, "%s", "Updated interface")
}

/*
Update TrackerApp underlying paramenters. Runs every tick.

Do it here instead of doing everything in updateInterface.
*/
func (t *TrackerApp) refreshState() {
	logger.Log(logger.Verbose, logger.BlueColor, "%s", "Refreshing state")

	now := time.Now()

	t.Mutex.Lock()
	// no state updates if not running.
	if !t.IsRunning {
		logger.Log(logger.Verbose1, logger.CyanColor, "%s", "No need to refresh state")
		t.LastTickStart = now
		t.Mutex.Unlock()
		return
	}

	// the way it works right now is that if we input anything at the very end of the
	// tick then the whole tick will get near 100% activity.
	// that's why we need to keep tick size small right now (500ms will do).
	// UI and activity share same ticker at the moment.
	// later we can implement a separate tick that would sample activity in shorter periods
	// calculate active time for the last tick
	idleMs := tryXprintidle() // milliseconds since last input (may be -1 on error)
	t.LastTickActiveDuration = 0
	lastTickDurationMs := time.Since(t.LastTickStart).Milliseconds()
	if lastTickDurationMs > 0 && idleMs >= 0 {
		idleInWindow := Min(idleMs, lastTickDurationMs)
		activeMs := lastTickDurationMs - idleInWindow
		t.LastTickActiveDuration = time.Duration(activeMs) * time.Millisecond
	}

	// worked before this run + this run
	workedThisRun := now.Sub(t.RunStart)
	workedThisRunOnTask := now.Sub(t.TaskRunStart)
	t.WorkedToday = t.WorkedTodayBeforeStartingThisRun + workedThisRun
	t.TimeByTask[t.CurrentTaskName] = t.TimeByTaskBeforeStartingThisRun[t.CurrentTaskName] + workedThisRunOnTask
	// add last active duration to use later
	t.ActiveToday += t.LastTickActiveDuration
	// add last active duration to t.ActiveDuringThisChunk (it's emptied on each flush)
	t.ActiveDuringThisChunk += t.LastTickActiveDuration
	t.LastTickStart = now
	t.Mutex.Unlock()

	logger.Log(logger.Verbose1, logger.GreenColor, "%s", "Refreshed state")
}

// Runs when we press on start/stop button
func (t *TrackerApp) flipSwitch() {
	logger.Log(logger.Verbose, logger.BlueColor, "%s", "Flipping switch")
	t.Mutex.Lock()
	if !t.IsRunning {
		// starting
		t.IsRunning = true
		now := time.Now()
		t.RunStart = now
		t.TaskRunStart = now
		t.ChunkStart = now
	} else {
		// stopping
		t.IsRunning = false
		t.LastTickActiveDuration = 0 // empty this to show 0% when idle
		// set new t.WorkedTodayBeforeStartingThisRun
		t.WorkedTodayBeforeStartingThisRun = t.WorkedToday
		// set new t.TimeByTaskBeforeStartingThisRun
		maps.Copy(t.TimeByTaskBeforeStartingThisRun, t.TimeByTask)
	}
	t.CurrentTaskName = "" // set to empty here, can be overriden later
	t.Mutex.Unlock()
	logger.Log(logger.Verbose1, logger.GreenColor, "%s", "Flipped switch")
}

func (t *TrackerApp) flushChunkIfRunning() {
	t.Mutex.Lock()
	if t.IsRunning {
		now := time.Now()
		e := flushChunk(t.CurrentFilePath, t.ChunkStart, now, t.ActiveDuringThisChunk, t.CurrentTaskName)
		if e != nil {
			e.QuitIf("error") // don't expect any errors here, so quit if found one
		}
		t.ActiveDuringThisChunk = 0
		t.ChunkStart = now
	}
	t.Mutex.Unlock()
}
