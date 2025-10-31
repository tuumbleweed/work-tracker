package worktracker

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

// tinyButton wraps an icon-only button into a fixed cell (e.g., 28x28).
func tinyButton(icon fyne.Resource, onTap func()) fyne.CanvasObject {
	btn := widget.NewButtonWithIcon("", icon, onTap)
	cell := container.New(layout.NewGridWrapLayout(fyne.NewSize(40, 40)), btn)
	return cell
}

// tinyHeaderCell keeps the first header column narrow, matching the play button size.
func tinyHeaderCell() fyne.CanvasObject {
	// could put " " or an icon legend here if you like
	lbl := widget.NewLabel("")
	return container.New(layout.NewGridWrapLayout(fyne.NewSize(40, 40)), lbl)
}

func (t *TrackerApp) makeTasksUI(tasks []Task) *fyne.Container {
	// Bigger, bold, centered title
	sectionTitle := canvas.NewText("Tasks", theme.Color(theme.ColorNameForeground))
	sectionTitle.Alignment = fyne.TextAlignCenter
	sectionTitle.TextStyle = fyne.TextStyle{Bold: true}
	sectionTitle.TextSize = theme.TextSize() * 1.6

	// Header: small left cell + 4 equal data columns
	headerRight := container.NewGridWithColumns(4,
		labelHeader("Task"),
		labelHeader("Description"),
		labelHeader("Created"),
		labelHeader("Overall Hours"),
	)
	header := container.NewBorder(nil, nil, tinyHeaderCell(), nil, headerRight)

	rows := container.NewVBox()
	for _, task := range tasks {
		// small play button on the left
		play := tinyButton(theme.MediaPlayIcon(), func() {
			t.onButtonTapped()
		})

		name := widget.NewLabel(task.Name)
		name.Wrapping = fyne.TextWrapWord

		desc := widget.NewLabel(task.Description)
		desc.Wrapping = fyne.TextWrapWord

		created := widget.NewLabel(task.CreatedAt)
		overall := widget.NewLabel("0.0 h") // dummy for now

		// Right side: 4 equal columns; left is the tiny play cell.
		right := container.NewGridWithColumns(4, name, desc, created, overall)
		row := container.NewBorder(nil, nil, play, nil, right)

		rows.Add(row)
	}

	return container.NewVBox(sectionTitle, header, rows)
}

func labelHeader(s string) *widget.Label {
	l := widget.NewLabelWithStyle(s, fyne.TextAlignLeading, fyne.TextStyle{Bold: true})
	l.Wrapping = fyne.TextWrapWord
	return l
}
