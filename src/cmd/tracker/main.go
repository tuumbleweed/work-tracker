// you can add any code you want here but don't commit it.
// keep it empty for future projects and for use ase a template.
package main

import (
	"flag"
	"time"

	"my-project/src/pkg/config"
	"my-project/src/pkg/logger"
	"my-project/src/pkg/util"
	"my-project/src/pkg/worktracker"
)

func main() {
	util.CheckIfEnvVarsPresent([]string{})
	// common flags
	logLevelOverride := flag.Int("log-level", -1, "Log level. Default is whatever value is in configuration file. Keep at -1 to not override.")
	logDirOverride := flag.String("log-dir", "", "File directory at which to save log files. Keep empty to use configuration file instead.")
	configPath := flag.String("config", "./cfg/config.json", "Path to your configuration file.")
	// program's custom flags
	uiInterval := flag.Duration("ui-update-interval", 1*time.Second, "UI update period (e.g. 2m, 10m, 1h)")
	chunkInterval := flag.Duration("chunk-save-interval", 1*time.Second, "Autosave period (e.g. 2m, 10m, 1h)")
	workDir := flag.String("work-dir", "./out", "Directory for daily JSONL files")
	// parse and init config
	flag.Parse()
	config.InitializeConfig(*configPath, logger.LogLevel(*logLevelOverride), *logDirOverride)

	logger.Log(
		logger.Notice, logger.BoldBlueColor, "%s worktracker. --ui-interval %s, --save-interval %s, --work-dir %s. Config path: '%s'",
		"Running", *uiInterval, *chunkInterval, workDir, *configPath,
	)

	util.CreateDirIfDoesntExist(*workDir).QuitIf("error")

	_, e := worktracker.InitializeTrackerApp("Worktracker", "Work Tracker", *workDir, *uiInterval, *chunkInterval)
	e.QuitIf("error")
}
