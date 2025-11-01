// keep this file self-contained for now; we'll split into a package later.
package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"image/color"
	"math"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"time"

	er "work-tracker/src/pkg/error"
	"work-tracker/src/pkg/config"
	"work-tracker/src/pkg/logger"
	"work-tracker/src/pkg/util"

	"hash/fnv"
)

/*
JSONL input line from work-tracker.

We allow ActiveTime to come either as a JSON number (nanoseconds) or as a string
(parseable by time.ParseDuration, e.g. "999ms", "1.23s").
*/
type jsonDuration struct{ time.Duration }

func (d *jsonDuration) UnmarshalJSON(b []byte) error {
	s := strings.TrimSpace(string(b))
	// number => nanoseconds
	if len(s) > 0 && (s[0] == '-' || (s[0] >= '0' && s[0] <= '9')) && !strings.ContainsAny(s, `"abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ`) {
		var n int64
		dec := json.NewDecoder(strings.NewReader(s))
		err := dec.Decode(&n)
		if err != nil {
			return err
		}
		d.Duration = time.Duration(n)
		return nil
	}
	// string => parse duration
	var str string
	err := json.Unmarshal(b, &str)
	if err != nil {
		return err
	}
	parsed, perr := time.ParseDuration(str)
	if perr != nil {
		// try to parse bare integer string as nanoseconds
		strTrim := strings.TrimSpace(str)
		if strTrim != "" && strTrim[0] >= '0' && strTrim[0] <= '9' {
			var n2 int64
			dec := json.NewDecoder(strings.NewReader(strTrim))
			e2 := dec.Decode(&n2)
			if e2 == nil {
				d.Duration = time.Duration(n2)
				return nil
			}
		}
		return perr
	}
	d.Duration = parsed
	return nil
}

type Chunk struct {
	TaskName   string       `json:"task_name"`
	StartedAt  time.Time    `json:"started_at"`
	FinishedAt time.Time    `json:"finished_at"`
	ActiveTime jsonDuration `json:"active_time"`
}

/*
Per-day aggregation used for charts.
*/
type DaySummary struct {
	Date               time.Time                `json:"date"`
	TotalDuration      time.Duration            `json:"total_duration"`
	TotalActive        time.Duration            `json:"total_active"`
	TaskDurations      map[string]time.Duration `json:"task_durations"`
	SmoothedActiveTime time.Duration            `json:"smoothed_active_time"` // Σ (duration * smooth(active_ratio))
}

/*
Top-level aggregation across the whole selected range.
*/
type ReportTotals struct {
	TotalWorked   time.Duration
	TotalActive   time.Duration
	PerTaskTotals map[string]time.Duration
	TaskOrder     []string
}

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
Categorical palette with wide-separated hue bands (12+ families).
Mapping is deterministic per task name, but ensures bands differ a lot.
*/
type hueBand struct {
	hMin float64
	hMax float64
	s    float64
	l    float64
}

var paletteBands = []hueBand{
	{210, 230, 0.70, 0.52}, // blue
	{35, 45, 0.85, 0.50},   // amber
	{95, 110, 0.70, 0.48},  // lime
	{270, 290, 0.60, 0.52}, // purple
	{310, 330, 0.65, 0.52}, // magenta
	{50, 60, 0.90, 0.46},   // yellow
	{335, 350, 0.65, 0.52}, // pink
	{170, 185, 0.65, 0.50}, // teal
	{120, 135, 0.65, 0.50}, // green
	{190, 205, 0.70, 0.48}, // cyan
	{235, 255, 0.65, 0.52}, // indigo
	{20, 30, 0.80, 0.50},   // orange
}

func taskColorHex(id int, task string) string {
	h := fnv.New32a()
	_, _ = h.Write([]byte(task))
	hash := h.Sum32()

	bandIdx := int(uint32(id) % uint32(len(paletteBands)))
	band := paletteBands[bandIdx]

	// vary hue inside band
	inner := float64((hash>>8)%1000) / 1000.0 // 0..1
	hue := band.hMin + inner*(band.hMax-band.hMin)

	// vary lightness slightly across 3 steps
	lightSteps := []float64{-0.07, 0.0, +0.06}
	li := int((hash>>18)%uint32(len(lightSteps)))
	light := clamp01(band.l + lightSteps[li])

	r, g, b := hslToRGB(hue/360.0, band.s, light)
	return fmt.Sprintf("#%02X%02X%02X", r, g, b)
}

// HSL -> RGB helpers
func hslToRGB(h, s, l float64) (uint8, uint8, uint8) {
	if s == 0 {
		v := uint8(l * 255.0)
		return v, v, v
	}
	var q float64
	if l < 0.5 {
		q = l * (1 + s)
	} else {
		q = l + s - l*s
	}
	p := 2*l - q
	r := hue2rgb(p, q, h+1.0/3.0)
	g := hue2rgb(p, q, h)
	b := hue2rgb(p, q, h-1.0/3.0)
	return uint8(r*255.0 + 0.5), uint8(g*255.0 + 0.5), uint8(b*255.0 + 0.5)
}
func hue2rgb(p, q, t float64) float64 {
	if t < 0 {
		t += 1
	}
	if t > 1 {
		t -= 1
	}
	if t < 1.0/6.0 {
		return p + (q-p)*6*t
	}
	if t < 1.0/2.0 {
		return q
	}
	if t < 2.0/3.0 {
		return p + (q-p)*(2.0/3.0-t)*6
	}
	return p
}

/*
Smooth activity factor f∈[0,1] by exponent α = 1 - smooth (smooth∈[0,1]).
*/
func smoothFactor(f, smooth float64) float64 {
	if f <= 0 {
		return 0
	}
	if f >= 1 {
		return 1
	}
	alpha := 1.0 - smooth
	if alpha < 0.2 {
		alpha = 0.2
	}
	if alpha > 1.0 {
		alpha = 1.0
	}
	return math.Pow(f, alpha)
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
Read a single day file into a DaySummary. Missing file => empty summary (no error).
*/
func readDayFile(filePath string, date time.Time, smooth float64) (sum DaySummary, e *er.Error) {
	sum = DaySummary{
		Date:               date,
		TaskDurations:      make(map[string]time.Duration),
		TotalDuration:      0,
		TotalActive:        0,
		SmoothedActiveTime: 0,
	}

	_, statErr := os.Stat(filePath)
	if statErr != nil {
		if os.IsNotExist(statErr) {
			logger.Log(logger.Notice, logger.CyanColor, "%s missing day file '%s' (treated as 0)", "Skipping", filePath)
			return sum, nil
		}
		e = er.NewErrorECOL(statErr, "unable to stat day file", "path", filePath)
		return sum, e
	}

	f, openErr := os.Open(filePath)
	if openErr != nil {
		e = er.NewErrorECOL(openErr, "unable to open day file", "path", filePath)
		return sum, e
	}
	defer f.Close()

	sc := bufio.NewScanner(f)
	buf := make([]byte, 0, 64*1024)
	sc.Buffer(buf, 2*1024*1024)

	lineNumber := 0
	for sc.Scan() {
		line := strings.TrimSpace(sc.Text())
		lineNumber++
		if line == "" {
			continue
		}
		var ch Chunk
		uErr := json.Unmarshal([]byte(line), &ch)
		if uErr != nil {
			logger.Log(logger.Notice, logger.PurpleColor, "%s malformed JSON in '%s' line %d", "Skipping", filePath, lineNumber)
			continue
		}
		if !ch.FinishedAt.After(ch.StartedAt) {
			logger.Log(logger.Notice, logger.PurpleColor, "%s bad chunk time window in '%s' line %d", "Skipping", filePath, lineNumber)
			continue
		}
		dur := ch.FinishedAt.Sub(ch.StartedAt)
		active := ch.ActiveTime.Duration
		if active < 0 {
			active = 0
		}
		if active > dur {
			active = dur
		}
		sum.TotalDuration += dur
		sum.TotalActive += active

		task := ch.TaskName
		if strings.TrimSpace(task) == "" {
			task = "Unassigned Time"
		}
		sum.TaskDurations[task] += dur

		ratio := 0.0
		if dur > 0 {
			ratio = float64(active) / float64(dur)
		}
		sm := smoothFactor(ratio, smooth)
		sum.SmoothedActiveTime += time.Duration(float64(dur) * sm)
	}
	sErr := sc.Err()
	if sErr != nil {
		e = er.NewErrorECML(sErr, "scanner error while reading day file", "line",
			map[string]any{"path": filePath, "last_line": lineNumber})
		return sum, e
	}
	return sum, nil
}

/*
Build SVG radial progress ring (inline, Gmail-friendly).
*/
func buildRadialProgressSVG(size int, stroke int, percent float64, colorHex string) string {
	if percent < 0 {
		percent = 0
	}
	if percent > 100 {
		percent = 100
	}
	r := float64(size-stroke) / 2.0
	circ := 2.0 * math.Pi * r
	fill := circ * percent / 100.0
	offset := circ - fill
	cx := float64(size) / 2.0
	cy := float64(size) / 2.0
	font := int(float64(size) * 0.28)
	return fmt.Sprintf(
		`<svg width="%d" height="%d" viewBox="0 0 %d %d" xmlns="http://www.w3.org/2000/svg" style="display:block;">
  <circle cx="%.1f" cy="%.1f" r="%.1f" fill="none" stroke="#e6e6e6" stroke-width="%d"/>
  <circle cx="%.1f" cy="%.1f" r="%.1f" fill="none" stroke="%s" stroke-width="%d"
          stroke-linecap="round"
          stroke-dasharray="%.1f"
          stroke-dashoffset="%.1f"
          transform="rotate(-90 %.1f %.1f)"/>
  <text x="50%%" y="50%%" dominant-baseline="middle" text-anchor="middle" font-family="Arial, sans-serif" font-size="%d" fill="#222">%.1f%%</text>
</svg>`,
		size, size, size, size,
		cx, cy, r, stroke,
		cx, cy, r, colorHex, stroke,
		circ, offset, cx, cy,
		font, percent,
	)
}

/*
Day-of-week short label.
*/
func weekdayShort(t time.Time) string { return t.Weekday().String()[:3] }

/*
HTML helpers.
*/
func esc(s string) string {
	replacer := strings.NewReplacer(
		`&`, "&amp;",
		`<`, "&lt;",
		`>`, "&gt;",
		`"`, "&quot;",
		`'`, "&#39;",
	)
	return replacer.Replace(s)
}

/*
Format "Weekly Report — 25 Oct 2025 – 31 Oct 2025".
*/
func rangeTitle(start, end time.Time) string {
	if start.Equal(end) {
		return fmt.Sprintf("Weekly Report — %s", start.Format("02 Jan 2006"))
	}
	if start.Year() == end.Year() && start.Month() == end.Month() {
		return fmt.Sprintf("Weekly Report — %s – %s %d", start.Format("02"), end.Format("02 Jan"), end.Year())
	}
	if start.Year() == end.Year() {
		return fmt.Sprintf("Weekly Report — %s – %s %d", start.Format("02 Jan"), end.Format("02 Jan"), end.Year())
	}
	return fmt.Sprintf("Weekly Report — %s – %s", start.Format("02 Jan 2006"), end.Format("02 Jan 2006"))
}

/*
Render the entire HTML report into a buffer.

barRef      -> target duration label (e.g., 12m).
barHeightPx -> pixel height that corresponds to barRef (used to scale bars).
*/
func renderHTMLReport(buf *bytes.Buffer, daySummaries []DaySummary, totals ReportTotals, barRef time.Duration, barHeightPx int, startDate, endDate time.Time) {
	// ---------- precompute ----------
	refSeconds := barRef.Seconds()
	if refSeconds <= 0 {
		refSeconds = 1 // avoid div-by-zero; degenerate but safe
	}
	pxPerSecond := float64(barHeightPx) / refSeconds

	// helper: convert a duration to pixel height (at least 1px if positive)
	segHeight := func(d time.Duration) int {
		if d <= 0 {
			return 0
		}
		h := int(math.Round(d.Seconds() * pxPerSecond))
		if h == 0 {
			h = 1
		}
		return h
	}

	// Precompute per-day container heights and row max for "Time by Day"
	dayContainerHeights := make([]int, len(daySummaries))
	maxDayRowHeight := barHeightPx
	for i, ds := range daySummaries {
		h := segHeight(ds.TotalDuration)
		container := h
		if container < barHeightPx {
			container = barHeightPx
		}
		dayContainerHeights[i] = container
		if container > maxDayRowHeight {
			maxDayRowHeight = container
		}
	}

	// Precompute per-day container heights and row max for "Activity×Time"
	smContainerHeights := make([]int, len(daySummaries))
	maxSmRowHeight := barHeightPx
	for i, ds := range daySummaries {
		h := segHeight(ds.SmoothedActiveTime)
		container := h
		if container < barHeightPx {
			container = barHeightPx
		}
		smContainerHeights[i] = container
		if container > maxSmRowHeight {
			maxSmRowHeight = container
		}
	}

	// task listing (sorted by total desc)
	taskNames := make([]string, 0, len(totals.PerTaskTotals))
	for k := range totals.PerTaskTotals {
		taskNames = append(taskNames, k)
	}
	sort.Slice(taskNames, func(i, j int) bool {
		di := totals.PerTaskTotals[taskNames[i]]
		dj := totals.PerTaskTotals[taskNames[j]]
		if di == dj {
			return taskNames[i] < taskNames[j]
		}
		return di > dj
	})

	avgActivity := 0.0
	if totals.TotalWorked > 0 {
		avgActivity = (float64(totals.TotalActive) / float64(totals.TotalWorked)) * 100.0
	}
	activityHex := colorToHex(barColorFor(avgActivity))

	// Activity color legend swatches
	hex0 := colorToHex(barColorFor(0))
	hex50 := colorToHex(barColorFor(50))
	hex75 := colorToHex(barColorFor(75))
	hex100 := colorToHex(barColorFor(100))

	// Chart geometry for centering (per-day width = bar width + left/right padding)
	const barW = 28
	const pad = 8
	perDayW := barW + 2*pad
	chartW := len(daySummaries) * perDayW // used to center the inner table

	// ---------- HTML ----------
	fmt.Fprintf(buf, `<!doctype html>
<html>
  <head>
    <meta charset="utf-8">
    <title>%s</title>
  </head>
  <body style="margin:0;padding:0;background:#fafafa;">
    <!-- Wrapper -->
    <table role="presentation" width="100%%" cellpadding="0" cellspacing="0" style="max-width:760px;margin:0 auto;background:#fff;border-collapse:collapse;">
      <tr>
        <td style="font-family:Arial, sans-serif;color:#222;font-size:16px;padding:12px 18px;border-bottom:1px solid #eee;">
          %s
        </td>
      </tr>

      <!-- Summary Section -->
      <tr>
        <td align="center" style="padding:14px 8px;">
          <table role="presentation" cellpadding="0" cellspacing="0" style="border-collapse:collapse;">
            <tr valign="middle">
              <td align="center" style="padding:0 16px;">
                <div style="font-family:Arial, sans-serif;font-size:13px;color:#666;padding-bottom:4px;">Total Worked</div>
                <div style="font-family:Arial, sans-serif;font-size:28px;color:#111;font-weight:bold;">%s</div>
              </td>
              <td align="center" style="padding:0 16px;">
                <div style="font-family:Arial, sans-serif;font-size:13px;color:#666;padding-bottom:4px;">Avg Activity</div>
                %s
              </td>
            </tr>
          </table>
        </td>
      </tr>

      <!-- Tasks in period (vertical list, centered) -->
      <tr>
        <td align="center" style="padding:4px 12px 10px 12px;">
          <div style="font-family:Arial, sans-serif;font-size:14px;color:#444;padding-bottom:6px;">Tasks in period</div>
          <table role="presentation" cellpadding="0" cellspacing="0" style="border-collapse:collapse;">
`, rangeTitle(startDate, endDate), rangeTitle(startDate, endDate), formatDuration(totals.TotalWorked), buildRadialProgressSVG(96, 10, avgActivity, activityHex))

	for i, name := range taskNames {
		dur := totals.PerTaskTotals[name]
		c := taskColorHex(i, name)
		fmt.Fprintf(buf, `            <tr>
              <td style="padding:4px 10px;font-family:Arial, sans-serif;font-size:13px;color:#333;vertical-align:middle;text-align:center;">
                <table role="presentation" cellpadding="0" cellspacing="0" style="border-collapse:collapse;margin:0 auto;">
                  <tr>
                    <td style="background:%s;width:12px;height:12px;line-height:0;font-size:0;">&nbsp;</td>
                    <td style="padding-left:8px;">%s&nbsp;<span style="color:#666;">— %s</span></td>
                  </tr>
                </table>
              </td>
            </tr>
`, c, esc(name), esc(formatDuration(dur)))
	}

	fmt.Fprintf(buf, `          </table>
        </td>
      </tr>

      <!-- Time by Day (stacked per task) -->
      <tr>
        <td align="center" style="padding:15px 0 10px 0;">
          <div style="font-family:Arial, sans-serif;color:#222;font-size:14px;">Time by Day (%s baseline)</div>
        </td>
      </tr>

      <!-- Centered chart wrapper with top-left label INSIDE (doesn't affect centering) -->
      <tr>
        <td align="center" style="padding:2px 0 0 0;">
          <table role="presentation" cellpadding="0" cellspacing="0" width="%d" style="border-collapse:collapse;">
            <tr>
              <td>
                <table role="presentation" cellpadding="0" cellspacing="0" style="border-collapse:collapse;margin:0 auto;">
                  <tr valign="bottom">
`, esc(formatDuration(barRef)), chartW)

	for i, dsum := range daySummaries {
		containerH := dayContainerHeights[i]
		dayTotalH := segHeight(dsum.TotalDuration)
		if dayTotalH > containerH {
			dayTotalH = containerH
		}
		topSpacer := containerH - dayTotalH
		if topSpacer < 0 {
			topSpacer = 0
		}

		// order tasks by global order, include only present tasks
		dayTasks := make([]string, 0, len(taskNames))
		for _, tname := range taskNames {
			if dsum.TaskDurations[tname] > 0 {
				dayTasks = append(dayTasks, tname)
			}
		}

		fmt.Fprintf(buf, `                    <!-- Day bar -->
                    <td align="center" style="padding:0 %dpx;">
                      <table role="presentation" cellpadding="0" cellspacing="0" style="border-collapse:collapse;">
                        <tr><td style="background:#e0e0e0;width:%dpx;height:%dpx;line-height:0;font-size:0;">
                          <table role="presentation" cellpadding="0" cellspacing="0" style="border-collapse:collapse;width:%dpx;">
`, pad, barW, containerH, barW)

		if topSpacer > 0 {
			fmt.Fprintf(buf, `                            <tr><td style="height:%dpx;line-height:0;font-size:0;">&nbsp;</td></tr>
`, topSpacer)
		}
		for i, tname := range dayTasks {
			seg := segHeight(dsum.TaskDurations[tname])
			if seg <= 0 {
				continue
			}
			fmt.Fprintf(buf, `                            <tr><td style="background:%s;height:%dpx;line-height:0;font-size:0;">&nbsp;</td></tr>
`, taskColorHex(i, tname), seg)
		}

		fmt.Fprintf(buf, `                          </table>
                        </td></tr>
                      </table>
                      <div style="font-family:Arial, sans-serif;font-size:12px;color:#555;padding-top:6px;">%s</div>
                    </td>
`, esc(weekdayShort(dsum.Date)))
	}

	fmt.Fprintf(buf, `                  </tr>
                </table>
              </td>
            </tr>
          </table>
        </td>
      </tr>

      <!-- Legend for tasks -->
      <tr>
        <td align="center" style="padding:6px 0 10px 0;">
          <table role="presentation" cellpadding="0" cellspacing="0" style="border-collapse:collapse;">
            <tr>
`)
	for idx, name := range taskNames {
		if idx > 0 {
			fmt.Fprintf(buf, `              <td style="width:12px;">&nbsp;</td>
`)
		}
		fmt.Fprintf(buf, `              <td style="background:%s;width:10px;height:10px;line-height:0;font-size:0;">&nbsp;</td>
              <td style="font-family:Arial, sans-serif;font-size:12px;color:#555;padding:0 0 0 6px;">%s</td>
`, taskColorHex(idx, name), esc(name))
	}
	fmt.Fprintf(buf, `            </tr>
          </table>
        </td>
      </tr>

      <!-- Activity × Time -->
      <tr>
        <td align="center" style="padding:15px 0 10px 0;">
          <div style="font-family:Arial, sans-serif;color:#222;font-size:14px;">Activity × Time</div>
        </td>
      </tr>

      <!-- Centered chart wrapper with top-left label for the smoothed chart -->
      <tr>
        <td align="center" style="padding:2px 0 6px 0;">
          <table role="presentation" cellpadding="0" cellspacing="0" width="%d" style="border-collapse:collapse;">
            <tr>
              <td>
                <table role="presentation" cellpadding="0" cellspacing="0" style="border-collapse:collapse;margin:0 auto;">
                  <tr valign="bottom">
`, chartW)

	for i, dsum := range daySummaries {
		dayPct := 0.0
		if dsum.TotalDuration > 0 {
			dayPct = (float64(dsum.TotalActive) / float64(dsum.TotalDuration)) * 100.0
		}
		hex := colorToHex(barColorFor(dayPct))

		containerH := smContainerHeights[i]
		h := segHeight(dsum.SmoothedActiveTime)
		if h > containerH {
			h = containerH
		}
		top := containerH - h
		if top < 0 {
			top = 0
		}

		fmt.Fprintf(buf, `                    <td align="center" style="padding:0 %dpx;">
                      <table role="presentation" cellpadding="0" cellspacing="0" style="border-collapse:collapse;">
                        <tr><td style="background:#e0e0e0;width:%dpx;height:%dpx;line-height:0;font-size:0;">
                          <table role="presentation" cellpadding="0" cellspacing="0" style="border-collapse:collapse;width:%dpx;">
                            <tr><td style="height:%dpx;line-height:0;font-size:0;">&nbsp;</td></tr>
                            <tr><td style="background:%s;height:%dpx;line-height:0;font-size:0;">&nbsp;</td></tr>
                          </table>
                        </td></tr>
                      </table>
                      <div style="font-family:Arial, sans-serif;font-size:12px;color:#555;padding-top:6px;">%s</div>
                    </td>
`, pad, barW, containerH, barW, top, hex, h, esc(weekdayShort(dsum.Date)))
	}

	fmt.Fprintf(buf, `                  </tr>
                </table>
              </td>
            </tr>
          </table>
        </td>
      </tr>

      <!-- Activity × Time legend -->
      <tr>
        <td align="center" style="padding:2px 0 20px 0;">
          <table role="presentation" cellpadding="0" cellspacing="0" style="border-collapse:collapse;">
            <tr>
              <td style="background:%[1]s;width:10px;height:10px;line-height:0;font-size:0;">&nbsp;</td>
              <td style="font-family:Arial, sans-serif;font-size:12px;color:#555;padding:0 12px 0 6px;">%[5]s × 0%%</td>

              <td style="background:%[2]s;width:10px;height:10px;line-height:0;font-size:0;">&nbsp;</td>
              <td style="font-family:Arial, sans-serif;font-size:12px;color:#555;padding:0 12px 0 6px;">%[5]s × 50%%</td>

              <td style="background:%[3]s;width:10px;height:10px;line-height:0;font-size:0;">&nbsp;</td>
              <td style="font-family:Arial, sans-serif;font-size:12px;color:#555;padding:0 12px 0 6px;">%[5]s × 75%%</td>

              <td style="background:%[4]s;width:10px;height:10px;line-height:0;font-size:0;">&nbsp;</td>
              <td style="font-family:Arial, sans-serif;font-size:12px;color:#555;padding:0 0 0 6px;">%[5]s × 100%%</td>
            </tr>
          </table>
        </td>
      </tr>

    </table> <!-- end white wrapper -->

	<!-- Footer OUTSIDE content area -->
    <table role="presentation" width="100%%" cellpadding="0" cellspacing="0" style="border-collapse:collapse;margin:0 auto;">
      <tr>
        <td align="center" style="padding:20px 0;">
          <div style="font-family:Arial, sans-serif;font-size:12px;color:#888;">
            Generated by Work Tracker
          </div>
        </td>
      </tr>
    </table>

  </body>
</html>
`, hex0, hex50, hex75, hex100, esc(formatDuration(barRef)))
}


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

/*
Build the report: read files, aggregate, render HTML, write to disk.
*/
func buildReport(inputDir string, startDate, endDate time.Time, outPath string, barRef time.Duration, smooth float64) (e *er.Error) {
	logger.Log(logger.Notice, logger.BlueColor, "%s files from '%s' for '%s'..'%s'", "Reading", inputDir, startDate.Format("02-01-2006"), endDate.Format("02-01-2006"))

	dates := enumerateDates(startDate, endDate)
	daySummaries := make([]DaySummary, 0, len(dates))
	totals := ReportTotals{
		PerTaskTotals: make(map[string]time.Duration),
	}

	for _, d := range dates {
		fp := dayFilePath(inputDir, d)
		sum, rerr := readDayFile(fp, d, smooth)
		if rerr != nil {
			return rerr
		}
		daySummaries = append(daySummaries, sum)

		totals.TotalWorked += sum.TotalDuration
		totals.TotalActive += sum.TotalActive
		for k, v := range sum.TaskDurations {
			totals.PerTaskTotals[k] += v
		}
	}

	for k := range totals.PerTaskTotals {
		totals.TaskOrder = append(totals.TaskOrder, k)
	}
	sort.Slice(totals.TaskOrder, func(i, j int) bool {
		di := totals.PerTaskTotals[totals.TaskOrder[i]]
		dj := totals.PerTaskTotals[totals.TaskOrder[j]]
		if di == dj {
			return totals.TaskOrder[i] < totals.TaskOrder[j]
		}
		return di > dj
	})

	var buf bytes.Buffer
	renderHTMLReport(&buf, daySummaries, totals, barRef, 200, startDate, endDate)

	err := os.WriteFile(outPath, buf.Bytes(), 0o644)
	if err != nil {
		e = er.NewErrorECOL(err, "failed to write HTML report", "path", outPath)
		return e
	}
	logger.Log(logger.Notice, logger.GreenColor, "%s report to '%s' (%s, %d days)", "Wrote", outPath, formatDuration(totals.TotalWorked), len(daySummaries))
	return nil
}

func main() {
	util.CheckIfEnvVarsPresent([]string{})

	// common flags
	logLevelOverride := flag.Int("log-level", -1, "Log level. Default is whatever value is in configuration file. Keep at -1 to not override.")
	logDirOverride := flag.String("log-dir", "", "File directory at which to save log files. Keep empty to use configuration file instead.")
	configPath := flag.String("config", "./cfg/config.json", "Path to your configuration file.")

	// program's custom flags
	flagStart := flag.String("start", "", "Start date (inclusive) in DD-MM-YYYY; empty => this Monday")
	flagEnd := flag.String("end", "", "End date (inclusive) in DD-MM-YYYY; empty => this Sunday (or start if start set)")
	flagInputDir := flag.String("dir", "./out", "Directory with day JSONL files")
	flagOutputPath := flag.String("output", "./out/report.html", "Path to write the HTML report")
	flagTZ := flag.String("tz", "America/Bogota", "IANA timezone for week boundaries and display")
	flagBarRef := flag.Duration("ref", 12*time.Hour, "Reference duration for the horizontal marker line (N hours)")
	flagSmooth := flag.Float64("smooth", 0.0, "Activity smoothing in [0..1], 0=linear, 1=strong")

	// parse and init config
	flag.Parse()
	config.InitializeConfig(*configPath, logger.LogLevel(*logLevelOverride), *logDirOverride)

	logger.Log(logger.Notice, logger.BoldBlueColor, "%s report entrypoint. Config path: '%s'", "Running", *configPath)

	// Resolve timezone
	loc, tzErr := time.LoadLocation(*flagTZ)
	if tzErr != nil {
		er.NewErrorECOL(tzErr, "failed to load timezone", "tz", *flagTZ).QuitIf("error")
	}

	// Determine date range
	var startDate, endDate time.Time
	var err error

	if *flagStart == "" && *flagEnd == "" {
		startDate, endDate = currentWeekRange(loc)
	} else if *flagStart != "" && *flagEnd == "" {
		startDate, err = parseDMY(*flagStart, loc)
		if err != nil {
			er.NewErrorECOL(err, "failed to parse start date", "start", *flagStart).QuitIf("error")
		}
		endDate = startDate
	} else if *flagStart == "" && *flagEnd != "" {
		endDate, err = parseDMY(*flagEnd, loc)
		if err != nil {
			er.NewErrorECOL(err, "failed to parse end date", "end", *flagEnd).QuitIf("error")
		}
		startDate = endDate
	} else {
		startDate, err = parseDMY(*flagStart, loc)
		if err != nil {
			er.NewErrorECOL(err, "failed to parse start date", "start", *flagStart).QuitIf("error")
		}
		endDate, err = parseDMY(*flagEnd, loc)
		if err != nil {
			er.NewErrorECOL(err, "failed to parse end date", "end", *flagEnd).QuitIf("error")
		}
	}

	// Build the report
	e := buildReport(*flagInputDir, startDate, endDate, *flagOutputPath, *flagBarRef, *flagSmooth)
	e.QuitIf("error")

	// Open in Chrome
	logger.Log(logger.Notice, logger.BoldBlueColor, "%s a file '%s' with %s", "Opening", *flagOutputPath, "google-chrome")
	err = exec.Command("google-chrome", *flagOutputPath).Start()
	er.QuitIfError(err, "Unable to open html report with google chrome")
}
