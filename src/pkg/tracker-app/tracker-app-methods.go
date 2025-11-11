package trackerapp

import (
	"maps"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"

	tl "github.com/tuumbleweed/tintlog/logger"
	"github.com/tuumbleweed/tintlog/palette"
)

func (t *TrackerApp) Start() {
	tl.Log(tl.Notice, palette.BlueBold, "%s", "Running work tracker app...")

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
			tl.Log(tl.Info, palette.Cyan, "%s. Task name: '%s'", "Stopping task", currentTaskName)
		} else {
			tl.Log(tl.Info, palette.Cyan, "%s. Task name: '%s'", "Starting task", currentTaskName)
		}
	}
	t.Window.SetCloseIntercept(t.onClose)

	t.setContent()

	go t.tickLoop()
	go t.flushLoop()

	t.updateInterface() // initial
	t.Window.ShowAndRun()

	tl.Log(tl.Notice, palette.GreenBold, "%s", "Closing work tracker app")
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
	tl.Log(tl.Verbose, palette.Blue, "%s", "Updating interface")

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

	for _, tableRow := range tableRows {
		setRowImportance(tableRow, widget.MediumImportance)
	}

	var currentTaskNameDisplay string // this is show above the clock
	if currentTaskName == "" {
		if isRunning {
			currentTaskNameDisplay = "Unassigned Task"
		} else {
			currentTaskNameDisplay = "Not Tracking"
		}
	} else {
		currentTaskNameDisplay = currentTaskName
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
		t.TaskLabel.Text = currentTaskNameDisplay
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
		} else {
			t.Button.SetText("Start")
			showStopped(t.Button)
		}

		// update table rows
		for taskName, tableRow := range tableRows {
			tableRow.TimeLabel.Text = formatDuration(timeByTask[taskName])
			tableRow.TimeLabel.Refresh()
		}

		if currentTaskName != "" {
			setRowImportance(tableRows[currentTaskName], widget.HighImportance)
		}
	})
	tl.Log(tl.Verbose1, palette.Green, "%s", "Updated interface")
}

/*
Update TrackerApp underlying paramenters. Runs every tick.

Do it here instead of doing everything in updateInterface.
*/
func (t *TrackerApp) refreshState() {
	tl.Log(tl.Verbose, palette.Blue, "%s", "Refreshing state")

	now := time.Now()

	t.Mutex.Lock()
	// no state updates if not running.
	if !t.IsRunning {
		tl.Log(tl.Verbose1, palette.Cyan, "%s", "No need to refresh state")
		t.LastTickStart = now
		t.Mutex.Unlock()
		return
	}

	idleMs := tryXprintidle() // milliseconds since last input (may be -1 on error)
	t.LastTickActiveDuration = 0
	lastTickDurationMs := time.Since(t.LastTickStart).Milliseconds()
	var activeMs int64
	if idleMs >= lastTickDurationMs {
		// if was idle this whole time block or longer
		// then mark time block as non-active
		activeMs = 0
	} else {
		// but otherwise make it fully active
		activeMs = lastTickDurationMs
	}
	t.LastTickActiveDuration = time.Duration(activeMs) * time.Millisecond

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

	tl.Log(tl.Verbose1, palette.Green, "%s", "Refreshed state")
}

// Runs when we press on start/stop button
func (t *TrackerApp) flipSwitch() {
	tl.Log(tl.Verbose, palette.Blue, "%s", "Flipping switch")
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
	tl.Log(tl.Verbose1, palette.Green, "%s", "Flipped switch")
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
