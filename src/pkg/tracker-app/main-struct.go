package trackerapp

import (
	"sync"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/widget"
)

type TrackerApp struct {
	App    fyne.App
	Window fyne.Window

	// UI elements
	Title              *canvas.Text
	TaskLabel          *canvas.Text
	Clock              *canvas.Text
	AverageActivityBar *ActivityBar
	CurrentActivityBar *ActivityBar
	Button             *widget.Button
	TableRows          map[string]TableRow
	TasksContainer     *fyne.Container

	// tickers
	Ticker        *time.Ticker  // UI and activity clock
	FlushTicker   *time.Ticker  // file chunk save clock
	TickInterval  time.Duration // for Ticker
	FlushInterval time.Duration // for FlushTicker
	done          chan struct{}

	// dirs
	Workdir         string
	CurrentDateID   string
	CurrentFilePath string

	// time
	WorkedTodayBeforeStartingThisRun time.Duration // for how long user tracked time today
	WorkedToday                      time.Duration // for how long user tracked time today
	ActiveToday                      time.Duration // how much out of that time user was active
	ActiveDuringThisChunk            time.Duration // how long user been active during this chunk
	LastTickActiveDuration           time.Duration // how much out of that user was active
	TimeByTaskBeforeStartingThisRun  map[string]time.Duration
	TimeByTask                       map[string]time.Duration

	// mutex
	Mutex sync.Mutex

	// run info
	IsRunning       bool
	RunStart        time.Time // when last pressed "start" button
	TaskRunStart    time.Time // when last pressed "start" button
	ChunkStart      time.Time // when last time chunk was saved
	LastTickStart   time.Time // when last tick has started
	CurrentTaskName string    // which task is running right now, can be empty
}

type TableRow struct {
	Button           *widget.Button
	NameLabel        *widget.Label
	DescriptionLabel *widget.Label
	CreatedAtLabel   *widget.Label
	TimeLabel        *widget.Label
}
