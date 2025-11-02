package report

import (
	"fmt"
	"time"
)

// Format adaptive titles like:
// "Daily Report — 02 Nov 2025"
// "Weekly Report — 25 – 31 Oct 2025"
// "Monthly Report — Oct 2025"
// "Quarterly Report — Q4 2025"
// "Yearly Report — 2025"
// "Custom Report — 25 Oct 2025 – 04 Nov 2025"
func reportTitle(start, end time.Time) string {
	start = start.Local()
	end = end.Local()

	label := periodLabel(start, end)

	switch label {
	case "Daily":
		return fmt.Sprintf("Daily Report — %s", start.Format("02 Jan 2006"))

	case "Weekly":
		// same year? print year once
		if start.Year() == end.Year() && start.Month() == end.Month() {
			return fmt.Sprintf("Weekly Report — %s – %s %d", start.Format("02"), end.Format("02 Jan"), end.Year())
		}
		if start.Year() == end.Year() {
			return fmt.Sprintf("Weekly Report — %s – %s %d", start.Format("02 Jan"), end.Format("02 Jan"), end.Year())
		}
		return fmt.Sprintf("Weekly Report — %s – %s", start.Format("02 Jan 2006"), end.Format("02 Jan 2006"))

	case "Monthly":
		return fmt.Sprintf("Monthly Report — %s %d", start.Format("Jan"), start.Year())

	case "Quarterly":
		_, q := quarterOf(start)
		return fmt.Sprintf("Quarterly Report — Q%d %d", q, start.Year())

	case "Yearly":
		return fmt.Sprintf("Yearly Report — %d", start.Year())

	default: // "Custom"
		if start.Equal(end) {
			return fmt.Sprintf("Report — %s", start.Format("02 Jan 2006"))
		}
		if start.Year() == end.Year() && start.Month() == end.Month() {
			return fmt.Sprintf("Report — %s – %s %d", start.Format("02"), end.Format("02 Jan"), end.Year())
		}
		if start.Year() == end.Year() {
			return fmt.Sprintf("Report — %s – %s %d", start.Format("02 Jan"), end.Format("02 Jan"), end.Year())
		}
		return fmt.Sprintf("Report — %s – %s", start.Format("02 Jan 2006"), end.Format("02 Jan 2006"))
	}
}

func periodLabel(start, end time.Time) string {
	// normalize to date-only
	sd := time.Date(start.Year(), start.Month(), start.Day(), 0, 0, 0, 0, start.Location())
	ed := time.Date(end.Year(), end.Month(), end.Day(), 0, 0, 0, 0, end.Location())

	// Daily
	if sd.Equal(ed) {
		return "Daily"
	}

	// Yearly (full calendar year)
	if start.Year() == end.Year() &&
		isStartOfYear(sd) && isEndOfYear(ed) {
		return "Yearly"
	}

	// Quarterly (full quarter)
	if start.Year() == end.Year() {
		qs, qn := quarterOf(sd)
		qe, qn2 := quarterOf(ed)
		if qn == qn2 && qs.Equal(sd) && qe.Equal(ed) {
			return "Quarterly"
		}
	}

	// Monthly (full calendar month)
	if start.Year() == end.Year() && start.Month() == end.Month() &&
		isStartOfMonth(sd) && isEndOfMonth(ed) {
		return "Monthly"
	}

	// Weekly (same ISO week & year)
	y1, w1 := sd.ISOWeek()
	y2, w2 := ed.ISOWeek()
	if y1 == y2 && w1 == w2 {
		return "Weekly"
	}

	return "Custom"
}

func isStartOfMonth(t time.Time) bool { return t.Day() == 1 }
func isEndOfMonth(t time.Time) bool   { return t.Day() == lastDayOfMonth(t.Year(), t.Month()) }
func lastDayOfMonth(y int, m time.Month) int {
	return time.Date(y, m+1, 0, 0, 0, 0, 0, time.Local).Day()
}

func isStartOfYear(t time.Time) bool { return t.Month() == time.January && t.Day() == 1 }
func isEndOfYear(t time.Time) bool   { return t.Month() == time.December && t.Day() == 31 }

func quarterOf(t time.Time) (startOfQuarter time.Time, q int) {
	q = (int(t.Month())-1)/3 + 1
	firstMonth := time.Month((q-1)*3 + 1)
	// start = first day of quarter
	startOfQuarter = time.Date(t.Year(), firstMonth, 1, 0, 0, 0, 0, t.Location())
	// caller can compute endOfQuarter as startOfQuarter+3mo-1day if needed
	// but for detection we’ll compute it here:
	return startOfQuarter, q
}
