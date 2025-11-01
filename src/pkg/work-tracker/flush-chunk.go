package worktracker

import (
	"encoding/json"
	"errors"
	"os"
	"time"

	er "work-tracker/src/pkg/error"
	"work-tracker/src/pkg/logger"
)

func flushChunk(
	filePath string, start, end time.Time, ActiveDuringThisChunk time.Duration,
	currentTaskName string,
) (e *er.Error) {

	logger.Log(logger.Detailed, logger.BlueColor, "%s chunk to file: '%s'", "Flushing", filePath)

	if !end.After(start) {
		e = er.NewError(errors.New("bad chunk"), "!end.After(start)", map[string]any{
			"start": start,
			"end":   end,
		})
		logger.Log(logger.Error, logger.RedColor, "Failed to flush chunk: %v", e)
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
		logger.Log(logger.Error, logger.RedColor, "Failed to append chunk: %v", e)
		return e
	}

	logger.Log(logger.Detailed1, logger.GreenColor, "%s chunk to file: '%s'", "Flushed", filePath)
	return nil
}

func appendChunk(filePath string, chunk Chunk) (e *er.Error) {
	// open a file
	f, err := os.OpenFile(filePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o644)
	if err != nil {
		e = er.NewError(err, "failed to open file for appending", map[string]any{
			"file_path": filePath,
		})
		return e
	}
	defer f.Close()

	// marshal the chunk
	b, err := json.Marshal(chunk)
	if err != nil {
		e = er.NewError(err, "failed to marshal chunk", map[string]any{
			"file_path": filePath,
			"chunk":     chunk,
		})
		return e
	}

	// write it
	_, err = f.Write(append(b, '\n'))
	if err != nil {
		e = er.NewError(err, "failed to write chunk to file", map[string]any{
			"file_path": filePath,
			"chunk":     chunk,
		})
		return e
	}

	return nil
}
