package worktracker

import (
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/widget"
)

type TrackerApp struct {
	WorktrackerApp    fyne.App
	WorktrackerWindow fyne.Window

	// UI elements
	Clock  *widget.Label
	Status *widget.Label
	Button *widget.Button

	// tickers
	UITicker        *time.Ticker // 1s UI clock
	ChunkSaveTicker *time.Ticker // file chunk save clock

	// dirs
	Workdir         string
	CurrentDateID   string
	CurrentFilePath string

	// time
	TotalDuration   time.Duration // for how long user tracked time today
	TotalActiveTime time.Duration // how much out of that time user was active

	// activity
	InstantActivity time.Duration // recent poll window
}
