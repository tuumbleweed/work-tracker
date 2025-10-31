package worktracker

import (
	"fmt"
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
	t.Button.OnTapped = t.onButtonTapped
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
		t.Title,
		t.Clock,
		t.Status,
		container.NewCenter(t.Button),
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

	// get the data
	t.Mutex.Lock()
	isRunning := t.IsRunning
	workedToday := t.WorkedToday
	activeToday := t.ActiveToday
	lastTickActiveDuration := t.LastTickActiveDuration
	t.Mutex.Unlock()

	activeToday = Clamp(activeToday, 0, workedToday)
	todayAverageActivityPrecentage := getActivityPercentage(activeToday, workedToday)
	lastTickActivityPercentage := getActivityPercentage(lastTickActiveDuration, t.TickInterval)

	clockText := formatDuration(workedToday)
	activityText := fmt.Sprintf("Average activity: %.1f%%, Current activity: %.1f%%", todayAverageActivityPrecentage, lastTickActivityPercentage)

	fyne.Do(func() {
		// update clock
		t.Clock.SetText(clockText)
		// update activity
		t.Status.SetText(activityText)

		// update button
		if isRunning {
			t.Button.SetText("Stop")
			t.Button.SetIcon(theme.MediaPauseIcon())
		} else {
			t.Button.SetText("Start")
			t.Button.SetIcon(theme.MediaPlayIcon())
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

	idleMs := tryXprintidle() // milliseconds since last input (may be -1 on error)

	// calculate active time for the last tick
	t.LastTickActiveDuration = 0
	lastTickDurationMs := time.Since(t.LastTickStart).Milliseconds()
	if lastTickDurationMs > 0 && idleMs >= 0 {
		idleInWindow := Min(idleMs, lastTickDurationMs)
		activeMs := lastTickDurationMs - idleInWindow
		t.LastTickActiveDuration = time.Duration(activeMs) * time.Millisecond
	}

	// worked before this run + this run
	t.WorkedToday = t.WorkedTodayBeforeStartingThisRun + now.Sub(t.RunStart)
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
		t.ChunkStart = now
	} else {
		// stopping
		t.IsRunning = false
		// set new t.WorkedTodayBeforeStartingThisRun
		t.WorkedTodayBeforeStartingThisRun = t.WorkedToday
	}
	t.Mutex.Unlock()
	logger.Log(logger.Verbose1, logger.GreenColor, "%s", "Flipped switch")
}

func (t *TrackerApp) flushChunkIfRunning() {
	t.Mutex.Lock()
	if t.IsRunning {
		now := time.Now()
		e := flushChunk(t.CurrentFilePath, t.ChunkStart, now, t.ActiveDuringThisChunk)
		if e != nil {
			e.QuitIf("error") // don't expect any errors here, so quit if found one
		}
		t.ActiveDuringThisChunk = 0
		t.ChunkStart = now
	}
	t.Mutex.Unlock()
}
