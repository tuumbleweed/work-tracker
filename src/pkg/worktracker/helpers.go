package worktracker

import (
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
