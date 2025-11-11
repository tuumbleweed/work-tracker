package trackerapp

import (
	"fmt"
	"image/color"
	"os/exec"
	"path/filepath"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"golang.org/x/exp/constraints"
)

// returns "DD-MM-YYYY" for t
func dateID(t time.Time) string {
	return t.Format("02-01-2006")
}

func dayFilePath(workDir, dateId string) string {
	return filepath.Join(workDir, dateId+".jsonl")
}

func formatDuration(d time.Duration) string {
	if d < 0 {
		d = 0
	}
	sec := int(d.Seconds())
	h := sec / 3600
	m := (sec % 3600) / 60
	s := sec % 60
	return fmt.Sprintf("%02d:%02d:%02d", h, m, s)
}

func getActivityPercentage[T int64 | time.Duration](active, total T) float64 {
	var zero T
	if total == zero {
		return 0
	}
	active = Clamp(active, 0, total)
	return (float64(active) / float64(total)) * 100
}

// tryXprintidle returns idle ms if xprintidle works, else -1
func tryXprintidle() int64 {
	out, err := exec.Command("xprintidle").Output()
	if err != nil {
		return -1
	}
	var ms int64
	_, err = fmt.Sscanf(string(out), "%d", &ms)
	if err != nil {
		return -1
	}
	return ms
}

func Min[T constraints.Ordered](a, b T) T {
	if a < b {
		return a
	}
	return b
}

// Clamp returns v clamped between min and max (inclusive).
func Clamp[T constraints.Ordered](v, min, max T) T {
	if v < min {
		return min
	}
	if v > max {
		return max
	}
	return v
}

func vgap(wpx, hpx float32) fyne.CanvasObject {
	r := canvas.NewRectangle(color.NRGBA{0, 0, 0, 0}) // transparent
	r.SetMinSize(fyne.NewSize(wpx, hpx))
	return r
}

func clamp01(x float64) float64 {
	if x < 0 {
		return 0
	}
	if x > 1 {
		return 1
	}
	return x
}

func lerp(a, b, t float64) float64 { return a + (b-a)*t }

func lerpColor(c1, c2 color.NRGBA, t float64) color.NRGBA {
	return color.NRGBA{
		R: uint8(lerp(float64(c1.R), float64(c2.R), t) + 0.5),
		G: uint8(lerp(float64(c1.G), float64(c2.G), t) + 0.5),
		B: uint8(lerp(float64(c1.B), float64(c2.B), t) + 0.5),
		A: uint8(lerp(float64(c1.A), float64(c2.A), t) + 0.5),
	}
}

// 0–50%: red→yellow, 50–75%: yellow→green, 75–100%: blueBase→bluePeak (no teal)
func barColorFor(p float64) color.Color {
	red := color.NRGBA{R: 220, G: 60, B: 60, A: 255}     // 0%
	yellow := color.NRGBA{R: 235, G: 190, B: 50, A: 255} // 50%
	green := color.NRGBA{R: 60, G: 180, B: 90, A: 255}   // 75%
	// green peak at 100% (same hue family, different brightness/sat)
	greenPeak := color.NRGBA{R: 20, G: 180, B: 45, A: 255} // deeper/richer toward 100%

	t := clamp01(p / 100.0)

	switch {
	case t <= 0.5:
		// 0..50%: red -> yellow
		return lerpColor(red, yellow, t*2.0)
	case t <= 0.75:
		// 50..75%: yellow -> green
		return lerpColor(yellow, green, (t-0.5)*4.0)
	default:
		// 75..100%: start at clearly blue immediately, then deepen the blue
		u := (t - 0.75) * 4.0 // map [0.75,1] → [0,1]
		return lerpColor(green, greenPeak, u)
	}
}
