package worktracker

import (
	"encoding/json"
	"os"
	"time"

	er "work-tracker/src/pkg/error"
	"work-tracker/src/pkg/logger"
	"work-tracker/src/pkg/util"
)

type Task struct {
	Name        string    `json:"task_name"`
	Description string    `json:"task_description"`
	CreatedAt   time.Time `json:"created_at"`
}

func loadTasks(path string) (tasks []Task, e *er.Error) {
	logger.Log(logger.Info, logger.BlueColor, "%s tasks list from '%s'", "Loading", path)

	if !util.FileExists(path) {
		return tasks, nil
	}
	raw, err := os.ReadFile(path)
	if err != nil {
		return nil, er.NewError(err, "failed to read tasks.json", path)
	}
	if err := json.Unmarshal(raw, &tasks); err != nil {
		return nil, er.NewError(err, "failed to parse tasks.json", string(raw))
	}

	logger.Log(logger.Info1, logger.GreenColor, "%s %s tasks list from '%s'", "Loaded", len(tasks), path)
	return tasks, nil
}
