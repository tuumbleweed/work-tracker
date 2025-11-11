package trackerapp

import (
	"encoding/json"
	"errors"
	"os"
	"time"

	tl "github.com/tuumbleweed/tintlog/logger"
	"github.com/tuumbleweed/tintlog/palette"
	"github.com/tuumbleweed/xerr"
)

func flushChunk(
	filePath string, start, end time.Time, ActiveDuringThisChunk time.Duration,
	currentTaskName string,
) (e *xerr.Error) {

	tl.Log(tl.Detailed, palette.Blue, "%s chunk to file: '%s'", "Flushing", filePath)

	if !end.After(start) {
		e = xerr.NewError(errors.New("bad chunk"), "!end.After(start)", map[string]any{
			"start": start,
			"end":   end,
		})
		tl.Log(tl.Error, palette.Red, "Failed to flush chunk: %v", e)
		return e
	}

	// remove monotonic component
	start = start.Round(0)
	end = end.Round(0)

	duration := end.Sub(start)
	// clamp it between 0 and 100%
	ActiveDuringThisChunk = Clamp(ActiveDuringThisChunk, 0, duration)

	chunk := Chunk{
		TaskName:   currentTaskName,
		StartedAt:  start,
		FinishedAt: end,
		ActiveTime: ActiveDuringThisChunk,
	}

	e = appendChunk(filePath, chunk)
	if e != nil {
		tl.Log(tl.Error, palette.Red, "Failed to append chunk: %v", e)
		return e
	}

	tl.Log(tl.Detailed1, palette.Green, "%s chunk to file: '%s'", "Flushed", filePath)
	return nil
}

func appendChunk(filePath string, chunk Chunk) (e *xerr.Error) {
	// open a file
	f, err := os.OpenFile(filePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o644)
	if err != nil {
		e = xerr.NewError(err, "failed to open file for appending", map[string]any{
			"file_path": filePath,
		})
		return e
	}
	defer f.Close()

	// marshal the chunk
	b, err := json.Marshal(chunk)
	if err != nil {
		e = xerr.NewError(err, "failed to marshal chunk", map[string]any{
			"file_path": filePath,
			"chunk":     chunk,
		})
		return e
	}

	// write it
	_, err = f.Write(append(b, '\n'))
	if err != nil {
		e = xerr.NewError(err, "failed to write chunk to file", map[string]any{
			"file_path": filePath,
			"chunk":     chunk,
		})
		return e
	}

	return nil
}
