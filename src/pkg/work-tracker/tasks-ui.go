package worktracker

import (
	"image/color"
	"maps"
	"time"
	"work-tracker/src/pkg/logger"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

// column widths (px) – tweak to taste
const (
	colPlayButtonWidth  = 80
	colNameWidth        = 260
	colDescriptionWidth = 420
	colCreatedAtWidth   = 260
	colHoursWidth       = 100
	// single-line row height
	rowHeight = 50
)

func (t *TrackerApp) makeTasksUI(tasks []Task) *fyne.Container {
	t.TableRows = make(map[string]TableRow)
	// Title
	sectionTitle := canvas.NewText("Tasks", theme.Color(theme.ColorNameForeground))
	sectionTitle.Alignment = fyne.TextAlignCenter
	sectionTitle.TextStyle = fyne.TextStyle{Bold: true}
	sectionTitle.TextSize = theme.TextSize() * 1.6

	// header
	leftHeader := container.NewHBox(
		fixedCell(labelHeader(""), colPlayButtonWidth),
		fixedCell(labelHeader("Task"), colNameWidth),
	)
	rightHeader := container.NewHBox(
		fixedCell(labelHeader("Created At"), colCreatedAtWidth),
		fixedCell(labelHeader("Hours"), colHoursWidth),
	)
	descHead := minWidth(labelHeader("Description"), colDescriptionWidth) // e.g. colDescriptionWidth px minimum
	header := container.NewBorder(nil, nil, leftHeader, rightHeader, descHead)

	// rows
	rows := container.NewVBox()
	for _, task := range tasks {
		// left group: ▶ + Task (both fixed widths)
		_, nameCanvas := fixedCellCenteredTruncated(task.Name, colNameWidth)
		rowPlayButton, playCell := smallButton(theme.MediaPlayIcon(), nil)
		leftBox := container.NewHBox(playCell, nameCanvas)

		// need to use mutex with those functions
		rowPlayButton.OnTapped = func() {
			t.Mutex.Lock()
			isRunning := t.IsRunning
			allTableRows := t.TableRows
			previousTaskName := t.CurrentTaskName
			t.Mutex.Unlock()
			var newTaskName string = task.Name
			var taskRunStart time.Time

			if isRunning {
				// it's running which means we either stop it or start another one

				// first showStopped all the buttons

				// if previous task name NOT equal to the new one - keep running,
				// then showRunning the playButton
				// if previous task name IS equal to the new one - flipSwitch (onButtonTapped)

				// we can first showStopped all of the buttons
				for _, tableRow := range allTableRows {
					showStopped(tableRow.Button)
				}
				// then handle the differences
				if previousTaskName != newTaskName {
					showRunning(rowPlayButton)

					// make sure to update t.TaskRunStart so that new task does not receive additional time
					taskRunStart = time.Now()
					newTaskName = task.Name

					logger.Log(logger.Info, logger.CyanColor, "%s at %s. Previous: '%s', New: '%s'", "Switching tasks", taskRunStart.String(), previousTaskName, newTaskName)
				} else {
					t.onButtonTapped()
					newTaskName = "" // reset newTaskName on stopping
					logger.Log(logger.Info, logger.CyanColor, "%s. Previous: '%s', New: '%s'", "Stopping task", previousTaskName, newTaskName)
				}
			} else {
				// it's not running which means we are starting a new task

				// flip the switch then show running

				t.onButtonTapped()
				showRunning(rowPlayButton)
				newTaskName = task.Name
				logger.Log(logger.Info, logger.CyanColor, "%s. Previous: '%s', New: '%s'", "Startng new task", previousTaskName, newTaskName)
			}

			// now set t.CurrentTaskName to newTaskName
			t.Mutex.Lock()
				t.CurrentTaskName = newTaskName
				if !taskRunStart.IsZero() {
					t.TaskRunStart = taskRunStart
					// save the progress
					maps.Copy(t.TimeByTaskBeforeStartingThisRun, t.TimeByTask)
				}
			t.Mutex.Unlock()
			t.updateInterface() // update interface to show current task name right away
		}

		// center: Description (expands; ellipsis)
		description := flexVCenterTruncated(task.Description)

		// right group: Created + Hours (both fixed)
		_, createdAtCanvas := fixedCellCenteredTruncated(task.CreatedAt.Format("Mon Jan 02 2006 15:04:05"), colCreatedAtWidth)
		timeLabel, timeCanvas := fixedCellCenteredTruncated(t.TimeByTask[task.Name].String(), colHoursWidth)
		right := container.NewHBox(createdAtCanvas, timeCanvas)

		t.TableRows[task.Name] = TableRow{
			Button:    rowPlayButton,
			TimeLabel: timeLabel,
		}

		row := container.NewBorder(nil, nil, leftBox, right, description)
		rows.Add(row)
	}

	return container.NewVBox(sectionTitle, header, rows)
}

func labelHeader(s string) *widget.Label {
	l := widget.NewLabelWithStyle(s, fyne.TextAlignLeading, fyne.TextStyle{Bold: true})
	l.Wrapping = fyne.TextWrap(fyne.TextTruncateClip)
	return l
}

func fixedCell(obj fyne.CanvasObject, w int) fyne.CanvasObject {
	return container.New(
		layout.NewGridWrapLayout(fyne.NewSize(float32(w), rowHeight)),
		obj,
	)
}

// minWidth wraps obj so it can expand but never be narrower than minW.
func minWidth(obj fyne.CanvasObject, minW int) fyne.CanvasObject {
	shim := canvas.NewRectangle(color.Transparent)
	shim.SetMinSize(fyne.NewSize(float32(minW), rowHeight)) // rowH if you want a row-height floor
	return container.NewStack(shim, obj)                    // shim contributes MinSize; obj draws on top
}

// fixedTruncVCenterCell gives obj the full cell size (so truncation works) and
// centers it vertically with spacers. Text stays left-aligned.
func fixedCellCenteredTruncated(text string, w int) (label *widget.Label, canvas fyne.CanvasObject) {
	l := widget.NewLabel(text)
	l.Wrapping = fyne.TextWrapOff            // no wrapping
	l.Truncation = fyne.TextTruncateEllipsis // "…" when too long
	l.Alignment = fyne.TextAlignLeading      // left aligned

	// VBox with spacers => vertical center; MaxLayout => child gets full cell size
	v := container.NewVBox(layout.NewSpacer(), l, layout.NewSpacer())
	return l, container.New(
		layout.NewGridWrapLayout(fyne.NewSize(float32(w), rowHeight)),
		container.New(layout.NewStackLayout(), v), // make v fill the fixed cell
	)
}

func flexVCenterTruncated(text string) fyne.CanvasObject {
	l := widget.NewLabel(text)
	l.Wrapping = fyne.TextWrapOff
	l.Truncation = fyne.TextTruncateEllipsis
	l.Alignment = fyne.TextAlignLeading

	v := container.NewVBox(layout.NewSpacer(), l, layout.NewSpacer())
	// Fill all available center space in Border using StackLayout:
	return container.New(layout.NewStackLayout(), v)
}

// tinyButton centers the icon button inside a fixed cell.
// Return both: the actual button (to mutate later) and the wrapped cell for layout.
func smallButton(icon fyne.Resource, onTap func()) (*widget.Button, fyne.CanvasObject) {
	btn := widget.NewButtonWithIcon("", icon, onTap)

	centered := container.NewCenter(btn)
	cell := container.New(
		layout.NewGridWrapLayout(fyne.NewSize(colPlayButtonWidth, rowHeight)),
		centered,
	)
	return btn, cell
}

// button becomes orange, pause icon displayed
func showRunning(button *widget.Button) {
	fyne.Do(func() {
		button.Importance = widget.WarningImportance // Importance change needs a Refresh()
		button.SetIcon(theme.MediaPauseIcon())       // SetIcon calls Refresh
	})
}

// button becomes grey, play icon displayed
func showStopped(button *widget.Button) {
	fyne.Do(func() {
		button.Importance = widget.MediumImportance // Importance change needs a Refresh()
		button.SetIcon(theme.MediaPlayIcon())       // SetIcon calls Refresh
	})
}
