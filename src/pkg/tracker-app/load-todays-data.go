package trackerapp

import (
	"bufio"
	"encoding/json"
	"errors"
	"os"
	"strings"
	"time"

	tl "github.com/tuumbleweed/tintlog/logger"
	"github.com/tuumbleweed/tintlog/palette"
	"github.com/tuumbleweed/xerr"
)

/*
loadFileActivityAndDuration reads a per-day JSONL file and returns:

- totalDuration:   sum of (FinishedAt - StartedAt) across all valid chunks
- totalActiveTime: sum of chunk.ActiveTime across all valid chunks

It processes the file in one pass. Any malformed line (bad JSON)
or a chunk where FinishedAt is not after StartedAt triggers an immediate error return.
*/
func loadFileActivityAndDuration(filePath string) (totalDuration, totalActiveTime time.Duration, timeByTask map[string]time.Duration, e *xerr.Error) {
	tl.Log(tl.Notice, palette.Blue, "Reading %s and %s from '%s'", "activity", "duration", filePath)

	timeByTask = make(map[string]time.Duration)
	fileHandle, openErr := os.Open(filePath)
	if openErr != nil {
		// e = xerr.NewErrorECOL(openErr, "failed to open JSONL file", "path", filePath)
		// return totalDuration, totalActiveTime, e
		tl.Log(tl.Notice, palette.PurpleBold, "No such file: '%s', %s", filePath, "skipping this step")
		return 0, 0, timeByTask, nil
	}
	defer func() {
		closeErr := fileHandle.Close()
		if closeErr != nil && e == nil {
			e = xerr.NewErrorECOL(closeErr, "failed to close JSONL file", "path", filePath)
			tl.Log(tl.Notice, palette.Purple, "Premature exit: %s '%s'", "close failed for", filePath)
		}
	}()

	scanner := bufio.NewScanner(fileHandle)
	var lineNumber int64 = 0

	for scanner.Scan() {
		lineNumber++

		rawLine := scanner.Text()
		trimmedLine := strings.TrimSpace(rawLine)

		if trimmedLine == "" || strings.HasPrefix(trimmedLine, "#") {
			continue
		}

		var chunk Chunk
		unmarshalErr := json.Unmarshal([]byte(trimmedLine), &chunk)
		if unmarshalErr != nil {
			e = xerr.NewErrorECML(unmarshalErr, "failed to parse JSON chunk", "line",
				map[string]any{
					"line_number": lineNumber,
					"text":        trimmedLine,
				},
			)
			tl.Log(tl.Notice, palette.Purple, "Premature exit on malformed JSON at line %v in '%s'", lineNumber, filePath)
			return totalDuration, totalActiveTime, timeByTask, e
		}

		if chunk.StartedAt.IsZero() {
			e = xerr.NewErrorECML(errors.New("invalid chunk"), "invalid chunk: StartedAt is zero", "context",
				map[string]any{
					"line_number": lineNumber,
				},
			)
			tl.Log(tl.Notice, palette.Purple, "Premature exit: %s at line %v in '%s'", "zero StartedAt", lineNumber, filePath)
			return totalDuration, totalActiveTime, timeByTask, e
		}
		if chunk.FinishedAt.IsZero() {
			e = xerr.NewErrorECML(errors.New("invalid chunk"), "invalid chunk: FinishedAt is zero", "context",
				map[string]any{
					"line_number": lineNumber,
				},
			)
			tl.Log(tl.Notice, palette.Purple, "Premature exit: %s at line %v in '%s'", "zero FinishedAt", lineNumber, filePath)
			return totalDuration, totalActiveTime, timeByTask, e
		}
		if !chunk.FinishedAt.After(chunk.StartedAt) {
			e = xerr.NewErrorECML(errors.New("invalid time interval"), "invalid time interval: FinishedAt is not after StartedAt", "context",
				map[string]any{
					"line_number":      lineNumber,
					"chunk.StartedAt":  chunk.StartedAt,
					"chunk.FinishedAt": chunk.FinishedAt,
				},
			)
			tl.Log(tl.Notice, palette.Purple, "Premature exit on invalid interval at line %v in '%s'", lineNumber, filePath)
			return totalDuration, totalActiveTime, timeByTask, e
		}

		chunkInterval := chunk.FinishedAt.Sub(chunk.StartedAt)

		if chunk.ActiveTime < 0 || chunk.ActiveTime > chunkInterval {
			e = xerr.NewErrorECML(errors.New("invalid active time"), "invalid active time: must be within [0, duration]", "context",
				map[string]any{
					"line_number": lineNumber,
					"duration":    chunkInterval.String(),
					"active_time": chunk.ActiveTime.String(),
					"started_at":  chunk.StartedAt,
					"finished_at": chunk.FinishedAt,
				},
			)
			tl.Log(tl.Notice, palette.Purple, "Premature exit on invalid active time at line %v in '%s'", lineNumber, filePath)
			return totalDuration, totalActiveTime, timeByTask, e
		}

		totalDuration += chunkInterval
		totalActiveTime += chunk.ActiveTime
		timeByTask[chunk.TaskName] += chunkInterval
	}

	scanErr := scanner.Err()
	if scanErr != nil {
		e = xerr.NewErrorECOL(scanErr, "scanner error while reading JSONL file", "path", filePath)
		tl.Log(tl.Notice, palette.Purple, "Premature exit: %s '%s'", "scanner error in", filePath)
		return totalDuration, totalActiveTime, timeByTask, e
	}

	tl.Log(tl.Notice, palette.Green, "Computed totals for '%s'", filePath)
	return totalDuration, totalActiveTime, timeByTask, nil
}
