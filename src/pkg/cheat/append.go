package cheat

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"strings"
	"time"

	tl "github.com/tuumbleweed/tintlog/logger"
	"github.com/tuumbleweed/tintlog/palette"
	"github.com/tuumbleweed/xerr"
)

// Chunk uses the tracker's JSONL schema without importing the desktop UI.
type Chunk struct {
	TaskName   string        `json:"task_name"`
	StartedAt  time.Time     `json:"started_at"`
	FinishedAt time.Time     `json:"finished_at"`
	ActiveTime time.Duration `json:"active_time"`
}

// Append adds one chunk after the last record. An empty task inherits that record's task.
func Append(workDir string, date time.Time, after, minutes, activity int64, task string) (chunk Chunk, e *xerr.Error) {
	year, month, day := date.Format("2006"), strings.ToLower(date.Format("January")), date.Format("02")
	path := filepath.Join(workDir, year, month, fmt.Sprintf("%s_%s_%s.jsonl", day, month, year))
	fail := func(err error, message string) (Chunk, *xerr.Error) {
		return Chunk{}, xerr.NewErrorECOL(err, message, "path", path)
	}
	if after < 0 || minutes <= 0 || activity < 0 || activity > 100 || after > math.MaxInt64/int64(time.Minute) || minutes > math.MaxInt64/int64(time.Minute) {
		return fail(errors.New("require after >= 0, minutes > 0 (within time.Duration range), and activity in [0, 100]"), "invalid entry arguments")
	}
	f, err := os.OpenFile(path, os.O_RDWR|os.O_APPEND, 0o644)
	if err != nil {
		return fail(err, "unable to open existing day file")
	}
	defer func() {
		if err := f.Close(); err != nil && e == nil {
			e = xerr.NewErrorECOL(err, "unable to close day file", "path", path)
		}
	}()
	scanner := bufio.NewScanner(f)
	scanner.Buffer(make([]byte, 64*1024), 2*1024*1024)
	var last Chunk
	found := false
	for line := 1; scanner.Scan(); line++ {
		raw := strings.TrimSpace(scanner.Text())
		if raw == "" || strings.HasPrefix(raw, "#") {
			continue
		}
		var record Chunk
		if err := json.Unmarshal([]byte(raw), &record); err != nil {
			return fail(err, fmt.Sprintf("invalid JSON at line %d", line))
		}
		if record.StartedAt.IsZero() || !record.FinishedAt.After(record.StartedAt) {
			return fail(errors.New("invalid time window"), fmt.Sprintf("invalid entry at line %d", line))
		}
		last, found = record, true
	}
	if err := scanner.Err(); err != nil {
		return fail(err, "unable to read day file")
	}
	if !found {
		return fail(errors.New("no previous entry to anchor the new entry"), "empty day file")
	}
	if task == "" {
		task = last.TaskName
	}
	duration := time.Duration(minutes) * time.Minute
	start := last.FinishedAt.Add(time.Duration(after) * time.Minute)
	finish := start.Add(duration)
	// Keep the single entry within its day, using the previous entry's offset.
	if start.Format("2006-01-02") != date.Format("2006-01-02") || finish.Add(-time.Nanosecond).Format("2006-01-02") != date.Format("2006-01-02") {
		return fail(errors.New("entry would extend outside the selected day"), "invalid time window")
	}
	chunk = Chunk{TaskName: task, StartedAt: start, FinishedAt: finish, ActiveTime: duration / 100 * time.Duration(activity)}
	data, err := json.Marshal(chunk)
	if err != nil {
		return fail(err, "unable to marshal entry")
	}
	info, err := f.Stat()
	if err != nil {
		return fail(err, "unable to stat day file")
	}
	// Hand-edited files may not end in a newline.
	if info.Size() > 0 {
		var end [1]byte
		if _, err := f.ReadAt(end[:], info.Size()-1); err != nil {
			return fail(err, "unable to read last byte")
		}
		if end[0] != '\n' {
			data = append([]byte{'\n'}, data...)
		}
	}
	if _, err := f.Write(append(data, '\n')); err != nil {
		return fail(err, "unable to append entry")
	}
	tl.Log(tl.Notice, palette.Green, "Added %s to '%s': %s to %s, active %s", task, path, start.Format(time.RFC3339Nano), finish.Format(time.RFC3339Nano), chunk.ActiveTime)
	return chunk, nil
}
