package report

import (
	"bytes"
	"fmt"
	"html/template"
	"math"
	"sort"
	"time"
)

type reportTaskVM struct {
	ColorHex string
	Name     string
	Duration string
}

type reportTimeSegVM struct {
	ColorHex string
	HeightPx int
}

type reportTimeDayVM struct {
	ContainerHeightPx int
	TopSpacerPx       int
	Segments          []reportTimeSegVM

	HoursLabel string
	DayLabel   string
}

type reportActivityDayVM struct {
	ContainerHeightPx int
	TopSpacerPx       int
	BarHeightPx       int
	BarColorHex       string

	PctLabel string
	DayLabel string
}

type reportTemplateVM struct {
	Title string

	TotalWorked string

	AvgActivity float64
	// NOTE: if you want *zero* HTML generation in Go, convert buildSquares10HTML()
	// to return data and render it in the template.
	ActivitySquares template.HTML

	Tasks []reportTaskVM

	TimeByDayDays []reportTimeDayVM

	ActivityByTimeDays []reportActivityDayVM

	BarRefLabel string

	ChartW     int
	PadPx      int
	BarWPx     int
	PerDaySlot int

	WeeklyMode    bool
	ShowBarLabels bool

	Hex0   string
	Hex50  string
	Hex75  string
	Hex100 string
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
	_ = maxDayRowHeight

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
	_ = maxSmRowHeight

	// task listing (sorted by total desc)
	taskNames := make([]string, 0, len(totals.PerTaskTotals))
	for k := range totals.PerTaskTotals {
		taskNames = append(taskNames, k)
	}
	sort.Slice(taskNames, func(i, j int) bool {
		// pin "Unassigned Time" to the top
		ai, aj := taskNames[i], taskNames[j]
		if ai == "Unassigned Time" && aj != "Unassigned Time" {
			return true
		}
		if aj == "Unassigned Time" && ai != "Unassigned Time" {
			return false
		}

		// sort other cases
		di := totals.PerTaskTotals[ai]
		dj := totals.PerTaskTotals[aj]
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

	// ---------- Chart geometry (classic for ≤14 days, adaptive for >14) ----------
	nDays := len(daySummaries)
	if nDays < 1 {
		nDays = 1
	}
	weeklyMode := nDays <= 14
	showBarLabels := weeklyMode // hide in-bar labels when period > 2 weeks

	// Outputs we’ll use below
	barW := 28
	pad := 8
	perDaySlot := barW + 2*pad
	chartW := nDays * perDaySlot

	if !weeklyMode {
		// Adaptive geometry for long ranges
		const maxInnerW = 640 // inner width that fits your 760px wrapper
		const minBarW = 1
		const minPad = 1

		perDayW := int(math.Floor(float64(maxInnerW) / float64(nDays)))
		if perDayW < (minBarW + 1*minPad) {
			perDayW = minBarW + 1*minPad
		}
		pad = perDayW / 4 // ~25% of slot as padding
		if pad < minPad {
			pad = minPad
		}
		barW = perDayW - 1*pad
		if barW < minBarW {
			barW = minBarW
		}
		perDaySlot = barW + 1*pad
		chartW = nDays * perDaySlot
		if chartW > maxInnerW {
			chartW = maxInnerW
		}
	}

	// ---------- Label strategy (UNCHANGED semantics) ----------
	labelStep := 0
	useTicks := false
	_ = useTicks // keep semantics; labelStep is currently 0 in your snippet

	// ---------- Build view-model (NO HTML HERE) ----------
	tasksVM := make([]reportTaskVM, 0, len(taskNames))
	for i, name := range taskNames {
		dur := totals.PerTaskTotals[name]
		tasksVM = append(tasksVM, reportTaskVM{
			ColorHex: taskColorHex(i, name),
			Name:     name,
			Duration: formatDuration(dur),
		})
	}

	timeDaysVM := make([]reportTimeDayVM, 0, len(daySummaries))
	for dayIdx, dsum := range daySummaries {
		containerH := dayContainerHeights[dayIdx]

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

		segs := make([]reportTimeSegVM, 0, len(dayTasks))
		for segIdx, tname := range dayTasks {
			seg := segHeight(dsum.TaskDurations[tname])
			if seg <= 0 {
				continue
			}
			segs = append(segs, reportTimeSegVM{
				ColorHex: taskColorHex(segIdx, tname), // NOTE: matches your current behavior
				HeightPx: seg,
			})
		}

		label := ""
		if weeklyMode {
			label = weekdayShort(dsum.Date)
		} else {
			// keep your existing logic: labelStep currently 0 -> blanks
			if labelStep != 0 && (dayIdx%labelStep == 0) {
				label = weekdayShort(dsum.Date)
			}
		}

		timeDaysVM = append(timeDaysVM, reportTimeDayVM{
			ContainerHeightPx: containerH,
			TopSpacerPx:       topSpacer,
			Segments:          segs,
			HoursLabel:        fmt.Sprintf("%.1fh", dsum.TotalDuration.Hours()),
			DayLabel:          label,
		})
	}

	activityDaysVM := make([]reportActivityDayVM, 0, len(daySummaries))
	for dayIdx, dsum := range daySummaries {
		dayPct := 0.0
		if dsum.TotalDuration > 0 {
			dayPct = (float64(dsum.TotalActive) / float64(dsum.TotalDuration)) * 100.0
		}
		hex := colorToHex(barColorFor(dayPct))

		containerH := smContainerHeights[dayIdx]
		h := segHeight(dsum.SmoothedActiveTime)
		if h > containerH {
			h = containerH
		}
		top := containerH - h
		if top < 0 {
			top = 0
		}

		label := ""
		if weeklyMode {
			label = weekdayShort(dsum.Date)
		} else {
			if labelStep != 0 && (dayIdx%labelStep == 0) {
				label = weekdayShort(dsum.Date)
			}
		}

		activityDaysVM = append(activityDaysVM, reportActivityDayVM{
			ContainerHeightPx: containerH,
			TopSpacerPx:       top,
			BarHeightPx:       h,
			BarColorHex:       hex,
			PctLabel:          fmt.Sprintf("%.0f%%", dayPct),
			DayLabel:          label,
		})
	}

	vm := reportTemplateVM{
		Title: reportTitle(startDate, endDate),

		TotalWorked: formatDuration(totals.TotalWorked),

		AvgActivity:      avgActivity,
		ActivitySquares:  template.HTML(buildSquares10HTML(avgActivity, activityHex)),
		Tasks:            tasksVM,
		TimeByDayDays:    timeDaysVM,
		ActivityByTimeDays: activityDaysVM,

		BarRefLabel: formatDuration(barRef),

		ChartW:     chartW,
		PadPx:      pad,
		BarWPx:     barW,
		PerDaySlot: perDaySlot,

		WeeklyMode:    weeklyMode,
		ShowBarLabels: showBarLabels,

		Hex0:   hex0,
		Hex50:  hex50,
		Hex75:  hex75,
		Hex100: hex100,
	}

	tpl, err := template.ParseFiles("./cfg/report-template.html")
	if err != nil {
		// match old behavior style: fail hard rather than silently changing output
		panic(err)
	}
	if err := tpl.Execute(buf, vm); err != nil {
		panic(err)
	}
}
