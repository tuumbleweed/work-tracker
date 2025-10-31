package worktracker

import (
	"fmt"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"

	"work-tracker/src/pkg/logger"
)

func (t *TrackerApp) Start() {
	logger.Log(logger.Notice, logger.BoldBlueColor, "%s", "Running work tracker app...")

	// set functions
	t.Button.OnTapped = t.onButtonTapped
	t.Window.SetCloseIntercept(t.onClose)

	// build window elements
	title := widget.NewLabel("Today")
	title.Alignment = fyne.TextAlignCenter
	title.TextStyle = fyne.TextStyle{Bold: true}

	content := container.New(
		layout.NewVBoxLayout(),
		title,
		t.Clock,
		t.Status,
		container.NewCenter(t.Button),
	)
	t.Window.SetContent(container.NewPadded(content))

	go t.tickLoop()
	go t.flushLoop()

	t.updateInterface() // initial
	t.Window.ShowAndRun()

	logger.Log(logger.Notice, logger.BoldGreenColor, "%s", "Closing work tracker app")
}

func (t *TrackerApp) onButtonTapped() {
	now := time.Now()
	t.Mutex.Lock()
	if !t.IsRunning {
		// starting
		t.IsRunning = true
		t.RunStart = now
	} else {
		// stopping
		t.IsRunning = false
	}
	t.Mutex.Unlock()
	t.updateInterface()
}

func (t *TrackerApp) onClose() {
	close(t.done)
	t.Ticker.Stop()
	t.FlushTicker.Stop()

	// flush current run if any
	t.Mutex.Lock()
	isRunning := t.IsRunning
	// runStart := t.RunStart
	t.Mutex.Unlock()

	if isRunning {
		// write to file here
	}

	t.Window.Close()
}

func (t *TrackerApp) tickLoop() {
	for {
		select {
		case <-t.Ticker.C:
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
			t.saveChunk()
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
	// // now := time.Now()

	// get the data
	t.Mutex.Lock()
	isRunning := t.IsRunning
	// runStart := t.RunStart
	workedToday := t.WorkedToday
	activeToday := t.ActiveToday
	lastTickActiveDuration := t.LastTickActiveDuration
	t.Mutex.Unlock()

	todayAverageActivityPrecentage := getActivityPercentage(activeToday, workedToday)
	lastTickActivityPercentage := getActivityPercentage(lastTickActiveDuration, t.TickInterval)
	// fmt.Printf("Activity: %.1f%%\n", lastTickActivityPercentage) // Output: 75.4%

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
}

func (t *TrackerApp) saveChunk() {

}
