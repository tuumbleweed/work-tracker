// you can add any code you want here but don't commit it.
// keep it empty for future projects and for use ase a template.
package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"hash/fnv"
	"math"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"

	"work-tracker/src/pkg/config"
	er "work-tracker/src/pkg/error"
	"work-tracker/src/pkg/logger"
	"work-tracker/src/pkg/util"
)

/* ---------- Data Types ---------- */

// Chunk represents a single JSONL line from the per-day log.
type Chunk struct {
	TaskName   string    `json:"task_name"`
	StartedAt  time.Time `json:"started_at"`
	FinishedAt time.Time `json:"finished_at"`
	// active_time is stored as nanoseconds in the file
	ActiveTimeNs int64 `json:"active_time"`
}

// DayAggregation holds per-day totals and per-task durations.
type DayAggregation struct {
	Date           time.Time                 `json:"date"`
	TotalDuration  time.Duration             `json:"total_duration"`
	ActiveNsSum    int64                     `json:"active_ns_sum"`
	PerTaskDur     map[string]time.Duration  `json:"per_task_dur"`
	PerTaskActive  map[string]int64          `json:"per_task_active_ns"`
	EffectiveFocus time.Duration             `json:"effective_focus_duration"`
	Segments       []TaskSegmentForRendering `json:"-"`
}

// TaskSegmentForRendering is used only for deterministic ordered bar rendering.
type TaskSegmentForRendering struct {
	TaskName     string
	DurationPart time.Duration
	ColorHex     string
}

// PeriodTotals aggregates everything across the selected date range.
type PeriodTotals struct {
	StartDate           time.Time                `json:"start_date"`
	EndDate             time.Time                `json:"end_date"`
	TotalDuration       time.Duration            `json:"total_duration"`
	ActiveNsSum         int64                    `json:"active_ns_sum"`
	EffectiveFocusTotal time.Duration            `json:"effective_focus_total"`
	PerTaskTotals       map[string]time.Duration `json:"per_task_totals"`
	PerTaskActiveNs     map[string]int64         `json:"per_task_active_ns"`
	Days                []DayAggregation         `json:"days"`
}

/* ---------- Flags ---------- */

var (
	flagStartStr      = flag.String("start", "", "Start date DD-MM-YYYY (inclusive). If empty, defaults to this week's Monday (with -week-offset applied).")
	flagEndStr        = flag.String("end", "", "End date DD-MM-YYYY (inclusive). If empty, defaults to this week's Sunday (with -week-offset applied).")
	flagOutputPath    = flag.String("output", "./tmp/report.html", "Where to write the HTML report.")
	flagBarHeight     = flag.Int("bar-height", 200, "Height of each day's vertical bar in pixels.")
	flagActivityGamma = flag.Float64("activity-gamma", 0.7, "Gamma for smoothing Activity×Time (1=linear, <1 smooths differences, >1 emphasizes).")
	flagWeekOffset    = flag.Int("week-offset", 0, "When start/end are empty, offset the default week (Mon–Sun) by this many weeks. -1 for last week, +1 for next week, etc.")
)

/* ---------- Main ---------- */

func main() {
	util.CheckIfEnvVarsPresent([]string{})

	// common flags
	logLevelOverride := flag.Int("log-level", -1, "Log level. Default is whatever value is in configuration file. Keep at -1 to not override.")
	logDirOverride := flag.String("log-dir", "", "File directory at which to save log files. Keep empty to use configuration file instead.")
	configPath := flag.String("config", "./cfg/config.json", "Path to your configuration file.")

	// program's custom flags are declared globally above

	flag.Parse()
	config.InitializeConfig(*configPath, logger.LogLevel(*logLevelOverride), *logDirOverride)

	logger.Log(logger.Notice, logger.BoldBlueColor, "%s report entrypoint. Config path: '%s'", "Running", *configPath)

	// Resolve date range
	loc := time.Local
	startDate, endDate, e := resolveDateRange(*flagStartStr, *flagEndStr, *flagWeekOffset, loc)
	if e != nil {
		logger.Log(logger.Notice, logger.PurpleColor, "%s to resolve date range: '%s'", "Failed", e.ErrStr)
		os.Exit(1)
		return
	}
	logger.Log(logger.Notice, logger.BlueColor, "%s range '%s' -> '%s' (inclusive), week-offset '%v'", "Using", formatDate(startDate), formatDate(endDate), *flagWeekOffset)

	// Scan each day file and aggregate
	period, e := computePeriodTotals("./out", startDate, endDate, *flagActivityGamma, loc)
	if e != nil {
		logger.Log(logger.Notice, logger.PurpleColor, "%s to compute period totals: '%s'", "Failed", e.ErrStr)
		os.Exit(1)
		return
	}

	if period.TotalDuration <= 0 {
		logger.Log(logger.Notice, logger.CyanColor, "%s any tracked time between '%s' and '%s'. Writing an empty report anyway.", "No", formatDate(startDate), formatDate(endDate))
	}

	// Generate HTML (inline styles only)
	html := buildHTMLReport(period, *flagBarHeight, *flagActivityGamma)

	// Write output
	writeErr := os.MkdirAll(filepath.Dir(*flagOutputPath), 0o755)
	if writeErr != nil {
		e = er.NewError(writeErr, "failed to create output directory", *flagOutputPath)
		logger.Log(logger.Notice, logger.PurpleColor, "%s to prepare output dir '%s': '%s'", "Failed", *flagOutputPath, e.ErrStr)
		os.Exit(1)
		return
	}

	writeErr = os.WriteFile(*flagOutputPath, []byte(html), 0o644)
	if writeErr != nil {
		e = er.NewError(writeErr, "failed to write HTML file", *flagOutputPath)
		logger.Log(logger.Notice, logger.PurpleColor, "%s to write HTML to '%s': '%s'", "Failed", *flagOutputPath, e.ErrStr)
		os.Exit(1)
		return
	}

	logger.Log(logger.Notice, logger.GreenColor, "%s report to '%s' (days '%v', tasks '%v', total '%s', avg activity '%0.1f%%')",
		"Saved",
		*flagOutputPath,
		len(period.Days),
		len(period.PerTaskTotals),
		formatDuration(period.TotalDuration),
		averageActivityPct(period.ActiveNsSum, period.TotalDuration),
	)

	logger.Log(logger.Notice, logger.BoldBlueColor, "%s a file '%s' with %s", "Opening", *flagOutputPath, "google-chrome")
	err := exec.Command("google-chrome", *flagOutputPath).Start()
	er.QuitIfError(err, "Unable to open html report with google chrome")
}

/* ---------- Date Helpers ---------- */

/*
resolveDateRange determines the inclusive [start,end] days to report.
- If startStr/endStr provided (DD-MM-YYYY), uses those.
- Else computes this week's Monday..Sunday and applies weekOffset (± weeks).
*/
func resolveDateRange(startStr, endStr string, weekOffset int, loc *time.Location) (startDay time.Time, endDay time.Time, e *er.Error) {
	const ddmmyyyy = "02-01-2006"

	if strings.TrimSpace(startStr) != "" || strings.TrimSpace(endStr) != "" {
		var parseErr error
		if strings.TrimSpace(startStr) == "" || strings.TrimSpace(endStr) == "" {
			parseErr = fmt.Errorf("both -start and -end must be provided together in DD-MM-YYYY when either is set")
			e = er.NewError(parseErr, "bad date flags", fmt.Sprintf("start: '%s', end: '%s'", startStr, endStr))
			return
		}
		startDay, parseErr = time.ParseInLocation(ddmmyyyy, startStr, loc)
		if parseErr != nil {
			e = er.NewError(parseErr, "failed to parse -start", startStr)
			return
		}
		endDay, parseErr = time.ParseInLocation(ddmmyyyy, endStr, loc)
		if parseErr != nil {
			e = er.NewError(parseErr, "failed to parse -end", endStr)
			return
		}
		// Normalize to local midnight
		startDay = dayStart(startDay, loc)
		endDay = dayStart(endDay, loc)
		if endDay.Before(startDay) {
			parseErr = fmt.Errorf("end before start")
			e = er.NewErrorECOL(parseErr, "end date is before start date", "range", fmt.Sprintf("%s -> %s", formatDate(startDay), formatDate(endDay)))
			return
		}
		return
	}

	// Default: this week's Mon..Sun with week offset
	now := time.Now().In(loc)
	mon := mondayOfWeek(now, loc)
	if weekOffset != 0 {
		mon = mon.AddDate(0, 0, 7*weekOffset)
	}
	startDay = mon
	endDay = mon.AddDate(0, 0, 6)
	return
}

// mondayOfWeek returns local midnight Monday for the week containing t.
func mondayOfWeek(t time.Time, loc *time.Location) time.Time {
	t = dayStart(t, loc)
	wd := int(t.Weekday())
	// Go: Sunday=0, Monday=1, ... Saturday=6
	offset := 0
	if wd == 0 {
		offset = -6 // Sunday -> back to Monday
	} else {
		offset = 1 - wd
	}
	return t.AddDate(0, 0, offset)
}

func dayStart(t time.Time, loc *time.Location) time.Time {
	yr, mo, dy := t.In(loc).Date()
	return time.Date(yr, mo, dy, 0, 0, 0, 0, loc)
}

func formatDate(t time.Time) string {
	return t.Format("02-01-2006")
}

/* ---------- Aggregation ---------- */

/*
computePeriodTotals walks each day from start..end inclusive, reads ./out/DD-MM-YYYY.jsonl if present,
and aggregates per-day and per-task totals. Missing files count as zero-day.
EffectiveFocus uses duration * pow(activityFraction, gamma) per chunk.
*/
func computePeriodTotals(outDir string, startDay, endDay time.Time, gamma float64, loc *time.Location) (pt PeriodTotals, e *er.Error) {
	pt.StartDate = startDay
	pt.EndDate = endDay
	pt.PerTaskTotals = make(map[string]time.Duration)
	pt.PerTaskActiveNs = make(map[string]int64)

	days := []DayAggregation{}
	maxDays := int(endDay.Sub(startDay).Hours()/24) + 1

	i := 0
	for i < maxDays {
		d := startDay.AddDate(0, 0, i)
		dayAgg, dayErr := processSingleDay(outDir, d, gamma, loc)
		if dayErr != nil {
			// Hard error only if it's not "file missing"; missing file -> zero-day, we still append.
			if !os.IsNotExist(errorsUnwrap(dayErr.Err)) {
				e = er.NewErrorECOL(dayErr.Err, "failed to process day", "date", formatDate(d))
				return
			}
			// Build empty day
			dayAgg = DayAggregation{
				Date:          d,
				TotalDuration: 0,
				ActiveNsSum:   0,
				PerTaskDur:    map[string]time.Duration{},
				PerTaskActive: map[string]int64{},
			}
		}

		// Accumulate period totals
		pt.TotalDuration += dayAgg.TotalDuration
		pt.ActiveNsSum += dayAgg.ActiveNsSum
		pt.EffectiveFocusTotal += dayAgg.EffectiveFocus
		accumulateMapDuration(pt.PerTaskTotals, dayAgg.PerTaskDur)
		accumulateMapInt64(pt.PerTaskActiveNs, dayAgg.PerTaskActive)

		days = append(days, dayAgg)
		i++
	}
	pt.Days = days

	// Pre-compute rendering segments for each day using deterministic task order (by total across period desc, then name)
	taskOrder := sortedTasksByTotal(pt.PerTaskTotals)
	colorCache := map[string]string{}
	for _, dn := range taskOrder {
		colorCache[dn] = taskColorHex(dn)
	}

	di := 0
	for di < len(pt.Days) {
		day := pt.Days[di]
		segments := []TaskSegmentForRendering{}
		ti := 0
		for ti < len(taskOrder) {
			task := taskOrder[ti]
			dur := day.PerTaskDur[task]
			if dur > 0 {
				segments = append(segments, TaskSegmentForRendering{
					TaskName:     task,
					DurationPart: dur,
					ColorHex:     colorCache[task],
				})
			}
			ti++
		}
		pt.Days[di].Segments = segments
		di++
	}

	return
}

/*
processSingleDay reads one file ./out/DD-MM-YYYY.jsonl and aggregates totals.
If file is not present, returns an os.IsNotExist error inside e.Err.
*/
func processSingleDay(outDir string, day time.Time, gamma float64, loc *time.Location) (agg DayAggregation, e *er.Error) {
	filename := filepath.Join(outDir, formatDate(day)+".jsonl")
	file, openErr := os.Open(filename)
	if openErr != nil {
		// Return error so caller can treat NotExist as a zero-day.
		e = er.NewErrorECOL(openErr, "failed to open day file", "filename", filename)
		return
	}
	defer file.Close()

	agg = DayAggregation{
		Date:          day,
		TotalDuration: 0,
		ActiveNsSum:   0,
		PerTaskDur:    map[string]time.Duration{},
		PerTaskActive: map[string]int64{},
	}

	sc := bufio.NewScanner(file)
	lineNum := 0
	for sc.Scan() {
		line := strings.TrimSpace(sc.Text())
		lineNum++

		if line == "" {
			continue
		}

		var tmp struct {
			TaskName   string `json:"task_name"`
			StartedAt  string `json:"started_at"`
			FinishedAt string `json:"finished_at"`
			ActiveNs   int64  `json:"active_time"`
		}
		unmarshalErr := json.Unmarshal([]byte(line), &tmp)
		if unmarshalErr != nil {
			// Log + continue (do not silently fail).
			ctx := map[string]any{
				"file":        filename,
				"line_number": lineNum,
				"text":        line,
			}
			_ = er.NewErrorECML(unmarshalErr, "failed to parse JSON chunk", "line", ctx) // constructed for context if needed
			logger.Log(logger.Notice, logger.CyanColor, "%s line '%v' in '%s': '%s'", "Skipping bad", lineNum, filename, unmarshalErr.Error())
			continue
		}

		start, parseErr := time.Parse(time.RFC3339Nano, tmp.StartedAt)
		if parseErr != nil {
			logger.Log(logger.Notice, logger.CyanColor, "%s start on line '%v' in '%s': '%s'", "Skipping bad", lineNum, filename, parseErr.Error())
			continue
		}
		finish, parseErr := time.Parse(time.RFC3339Nano, tmp.FinishedAt)
		if parseErr != nil {
			logger.Log(logger.Notice, logger.CyanColor, "%s finish on line '%v' in '%s': '%s'", "Skipping bad", lineNum, filename, parseErr.Error())
			continue
		}

		duration := finish.Sub(start)
		if duration <= 0 {
			logger.Log(logger.Notice, logger.CyanColor, "%s non-positive duration on line '%v' in '%s'", "Skipping", lineNum, filename)
			continue
		}

		activeNs := clampInt64(tmp.ActiveNs, 0, duration.Nanoseconds())
		taskName := normalizeTaskName(tmp.TaskName)

		// Update day totals
		agg.TotalDuration += duration
		agg.ActiveNsSum += activeNs
		agg.PerTaskDur[taskName] = agg.PerTaskDur[taskName] + duration
		agg.PerTaskActive[taskName] = agg.PerTaskActive[taskName] + activeNs

		// Effective focus (smoothed) for this chunk:
		af := 0.0
		if duration > 0 {
			af = float64(activeNs) / float64(duration.Nanoseconds()) // 0..1
			if af < 0 {
				af = 0
			}
			if af > 1 {
				af = 1
			}
		}
		weight := math.Pow(af, gamma)
		eff := time.Duration(float64(duration) * weight)
		agg.EffectiveFocus += eff
	}

	scErr := sc.Err()
	if scErr != nil {
		e = er.NewErrorECOL(scErr, "failed to read file", "filename", filename)
		return
	}

	return
}

/* ---------- Small Helpers ---------- */

func normalizeTaskName(s string) string {
	if strings.TrimSpace(s) == "" {
		return "(No task)"
	}
	return s
}

func accumulateMapDuration(dst map[string]time.Duration, src map[string]time.Duration) {
	for k, v := range src {
		dst[k] = dst[k] + v
	}
}

func accumulateMapInt64(dst map[string]int64, src map[string]int64) {
	for k, v := range src {
		dst[k] = dst[k] + v
	}
}

func clampInt64(v, lo, hi int64) int64 {
	if v < lo {
		return lo
	}
	if v > hi {
		return hi
	}
	return v
}

func averageActivityPct(activeNsSum int64, total time.Duration) float64 {
	if total <= 0 {
		return 0
	}
	return 100.0 * (float64(activeNsSum) / float64(total.Nanoseconds()))
}

func formatDuration(d time.Duration) string {
	// produce "1h 2m 3s" style; skip zeros
	neg := d < 0
	if neg {
		d = -d
	}
	totalSec := int64(d.Round(time.Second).Seconds())
	h := totalSec / 3600
	m := (totalSec % 3600) / 60
	s := totalSec % 60

	parts := []string{}
	if h > 0 {
		parts = append(parts, fmt.Sprintf("%vh", h))
	}
	if m > 0 {
		parts = append(parts, fmt.Sprintf("%vm", m))
	}
	if s > 0 || len(parts) == 0 {
		parts = append(parts, fmt.Sprintf("%vs", s))
	}
	out := strings.Join(parts, " ")
	if neg {
		return "-" + out
	}
	return out
}

/* ---------- Deterministic Task Color (string -> hex) ---------- */

// taskColorHex maps a task name to a stable, medium-saturated color (hex).
func taskColorHex(name string) string {
	h := fnvHash32(name)
	hue := float64(h % 360) // 0..359
	sat := 0.65             // medium saturation for email readability
	light := 0.55           // medium lightness
	r, g, b := hslToRgb(hue/360.0, sat, light)
	return fmt.Sprintf("#%02x%02x%02x", r, g, b)
}

func fnvHash32(s string) uint32 {
	h := fnv.New32a()
	_, _ = h.Write([]byte(s))
	return h.Sum32()
}

// HSL -> RGB (0..255)
func hslToRgb(h, s, l float64) (uint8, uint8, uint8) {
	if s == 0 {
		v := uint8(math.Round(l * 255))
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
	return uint8(math.Round(r * 255)), uint8(math.Round(g * 255)), uint8(math.Round(b * 255))
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

/* ---------- Activity -> Red→Yellow→Green color ---------- */

// summaryActivityColorHex returns a traffic-light color for avg activity.
// 0..50%: red→yellow, 50..75%: yellow→green, >75%: solid green.
func summaryActivityColorHex(avgPct float64) string {
	cl := func(v, lo, hi float64) float64 {
		if v < lo {
			return lo
		}
		if v > hi {
			return hi
		}
		return v
	}
	avgPct = cl(avgPct, 0, 100)

	red := hexToRGB("#e53935")
	yellow := hexToRGB("#ffb300")
	green := hexToRGB("#4caf50")

	var r, g, b int
	if avgPct <= 50 {
		t := avgPct / 50.0
		r, g, b = lerpRGB(red, yellow, t)
	} else if avgPct <= 75 {
		t := (avgPct - 50.0) / 25.0
		r, g, b = lerpRGB(yellow, green, t)
	} else {
		r, g, b = green[0], green[1], green[2]
	}
	return fmt.Sprintf("#%02x%02x%02x", r, g, b)
}

func hexToRGB(h string) [3]int {
	h = strings.TrimPrefix(h, "#")
	if len(h) != 6 {
		return [3]int{0, 0, 0}
	}
	r, _ := strconv.ParseInt(h[0:2], 16, 32)
	g, _ := strconv.ParseInt(h[2:4], 16, 32)
	b, _ := strconv.ParseInt(h[4:6], 16, 32)
	return [3]int{int(r), int(g), int(b)}
}

func lerpRGB(a [3]int, b [3]int, t float64) (int, int, int) {
	cl := func(x float64, lo, hi float64) float64 {
		if x < lo {
			return lo
		}
		if x > hi {
			return hi
		}
		return x
	}
	t = cl(t, 0, 1)
	r := int(math.Round((1-t)*float64(a[0]) + t*float64(b[0])))
	g := int(math.Round((1-t)*float64(a[1]) + t*float64(b[1])))
	bv := int(math.Round((1-t)*float64(a[2]) + t*float64(b[2])))
	return r, g, bv
}

/* ---------- Sorting ---------- */

func sortedTasksByTotal(m map[string]time.Duration) []string {
	type kv struct {
		Key string
		Val time.Duration
	}
	arr := make([]kv, 0, len(m))
	for k, v := range m {
		arr = append(arr, kv{k, v})
	}
	sort.Slice(arr, func(i, j int) bool {
		if arr[i].Val == arr[j].Val {
			return arr[i].Key < arr[j].Key
		}
		return arr[i].Val > arr[j].Val
	})
	out := make([]string, 0, len(arr))
	for _, it := range arr {
		out = append(out, it.Key)
	}
	return out
}

/* ---------- HTML Rendering (inline styles) ---------- */

func buildHTMLReport(period PeriodTotals, barHeight int, gamma float64) string {
	total := period.TotalDuration
	avgAct := averageActivityPct(period.ActiveNsSum, total)
	avgColor := summaryActivityColorHex(avgAct)

	// Prepare legend order (tasks sorted by total desc)
	taskOrder := sortedTasksByTotal(period.PerTaskTotals)

	// Precompute max per-day duration for vertical scaling
	var maxDay time.Duration
	for _, d := range period.Days {
		if d.TotalDuration > maxDay {
			maxDay = d.TotalDuration
		}
	}
	if maxDay <= 0 {
		maxDay = time.Second // avoid div-by-zero; keeps zero bars looking empty
	}

	sb := &strings.Builder{}
	title := fmt.Sprintf("Work-Tracker Report — %s to %s", formatDate(period.StartDate), formatDate(period.EndDate))

	// Wrapper
	fmt.Fprintf(sb, "<!DOCTYPE html><html><head><meta charset=\"utf-8\"><title>%s</title></head><body>", escapeHTML(title))
	fmt.Fprintf(sb, "<table role=\"presentation\" width=\"100%%\" cellpadding=\"0\" cellspacing=\"0\" style=\"max-width:680px;margin:0 auto;\">")

	/* ---------- Heading ---------- */
	fmt.Fprintf(sb, "<tr><td align=\"center\" style=\"font-family:Arial, sans-serif;color:#222;font-size:16px;padding:12px 0;\">%s</td></tr>", escapeHTML(title))

	/* ---------- Summary Section ---------- */
	fmt.Fprintf(sb, "<tr><td style=\"padding:8px 0;\">")
	fmt.Fprintf(sb, "<table role=\"presentation\" cellpadding=\"0\" cellspacing=\"0\" style=\"border-collapse:collapse;width:100%%;\">")

	// Totals row with avg activity badge
	fmt.Fprintf(sb, "<tr>")
	// Total time
	fmt.Fprintf(sb, "<td style=\"font-family:Arial, sans-serif;font-size:14px;color:#222;padding:6px 0;\">"+
		"<b>Total Worked:</b> %s</td>", escapeHTML(formatDuration(total)))
	// Average activity badge
	fmt.Fprintf(sb, "<td align=\"right\" style=\"font-family:Arial, sans-serif;font-size:14px;color:#222;padding:6px 0;\">"+
		"<span style=\"display:inline-block;background:%s;color:#fff;border-radius:4px;padding:4px 8px;\">Avg activity: %0.1f%%</span>"+
		"</td>", avgColor, avgAct)
	fmt.Fprintf(sb, "</tr>")

	// Effective focus (smoothed)
	fmt.Fprintf(sb, "<tr><td colspan=\"2\" style=\"font-family:Arial, sans-serif;font-size:13px;color:#555;padding:2px 0 10px 0;\">"+
		"<b>Activity×Time (γ=%0.2f):</b> %s</td></tr>", gamma, escapeHTML(formatDuration(period.EffectiveFocusTotal)))

	// Tasks list with colored squares
	fmt.Fprintf(sb, "<tr><td colspan=\"2\" style=\"padding:8px 0;\">")
	fmt.Fprintf(sb, "<table role=\"presentation\" cellpadding=\"0\" cellspacing=\"0\" style=\"border-collapse:collapse;width:100%%;\">")
	if len(taskOrder) == 0 {
		fmt.Fprintf(sb, "<tr><td style=\"font-family:Arial, sans-serif;font-size:13px;color:#777;\">No tasks in this period.</td></tr>")
	} else {
		for _, tname := range taskOrder {
			tdur := period.PerTaskTotals[tname]
			color := taskColorHex(tname)
			fmt.Fprintf(sb, "<tr>")
			fmt.Fprintf(sb, "<td style=\"padding:3px 6px 3px 0;width:18px;\">"+
				"<span style=\"display:inline-block;width:12px;height:12px;background:%s;\"></span></td>", color)
			fmt.Fprintf(sb, "<td style=\"font-family:Arial, sans-serif;font-size:13px;color:#333;padding:3px 0;\">%s</td>", escapeHTML(tname))
			fmt.Fprintf(sb, "<td align=\"right\" style=\"font-family:Arial, sans-serif;font-size:13px;color:#444;padding:3px 0;\">%s</td>", escapeHTML(formatDuration(tdur)))
			fmt.Fprintf(sb, "</tr>")
		}
	}
	fmt.Fprintf(sb, "</table>")
	fmt.Fprintf(sb, "</td></tr>")

	fmt.Fprintf(sb, "</table>") // end summary inner
	fmt.Fprintf(sb, "</td></tr>")

	/* ---------- Stacked Vertical Bars (per day, per task) ---------- */
	fmt.Fprintf(sb, "<tr><td align=\"center\" style=\"padding:10px 0;\">")
	fmt.Fprintf(sb, "<table role=\"presentation\" cellpadding=\"0\" cellspacing=\"0\" style=\"border-collapse:collapse;\">")
	fmt.Fprintf(sb, "<tr valign=\"bottom\">")

	// One column per day
	for _, d := range period.Days {
		// Outer cell for day
		fmt.Fprintf(sb, "<td align=\"center\" style=\"padding:0 10px;\">")
		// Bar container
		fmt.Fprintf(sb, "<table role=\"presentation\" cellpadding=\"0\" cellspacing=\"0\" style=\"border-collapse:collapse;\">")
		fmt.Fprintf(sb, "<tr><td style=\"background:#e0e0e0;width:28px;height:%vpx;line-height:0;font-size:0;\">", barHeight)
		fmt.Fprintf(sb, "<table role=\"presentation\" cellpadding=\"0\" cellspacing=\"0\" style=\"border-collapse:collapse;width:28px;\">")

		// Compute total pixels this day
		barPixels := 0
		if maxDay > 0 && d.TotalDuration > 0 {
			barPixels = int(math.Round(float64(barHeight) * (float64(d.TotalDuration) / float64(maxDay))))
			if barPixels < 0 {
				barPixels = 0
			}
			if barPixels > barHeight {
				barPixels = barHeight
			}
		}

		// Empty spacer at the top to align bottoms
		topSpacer := barHeight - barPixels
		if topSpacer < 0 {
			topSpacer = 0
		}
		fmt.Fprintf(sb, "<tr><td style=\"height:%vpx;line-height:0;font-size:0;\">&nbsp;</td></tr>", topSpacer)

		// Segments (bottom-up)
		if barPixels > 0 && len(d.Segments) > 0 {
			// Compute pixel heights for each segment proportionally
			pxs := make([]int, len(d.Segments))
			sumPx := 0
			for i, seg := range d.Segments {
				p := int(math.Round(float64(barPixels) * (float64(seg.DurationPart) / float64(d.TotalDuration))))
				pxs[i] = p
				sumPx += p
			}
			// Adjust rounding errors to fill exactly barPixels
			diff := barPixels - sumPx
			ai := 0
			for diff != 0 && len(pxs) > 0 {
				if diff > 0 {
					pxs[ai%len(pxs)]++
					diff--
				} else {
					if pxs[ai%len(pxs)] > 0 {
						pxs[ai%len(pxs)]--
						diff++
					}
				}
				ai++
			}
			// Render rows (each a colored block)
			for i, seg := range d.Segments {
				h := pxs[i]
				if h <= 0 {
					continue
				}
				fmt.Fprintf(sb, "<tr><td style=\"background:%s;height:%vpx;line-height:0;font-size:0;\">&nbsp;</td></tr>", seg.ColorHex, h)
			}
		}

		fmt.Fprintf(sb, "</table>") // inner bar
		fmt.Fprintf(sb, "</td></tr>")
		fmt.Fprintf(sb, "</table>") // bar container

		// Day label
		fmt.Fprintf(sb, "<div style=\"font-family:Arial, sans-serif;font-size:12px;color:#555;padding-top:8px;\">%s</div>", escapeHTML(d.Date.Format("Mon 02")))
		fmt.Fprintf(sb, "</td>")
	}

	fmt.Fprintf(sb, "</tr>")
	fmt.Fprintf(sb, "</table>")
	fmt.Fprintf(sb, "</td></tr>")

	/* ---------- Legend ---------- */
	if len(taskOrder) > 0 {
		fmt.Fprintf(sb, "<tr><td align=\"center\" style=\"padding:8px 0 14px 0;\">")
		fmt.Fprintf(sb, "<table role=\"presentation\" cellpadding=\"0\" cellspacing=\"0\" style=\"border-collapse:collapse;\">")
		// Render legend in rows with 3 items per row (email-friendly)
		perRow := 3
		for i := 0; i < len(taskOrder); i += perRow {
			fmt.Fprintf(sb, "<tr>")
			for j := 0; j < perRow && i+j < len(taskOrder); j++ {
				name := taskOrder[i+j]
				color := taskColorHex(name)
				fmt.Fprintf(sb, "<td style=\"background:%s;width:10px;height:10px;line-height:0;font-size:0;\">&nbsp;</td>", color)
				fmt.Fprintf(sb, "<td style=\"font-family:Arial, sans-serif;font-size:12px;color:#555;padding:0 12px 6px 6px;\">%s</td>", escapeHTML(name))
			}
			fmt.Fprintf(sb, "</tr>")
		}
		fmt.Fprintf(sb, "</table>")
		fmt.Fprintf(sb, "</td></tr>")
	}

	fmt.Fprintf(sb, "</table>") // wrapper
	fmt.Fprintf(sb, "</body></html>")
	return sb.String()
}

func escapeHTML(s string) string {
	r := strings.NewReplacer(
		`&`, "&amp;",
		`<`, "&lt;",
		`>`, "&gt;",
		`"`, "&quot;",
		`'`, "&#39;",
	)
	return r.Replace(s)
}

/* ---------- Error Unwrap Helper ---------- */

func errorsUnwrap(err error) error {
	type unwrapper interface {
		Unwrap() error
	}
	u, ok := err.(unwrapper)
	if ok && u.Unwrap() != nil {
		return u.Unwrap()
	}
	return err
}
