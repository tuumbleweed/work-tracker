package worktracker

import (
	"fmt"
	"path/filepath"
	"time"
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

func getActivityPercentage(activeDuration, workedDuration time.Duration) float64 {
	return (float64(activeDuration) / float64(workedDuration)) * 100
}
