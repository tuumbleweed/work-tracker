package report

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	tl "github.com/tuumbleweed/tintlog/logger"
	"github.com/tuumbleweed/tintlog/palette"
	"github.com/tuumbleweed/xerr"
)

/*
Build the report: read files, aggregate, render HTML, write to disk.

Input layout (new):
  <inputDir>/<YEAR>/<monthname>/<D>_<monthname>_<YEAR>.jsonl
Example:
  out/2026/january/23_january_2026.jsonl
*/
func BuildReport(inputDir string, startDate, endDate time.Time, outPath string, barRef time.Duration, smooth float64) (e *xerr.Error) {
	tl.Log(tl.Notice, palette.Blue, "%s files from '%s' for '%s'..'%s'",
		"Reading", inputDir, startDate.Format("02-01-2006"), endDate.Format("02-01-2006"),
	)

	dates := enumerateDates(startDate, endDate)
	daySummaries := make([]DaySummary, 0, len(dates))
	totals := ReportTotals{
		PerTaskTotals: make(map[string]time.Duration),
	}

	for _, d := range dates {
		fp := dayFilePathYM(inputDir, d) // <-- updated path scheme (year/month)
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

	// Ensure output directory exists (range can span years/months; outPath can be anywhere)
	outDir := filepath.Dir(outPath)
	if mkErr := os.MkdirAll(outDir, 0o755); mkErr != nil {
		return xerr.NewErrorECOL(mkErr, "failed to create report output directory", "dir", outDir)
	}

	if err := os.WriteFile(outPath, buf.Bytes(), 0o644); err != nil {
		return xerr.NewErrorECOL(err, "failed to write HTML report", "path", outPath)
	}

	tl.Log(tl.Notice, palette.Green, "%s report to '%s' (%s, %s days)",
		"Wrote", outPath, formatDuration(totals.TotalWorked), len(daySummaries),
	)
	return nil
}

// dayFilePathYM builds the per-day filepath for the new year/month layout.
// Layout:
//   <root>/<YYYY>/<monthname>/<D>_<monthname>_<YYYY>.jsonl
func dayFilePathYM(root string, d time.Time) string {
	year := d.Format("2006")
	month := strings.ToLower(d.Format("January")) // "january"
	day := d.Format("02")
	fname := fmt.Sprintf("%s_%s_%s.jsonl", day, month, year)
	return filepath.Join(root, year, month, fname)
}

// resolveRange loads the TZ and determines [startDate, endDate] from flags.
// Returns (*time.Location, startDate, endDate, *xerr.Error).
func ResolveRange(flagTZ, flagStart, flagEnd string) (*time.Location, time.Time, time.Time, *xerr.Error) {
	loc, tzErr := time.LoadLocation(flagTZ)
	if tzErr != nil {
		return nil, time.Time{}, time.Time{}, xerr.NewErrorECOL(tzErr, "failed to load timezone", "tz", flagTZ)
	}

	var (
		startDate time.Time
		endDate   time.Time
		err       error
	)

	switch {
	case flagStart == "" && flagEnd == "":
		startDate, endDate = currentWeekRange(loc)

	case flagStart != "" && flagEnd == "":
		startDate, err = parseDMY(flagStart, loc)
		if err != nil {
			return loc, time.Time{}, time.Time{}, xerr.NewErrorECOL(err, "failed to parse start date", "start", flagStart)
		}
		endDate = startDate

	case flagStart == "" && flagEnd != "":
		endDate, err = parseDMY(flagEnd, loc)
		if err != nil {
			return loc, time.Time{}, time.Time{}, xerr.NewErrorECOL(err, "failed to parse end date", "end", flagEnd)
		}
		startDate = endDate

	default:
		startDate, err = parseDMY(flagStart, loc)
		if err != nil {
			return loc, time.Time{}, time.Time{}, xerr.NewErrorECOL(err, "failed to parse start date", "start", flagStart)
		}
		endDate, err = parseDMY(flagEnd, loc)
		if err != nil {
			return loc, time.Time{}, time.Time{}, xerr.NewErrorECOL(err, "failed to parse end date", "end", flagEnd)
		}
	}

	return loc, startDate, endDate, nil
}
