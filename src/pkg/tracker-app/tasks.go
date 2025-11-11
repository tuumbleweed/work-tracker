package trackerapp

import (
	"encoding/json"
	"os"
	"time"

	"work-tracker/src/pkg/util"

	tl "github.com/tuumbleweed/tintlog/logger"
	"github.com/tuumbleweed/tintlog/palette"
	"github.com/tuumbleweed/xerr"
)

type Task struct {
	Name        string    `json:"task_name"`
	Description string    `json:"task_description"`
	CreatedAt   time.Time `json:"created_at"`
}

func loadTasks(path string) (tasks []Task, e *xerr.Error) {
	tl.Log(tl.Info, palette.Blue, "%s tasks list from '%s'", "Loading", path)

	if !util.FileExists(path) {
		return tasks, nil
	}
	raw, err := os.ReadFile(path)
	if err != nil {
		return nil, xerr.NewError(err, "failed to read tasks.json", path)
	}
	if err := json.Unmarshal(raw, &tasks); err != nil {
		return nil, xerr.NewError(err, "failed to parse tasks.json", string(raw))
	}

	tl.Log(tl.Info1, palette.Green, "%s %s tasks list from '%s'", "Loaded", len(tasks), path)
	return tasks, nil
}
