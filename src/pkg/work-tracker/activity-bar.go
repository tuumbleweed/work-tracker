package worktracker

import (
	"fmt"
	"image/color"
	"math"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

type ActivityBar struct {
	widget.BaseWidget

	Caption    string  // e.g., "Average activity"
	percent    float64 // 0..100
	col        color.Color
	WidthRatio float32 // fraction of available width to use for the bar (0..1), e.g. 0.8 for 80%
}

func NewActivityBar(caption string) *ActivityBar {
	ab := &ActivityBar{
		Caption:    caption,
		percent:    0,
		col:        color.NRGBA{R: 200, G: 40, B: 40, A: 255}, // start red-ish
		WidthRatio: 0.5,                                       // default to 80%
	}
	ab.ExtendBaseWidget(ab)
	return ab
}

func (a *ActivityBar) SetPercent(p float64) {
	if p < 0 {
		p = 0
	}
	if p > 100 {
		p = 100
	}
	a.percent = p
	a.col = barColorFor(p)
	a.Refresh()
}

func (a *ActivityBar) SetWidthRatio(r float32) {
	if r < 0 {
		r = 0
	}
	if r > 1 {
		r = 1
	}
	a.WidthRatio = r
	a.Refresh()
}

func (a *ActivityBar) Percent() float64 { return a.percent }

// --- widget.Renderer ---

type activityBarRenderer struct {
	a        *ActivityBar
	caption  *canvas.Text
	bg       *canvas.Rectangle
	fill     *canvas.Rectangle
	percentT *canvas.Text
	objects  []fyne.CanvasObject
}

func (a *ActivityBar) CreateRenderer() fyne.WidgetRenderer {
	cap := canvas.NewText(a.Caption, theme.Color(theme.ColorNameForeground))
	cap.Alignment = fyne.TextAlignCenter
	cap.TextSize = theme.TextSize()

	bg := canvas.NewRectangle(theme.Color(theme.ColorNameInputBackground))
	fill := canvas.NewRectangle(a.col)

	txt := canvas.NewText(fmt.Sprintf("%.1f%%", a.percent), theme.Color(theme.ColorNameForeground))
	txt.Alignment = fyne.TextAlignCenter
	txt.TextSize = theme.TextSize()

	r := &activityBarRenderer{
		a:        a,
		caption:  cap,
		bg:       bg,
		fill:     fill,
		percentT: txt,
	}
	r.objects = []fyne.CanvasObject{cap, bg, fill, txt}
	return r
}

func (r *activityBarRenderer) Layout(sz fyne.Size) {
	// Reserve caption height
	capH := r.caption.MinSize().Height
	r.caption.Move(fyne.NewPos(0, 0))
	r.caption.Resize(fyne.NewSize(sz.Width, capH))

	// Bar area below caption
	barY := capH + theme.Padding()/2
	barH := float32(math.Max(24, float64(theme.TextSize())*1.6))

	// Compute inner (shorter) width and center it
	innerW := float32(float64(sz.Width) * float64(r.a.WidthRatio))
	if innerW < 0 {
		innerW = 0
	}
	if innerW > sz.Width {
		innerW = sz.Width
	}
	innerX := (sz.Width - innerW) / 2

	// Background bar
	r.bg.Move(fyne.NewPos(innerX, barY))
	r.bg.Resize(fyne.NewSize(innerW, barH))

	// Fill width by percent
	fillW := float32(float64(innerW) * (r.a.percent / 100.0))
	r.fill.FillColor = r.a.col
	r.fill.Move(fyne.NewPos(innerX, barY))
	r.fill.Resize(fyne.NewSize(fillW, barH))

	// Centered percentage text over the bar (align to the inner bar width)
	r.percentT.Text = fmt.Sprintf("%.1f%%", r.a.percent)
	r.percentT.Move(fyne.NewPos(innerX, barY+(barH-r.percentT.MinSize().Height)/2))
	r.percentT.Resize(fyne.NewSize(innerW, r.percentT.MinSize().Height))
}

func (r *activityBarRenderer) MinSize() fyne.Size {
	// Caption + bar + padding
	h := r.caption.MinSize().Height + theme.Padding()/2 + float32(math.Max(24, float64(theme.TextSize())*1.6))
	w := float32(220) // sensible minimum width
	return fyne.NewSize(w, h)
}

func (r *activityBarRenderer) Refresh() {
	r.Layout(r.a.Size())
	canvas.Refresh(r.a)
}

func (r *activityBarRenderer) Destroy()                     {}
func (r *activityBarRenderer) Objects() []fyne.CanvasObject { return r.objects }
