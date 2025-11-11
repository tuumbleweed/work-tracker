package report

import (
	"bufio"
	"encoding/json"
	"math"
	"os"
	"strings"
	"time"

	tl "github.com/tuumbleweed/tintlog/logger"
	"github.com/tuumbleweed/tintlog/palette"
	"github.com/tuumbleweed/xerr"
)

/*
Read a single day file into a DaySummary. Missing file => empty summary (no error).
*/
func readDayFile(filePath string, date time.Time, smooth float64) (sum DaySummary, e *xerr.Error) {
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
			tl.Log(tl.Notice, palette.Cyan, "%s missing day file '%s' (treated as 0)", "Skipping", filePath)
			return sum, nil
		}
		e = xerr.NewErrorECOL(statErr, "unable to stat day file", "path", filePath)
		return sum, e
	}

	f, openErr := os.Open(filePath)
	if openErr != nil {
		e = xerr.NewErrorECOL(openErr, "unable to open day file", "path", filePath)
		return sum, e
	}
	defer f.Close()

	sc := bufio.NewScanner(f)
	buf := make([]byte, 0, 64*1024)
	sc.Buffer(buf, 2*1024*1024)

	lineNumber := 0
	sum.TaskDurations["Unassigned Time"] = 1 * time.Nanosecond // add this to have it take first (gray) color always, even if not present
	for sc.Scan() {
		line := strings.TrimSpace(sc.Text())
		lineNumber++
		if line == "" {
			continue
		}
		var ch Chunk
		uErr := json.Unmarshal([]byte(line), &ch)
		if uErr != nil {
			tl.Log(tl.Notice, palette.Purple, "%s malformed JSON in '%s' line %d", "Skipping", filePath, lineNumber)
			continue
		}
		if !ch.FinishedAt.After(ch.StartedAt) {
			tl.Log(tl.Notice, palette.Purple, "%s bad chunk time window in '%s' line %d", "Skipping", filePath, lineNumber)
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
		e = xerr.NewErrorECML(sErr, "scanner error while reading day file", "line",
			map[string]any{"path": filePath, "last_line": lineNumber})
		return sum, e
	}
	return sum, nil
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
