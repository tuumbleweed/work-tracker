// you can add any code you want here but don't commit it.
// keep it empty for future projects and for use ase a template.
package main

import (
	"flag"
	"time"

	tl "github.com/tuumbleweed/tintlog/logger"
	"github.com/tuumbleweed/tintlog/palette"

	"work-tracker/src/pkg/config"
	"work-tracker/src/pkg/util"
	worktracker "work-tracker/src/pkg/work-tracker"
)

func main() {
	util.CheckIfEnvVarsPresent([]string{})
	// common flags
	configPath := flag.String("config", "./cfg/config.json", "Path to your configuration file.")
	// program's custom flags
	uiInterval := flag.Duration("tick-interval", 1*time.Second, "UI and activity update period (e.g. 2m, 10m, 1h)")
	chunkInterval := flag.Duration("flush-interval", 10*time.Second, "Autosave period (e.g. 2m, 10m, 1h)")
	workDir := flag.String("work-dir", "./out", "Directory for daily JSONL files")
	tasksFilePath := flag.String("tasks", "./cfg/tasks.json", "File with tasks and their descriptions")
	// parse and init config
	flag.Parse()
	config.InitializeConfig(*configPath)

	tl.Log(
		tl.Notice, palette.BlueBold, "%s worktracker --ui-interval %s, --save-interval %s, --work-dir %s. Config path: '%s'",
		"Running", *uiInterval, *chunkInterval, workDir, *configPath,
	)

	util.CreateDirIfDoesntExist(*workDir).QuitIf("error")

	trackerApp, e := worktracker.InitializeTrackerApp("Worktracker", "Work Tracker", *workDir, *tasksFilePath, *uiInterval, *chunkInterval)
	e.QuitIf("error")
	trackerApp.Start()
}
