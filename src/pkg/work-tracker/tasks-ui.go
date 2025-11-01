package worktracker

import (
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

// column widths (px) – tweak to taste
const (
	colPlayButtonWidth  = 32
	colNameWidth        = 260
	colDescriptionWidth = 420
	colCreatedAtWidth   = 180
	colHoursWidth       = 90
	// single-line row height
	rowHeight = 50
)

func (t *TrackerApp) makeTasksUI(tasks []Task) *fyne.Container {
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
		fixedCell(labelHeader("Created"), colCreatedAtWidth),
		fixedCell(labelHeader("Hours"), colHoursWidth),
	)
	descHead := minWidth(labelHeader("Description"), colDescriptionWidth) // e.g. colDescriptionWidth px minimum
	header := container.NewBorder(nil, nil, leftHeader, rightHeader, descHead)

	// rows
	rows := container.NewVBox()
	for _, task := range tasks {
		// left group: ▶ + Task (both fixed widths)
		playButton := smallButton(theme.MediaPlayIcon(), t.onButtonTapped, colPlayButtonWidth, rowHeight)
		leftBox := container.NewHBox(
			playButton,
			fixedCellCenteredTruncated(task.Name, colNameWidth),
		)

		// center: Description (expands; ellipsis)
		description := flexVCenterTruncated(task.Description)

		// right group: Created + Hours (both fixed)
		right := container.NewHBox(
			fixedCellCenteredTruncated(task.CreatedAt, colCreatedAtWidth),
			fixedCellCenteredTruncated("0.0 h", colHoursWidth),
		)

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
func fixedCellCenteredTruncated(text string, w int) fyne.CanvasObject {
	l := widget.NewLabel(text)
	l.Wrapping = fyne.TextWrapOff        // no wrapping
	l.Truncation = fyne.TextTruncateEllipsis // "…" when too long
	l.Alignment = fyne.TextAlignLeading  // left aligned

	// VBox with spacers => vertical center; MaxLayout => child gets full cell size
	v := container.NewVBox(layout.NewSpacer(), l, layout.NewSpacer())
	return container.New(
		layout.NewGridWrapLayout(fyne.NewSize(float32(w), rowHeight)),
		container.New(layout.NewStackLayout(), v), // make v fill the fixed cell
	)
}

func flexVCenterTruncated(text string) fyne.CanvasObject {
	l := widget.NewLabel(text)
	l.Wrapping   = fyne.TextWrapOff
	l.Truncation = fyne.TextTruncateEllipsis
	l.Alignment  = fyne.TextAlignLeading

	v := container.NewVBox(layout.NewSpacer(), l, layout.NewSpacer())
	// Fill all available center space in Border using StackLayout:
	return container.New(layout.NewStackLayout(), v)
}

// tinyButton centers the icon button inside a fixed cell.
func smallButton(icon fyne.Resource, onTap func(), width, height float32) fyne.CanvasObject {
	btn := widget.NewButtonWithIcon("", icon, onTap)
	// Optional: make it visually lighter
	// btn.Importance = widget.LowImportance

	centered := container.NewCenter(btn)
	return container.New(
		layout.NewGridWrapLayout(fyne.NewSize(width, height)),
		centered, // centered inside the play button cell
	)
}
