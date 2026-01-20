package report

import (
	"fmt"
	"image/color"
	"math"
	"path/filepath"
	"strings"
	"time"
)

/*
Clamp helpers + lerp.
*/
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

/*
Activity color ramp: 0%→red, 50%→yellow, 75%→green, 100%→greenPeak (your ramp).
We return #RRGGBB for HTML.
*/
func barColorFor(p float64) color.Color {
	red := color.NRGBA{R: 220, G: 60, B: 60, A: 255}       // 0%
	yellow := color.NRGBA{R: 235, G: 190, B: 50, A: 255}   // 50%
	green := color.NRGBA{R: 60, G: 180, B: 90, A: 255}     // 75%
	greenPeak := color.NRGBA{R: 20, G: 180, B: 45, A: 255} // 100%

	t := clamp01(p / 100.0)
	switch {
	case t <= 0.5:
		return lerpColor(red, yellow, t*2.0)
	case t <= 0.75:
		return lerpColor(yellow, green, (t-0.5)*4.0)
	default:
		u := (t - 0.75) * 4.0
		return lerpColor(green, greenPeak, u)
	}
}
func colorToHex(c color.Color) string {
	r, g, b, _ := c.RGBA()
	return fmt.Sprintf("#%02X%02X%02X", uint8(r>>8), uint8(g>>8), uint8(b>>8))
}


/*
Human-friendly duration like "1h 2m 3s".
*/
func formatDuration(d time.Duration) string {
	if d <= 0 {
		return "0s"
	}
	secs := int64(d.Seconds() + 0.5) // round
	h := secs / 3600
	m := (secs % 3600) / 60
	s := secs % 60
	out := &strings.Builder{}
	if h > 0 {
		fmt.Fprintf(out, "%dh ", h)
	}
	if m > 0 {
		fmt.Fprintf(out, "%dm ", m)
	}
	if s > 0 && h == 0 {
		fmt.Fprintf(out, "%ds", s)
	}
	return strings.TrimSpace(out.String())
}

/*
Parse "DD-MM-YYYY" in a given location (00:00 that day).
*/
func parseDMY(s string, loc *time.Location) (time.Time, error) {
	t, err := time.ParseInLocation("02-01-2006", s, loc)
	if err != nil {
		return time.Time{}, err
	}
	return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, loc), nil
}

/*
Compute Monday..Sunday of the current week in the given location.
*/
func currentWeekRange(loc *time.Location) (time.Time, time.Time) {
	now := time.Now().In(loc)
	base := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, loc)
	wd := int(base.Weekday())
	if wd == 0 {
		wd = 7 // Sunday->7
	}
	monday := base.AddDate(0, 0, -(wd-1))
	sunday := monday.AddDate(0, 0, 6)
	return monday, sunday
}



/*
Day-of-week short label.
*/
func weekdayShort(t time.Time) string { return t.Weekday().String()[:3] }

/*
Enumerate dates.
*/
func enumerateDates(start, end time.Time) []time.Time {
	var dates []time.Time
	for d := start; !d.After(end); d = d.AddDate(0, 0, 1) {
		dates = append(dates, d)
	}
	return dates
}

/*
File name "DD-MM-YYYY.jsonl".
*/
func dayFilePath(dir string, date time.Time) string {
	return filepath.Join(dir, date.Format("02-01-2006")+".jsonl")
}


// Gmail-safe "10 squares" indicator.
// percent is 0..100; filled squares use fillHex, empty use #e6e6e6.
// Squares are flat (no border-radius) as requested.
func buildSquares10HTML(percent float64, fillHex string) string {
	if percent < 0 {
		percent = 0
	}
	if percent > 100 {
		percent = 100
	}
	filled := int(math.Floor(percent / 10.0))
	if filled < 0 {
		filled = 0
	}
	if filled > 10 {
		filled = 10
	}

	var b strings.Builder
	b.WriteString(`<div>`)
	for i := range 10 {
		col := "#e6e6e6"
		if i < filled {
			col = fillHex
		}
		// no right margin on the last square
		if i == 9 {
			fmt.Fprintf(&b, `<span style="display:inline-block;width:12px;height:12px;background:%s;"></span>`, col)
		} else {
			fmt.Fprintf(&b, `<span style="display:inline-block;width:12px;height:12px;background:%s;margin-right:4px;"></span>`, col)
		}
	}
	b.WriteString(`</div>`)
	return b.String()
}
