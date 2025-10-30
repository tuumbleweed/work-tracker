package worktracker

import (
	"time"
	
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"

	"my-project/src/pkg/logger"
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
		t.Mutex.Unlock()

		// update button
		t.Button.SetText("Stop")
		t.Button.SetIcon(theme.MediaPauseIcon())
		t.updateInterface()
		return
	} else {
		// stopping
		t.IsRunning = false
		t.Mutex.Unlock()

		// update button
		t.Button.SetText("Start")
		t.Button.SetIcon(theme.MediaPlayIcon())
		t.updateInterface()
	}
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

// update clock and activity labels
func (t *TrackerApp) updateInterface() {
	// now := time.Now()

	// // get the data
	// t.Mutex.Lock()
	// isRunning := t.IsRunning
	// runStart := t.RunStart
	// workedToday := t.WorkedToday
	// activeToday := t.ActiveToday
	// lastTickActiveDuration := t.LastTickActiveDuration
	// t.Mutex.Unlock()


}

func (t *TrackerApp) saveChunk() {

}