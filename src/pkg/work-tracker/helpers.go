package worktracker

import (
	"fmt"
	"os/exec"
	"path/filepath"
	"time"

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
	if total == 0 {
		return 0
	}
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
